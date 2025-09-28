/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/olive-io/bpmn/schema"
	"github.com/olive-io/bpmn/v2"
	"github.com/olive-io/bpmn/v2/pkg/tracing"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
)

type antLogger struct {
	lg *zap.SugaredLogger
}

func (l *antLogger) Printf(format string, args ...any) {
	l.lg.Debugf(format, args...)
}

type ProcessStat struct {
	*types.Process `json:",inline"`

	Definitions string `json:"definitions"`

	FlowNodes []*types.FlowNode `json:"flowNodes"`
}

func (p *ProcessStat) ID() int64 {
	return p.Id
}

type Scheduler struct {
	ctx     context.Context
	cancel  context.CancelFunc
	options *Options

	lg *zap.Logger

	queue *SyncPriorityQueue[*ProcessStat]

	executePool *ants.Pool

	engine *bpmn.Engine

	wmu      sync.RWMutex
	watchers map[string]*WatchChan
}

func NewScheduler(pctx context.Context, options *Options) (*Scheduler, error) {
	lg := options.Logger
	if lg == nil {
		lg = zap.NewNop()
	}

	poolOpts := []ants.Option{
		ants.WithPreAlloc(true),
		ants.WithLogger(&antLogger{lg: lg.Sugar()}),
	}
	executePool, err := ants.NewPool(options.ExecutePoolSize, poolOpts...)
	if err != nil {
		return nil, fmt.Errorf("create execute pool: %w", err)
	}

	ctx, cancel := context.WithCancel(pctx)
	queue := NewSync[*ProcessStat](func(v *ProcessStat) int64 {
		score := v.Priority
		if v.Stage == types.Process_Rollback ||
			v.Stage == types.Process_Destroy {
			score += 300
		}
		if v.Stage == types.Process_Commit {
			score += 200
		}
		return score
	})

	engine := bpmn.NewEngine(bpmn.WithEngineContext(ctx))

	scheduler := &Scheduler{
		ctx:         ctx,
		cancel:      cancel,
		lg:          lg,
		options:     options,
		queue:       queue,
		executePool: executePool,
		engine:      engine,
		watchers:    make(map[string]*WatchChan),
	}

	go scheduler.process(ctx)
	return scheduler, nil
}

func (sch *Scheduler) Execute(stat *ProcessStat) error {
	lg := sch.lg

	lg.Info("push new process",
		zap.Int64("pid", stat.Id),
		zap.Int64("priority", stat.Priority),
		zap.String("status", stat.Status.String()),
		zap.String("stage", stat.Stage.String()))

	sch.queue.Push(stat)
	return nil
}

func (sch *Scheduler) Watch(ctx context.Context, id string) *WatchChan {
	sch.lg.Sugar().Infof("add new watcher [%s]", id)

	wch := newWatchChan(ctx, id, sch.lg, sch.ctx.Done())

	sch.wmu.Lock()
	sch.watchers[id] = wch
	sch.wmu.Unlock()

	return wch
}

func (sch *Scheduler) process(ctx context.Context) {

	interval := time.Microsecond * 100
	timer := time.NewTimer(interval)
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case <-timer.C:
			timer.Reset(interval)

			sch.tick(ctx)
		}
	}

	sch.destroy()
}

func (sch *Scheduler) tick(ctx context.Context) {
	free := sch.executePool.Free()
	if free == 0 {
		return
	}
	value, ok := sch.queue.Pop()
	if !ok {
		return
	}
	stat := value.(*ProcessStat)
	if stat.Status > types.Process_Running {
		return
	}

	var err error
	defer func() {
		if err != nil {
			sch.queue.Push(stat)
		}
	}()

	err = sch.executePool.Submit(func() {
		execErr := sch.execute(ctx, stat)
		if execErr != nil {
			//TODO: handle error
		}
	})
}

func (sch *Scheduler) execute(ctx context.Context, stat *ProcessStat) error {
	lg := sch.lg.Sugar()

	var err error
	defer func() {
		stat.EndAt = time.Now().UnixMilli()
		stat.Stage = types.Process_Finish

		stat.Status = types.Process_Success
		if err != nil {
			stat.ErrMsg = err.Error()
			stat.Status = types.Process_Failed
		}

		sch.setProcess(stat.Process)
	}()

	activeStack := make([]*types.FlowNode, 0)
	nodeMapping := make(map[string]*types.FlowNode)
	for _, node := range stat.FlowNodes {
		nodeMapping[node.Name] = node
	}

	var definitions *schema.Definitions
	definitions, err = schema.Parse([]byte(stat.Definitions))
	if err != nil {
		return fmt.Errorf("parse definitions: %w", err)
	}

	options := make([]bpmn.Option, 0)
	if pctx := stat.Context; pctx != nil {
		dataObjects := map[string]any{}
		for name, dataObject := range pctx.DataObjects {
			dataObjects[name] = dataObject
		}
		options = append(options, bpmn.WithDataObjects(dataObjects))
		variables := map[string]any{}
		for name, variable := range pctx.Variables {
			variables[name] = variable
		}
		options = append(options, bpmn.WithVariables(variables))
	}

	var bp *bpmn.Process
	bp, err = sch.engine.NewProcess(definitions, options...)
	if err != nil {
		return fmt.Errorf("create process: %w", err)
	}

	pid := bp.Id().String()
	traces := bp.Tracer().Subscribe()
	defer bp.Tracer().Unsubscribe(traces)
	if err = bp.StartAll(ctx); err != nil {
		return fmt.Errorf("start bpmn process: %w", err)
	}

	defer func() {
		rctx := ctx
		if rctx.Err() != nil {
			rctx = sch.ctx
		}

		if err != nil {
			stat.Stage = types.Process_Rollback
			sch.setProcess(stat.Process)

			for i := len(activeStack) - 1; i >= 0; i-- {
				node := activeStack[i]
				node.Stage = types.FlowNode_Rollback
				sch.setFlowNode(node)

				lg.Infof("rollback task [%s][%s]", pid, node.FlowId)

				_ = sch.doTask(rctx, node)
			}
		}

		stat.Stage = types.Process_Destroy
		sch.setProcess(stat.Process)

		for i := len(activeStack) - 1; i >= 0; i-- {
			node := activeStack[i]
			node.Stage = types.FlowNode_Destroy
			sch.setFlowNode(node)

			lg.Infof("destroy task [%s][%s]", pid, node.FlowId)

			_ = sch.doTask(rctx, node)

			node.Stage = types.FlowNode_Finish
			sch.setFlowNode(node)
		}
	}()

	if stat.Status != types.Process_Running {
		stat.Uid = pid
		stat.StartAt = time.Now().UnixMilli()

		if id, ok := bp.Element().Id(); ok {
			stat.DefinitionsProcess = *id
		}
		stat.Status = types.Process_Running
		sch.setProcess(stat.Process)
	}

	stat.Stage = types.Process_Commit
	sch.setProcess(stat.Process)

	ech := make(chan error, 1)
	go func(ech chan<- error) {
		for {
			var trace tracing.ITrace
			select {
			case trace = <-traces:
			case <-ctx.Done():
			}

			trace = tracing.Unwrap(trace)
			switch tt := trace.(type) {
			case bpmn.VisitTrace:
				elem := tt.Node
				var fid string
				id, ok := elem.Id()
				if ok {
					fid = *id
				}
				var fname string
				name, ok := elem.Name()
				if ok {
					fname = *name
				}

				node, exists := nodeMapping[fid]
				if !exists {
					node = &types.FlowNode{
						Name:      fname,
						FlowId:    fid,
						FlowType:  parseElementType(elem),
						StartTime: time.Now().UnixMilli(),
						ProcessId: stat.Id,
						Stage:     types.FlowNode_Ready,
					}
					nodeMapping[fid] = node
					stat.FlowNodes = append(stat.FlowNodes, node)
					sch.setFlowNode(node)
				}

			case bpmn.LeaveTrace:
				elem := tt.Node
				var fid string
				id, ok := elem.Id()
				if ok {
					fid = *id
				}
				node, exists := nodeMapping[fid]
				if exists && node.EndTime == 0 {
					node.EndTime = time.Now().UnixMilli()
					node.Stage = types.FlowNode_Finish
					sch.setFlowNode(node)
				}

			case bpmn.CompletionTrace:
				elem := tt.Node
				var fid string
				id, ok := elem.Id()
				if ok {
					fid = *id
				}
				node, exists := nodeMapping[fid]
				if exists && node.EndTime == 0 {
					node.EndTime = time.Now().UnixMilli()
					node.Stage = types.FlowNode_Finish
					sch.setFlowNode(node)
				}

			case bpmn.TaskTrace:
				act := tt.GetActivity()
				elem := act.Element()

				var fid string
				id, ok := elem.Id()
				if ok {
					fid = *id
				}
				var tname string
				name, ok := elem.Name()
				if ok {
					tname = *name
				}
				lg.Infof("commit task [%s][%s]", pid, fid)

				flowNode, exists := nodeMapping[fid]
				if !exists {
					flowNode = &types.FlowNode{
						Name:      tname,
						FlowId:    fid,
						FlowType:  parseTaskType(act.Type()),
						StartTime: time.Now().UnixMilli(),
						ProcessId: stat.Id,
						Stage:     types.FlowNode_Commit,
						Status:    types.FlowNode_Running,
					}
					nodeMapping[fid] = flowNode
					stat.FlowNodes = append(stat.FlowNodes, flowNode)
					sch.setFlowNode(flowNode)
				}

				if flowNode.Stage != types.FlowNode_Commit &&
					flowNode.Stage != types.FlowNode_Rollback &&
					flowNode.Stage != types.FlowNode_Destroy {

					headers := tt.GetHeaders()
					properties := map[string]*types.Value{}
					dataObjects := map[string]*types.Value{}
					for k, v := range tt.GetProperties() {
						tv := schema.NewValue(v.Value())
						properties[k] = fromSchemaValue(tv)
					}
					for k, v := range tt.GetDataObjects() {
						tv := schema.NewValue(v.Value())
						dataObjects[k] = fromSchemaValue(tv)
					}

					flowNode.Headers = headers
					flowNode.Properties = properties
					flowNode.DataObjects = dataObjects
					flowNode.Stage = types.FlowNode_Commit
					sch.setFlowNode(flowNode)

					doOptions := make([]bpmn.DoOption, 0)
					flowNode.Status = types.FlowNode_Success
					doErr := sch.doTask(ctx, flowNode)
					if doErr != nil {
						flowNode.Status = types.FlowNode_Failed
						flowNode.ErrMsg = doErr.Error()
						doOptions = append(doOptions, bpmn.DoWithErr(err))
					}

					tt.Do(doOptions...)

					sch.setFlowNode(flowNode)

					activeStack = append(activeStack, flowNode)
				} else {
					tt.Do()
				}

			case bpmn.ActiveBoundaryTrace:

			case bpmn.ErrorTrace:
				ech <- tt.Error
				return
			case bpmn.CeaseFlowTrace:
				return
			default:
			}
		}
	}(ech)

	bp.WaitUntilComplete(ctx)
	select {
	case <-ctx.Done():
		err = ctx.Err()
		ctx = context.Background()
	case err = <-ech:
	default:
	}

	return err
}

func (sch *Scheduler) destroy() {
	sch.lg.Debug("destroying scheduler")

	sch.cancel()

	// release goroutines pool
	sch.lg.Debug("release scheduler execute pool")
	sch.executePool.Release()
	sch.lg.Debug("released scheduler execute pool")
}

func (sch *Scheduler) doTask(ctx context.Context, node *types.FlowNode) error {
	return nil
}

func (sch *Scheduler) setProcess(process *types.Process) {
	sch.sendWatchMsg(&WatchMessage{Process: process.DeepCopy()})
}

func (sch *Scheduler) setFlowNode(flowNode *types.FlowNode) {
	sch.sendWatchMsg(&WatchMessage{FlowNode: flowNode.DeepCopy()})
}

func (sch *Scheduler) sendWatchMsg(msg *WatchMessage) {
	deleted := make([]string, 0)
	sch.wmu.RLock()
	for name, w := range sch.watchers {
		if w.IsClosed() {
			deleted = append(deleted, name)
			continue
		}
		w.send(msg)
	}
	sch.wmu.RUnlock()

	for _, name := range deleted {
		sch.deleteWatcher(name)
	}
}

func (sch *Scheduler) deleteWatcher(name string) {
	sch.wmu.Lock()
	delete(sch.watchers, name)
	sch.wmu.Unlock()
}
