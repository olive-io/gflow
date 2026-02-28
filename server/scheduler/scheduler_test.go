/*
Copyright 2025 The gflow Authors

This program is offered under a commercial and under the AGPL license.
For AGPL licensing, see below.

AGPL licensing:
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package scheduler

import (
	"context"
	"embed"
	"encoding/xml"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	traceutil "github.com/olive-io/gflow/pkg/trace"
)

//go:embed testdata
var embedFS embed.FS

//go:embed testdata
var testdata embed.FS

func LoadTestFile(filename string, definitions any) {
	var err error
	src, err := testdata.ReadFile(filename)
	if err != nil {
		log.Fatalf("Can't read file %s: %v", filename, err)
	}
	err = xml.Unmarshal(src, definitions)
	if err != nil {
		log.Fatalf("XML unmarshalling error in %s: %v", filename, err)
	}
}

func TestScheduler(t *testing.T) {
	lg := otelzap.New(zap.NewExample())
	provider := traceutil.DefaultProvider()
	options := NewOptions(lg, provider.Tracer("test"))
	ctx := context.Background()
	sche, err := NewScheduler(ctx, options)
	if !assert.NoError(t, err) {
		return
	}
	wch := sche.Watch(ctx, "1")

	content, err := testdata.ReadFile("testdata/task.bpmn")
	if !assert.NoError(t, err) {
		return
	}

	process := &types.Process{
		Name:     "test",
		Uid:      "pa",
		Metadata: map[string]string{},
		Priority: 1,
		Args: &types.BpmnArgs{
			Headers:     map[string]string{},
			Properties:  map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		DefinitionsUid:     "Definitions_04fu1l0",
		DefinitionsVersion: 1,
		DefinitionsProcess: "test",
		Context: &types.ProcessContext{
			Variables:   map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		Attempts: 0,
		Status:   types.Process_Waiting,
	}

	stat := &ProcessStat{
		Process:     process,
		Definitions: string(content),
		FlowNodes:   []*types.FlowNode{},
	}

	err = sche.Execute(stat)
	if !assert.NoError(t, err) {
		return
	}

	for {
		rsp := wch.Recv()
		if rsp.Err != nil {
			t.Fatalf("rsp err: %v", rsp.Err)
		}

		if rsp.Process != nil {
			t.Logf("process: %#v", rsp.Process)
			if rsp.Process.Stage == types.Process_Finish {
				break
			}
		}
		if rsp.FlowNode != nil {
			t.Logf("flow node: %#v", rsp.FlowNode)
		}
	}
}

// TestPriorityQueue 测试优先级队列
func TestPriorityQueue(t *testing.T) {
	queue := NewSync[*ProcessStat](func(v *ProcessStat) int64 {
		return v.Priority
	})

	// 测试空队列
	assert.Equal(t, 0, queue.Len())
	v, ok := queue.Pop()
	assert.False(t, ok)
	assert.Nil(t, v)

	// 测试推入元素
	stat1 := &ProcessStat{
		Process: &types.Process{
			Id:       1,
			Priority: 1,
		},
	}
	stat2 := &ProcessStat{
		Process: &types.Process{
			Id:       2,
			Priority: 3,
		},
	}
	stat3 := &ProcessStat{
		Process: &types.Process{
			Id:       3,
			Priority: 2,
		},
	}

	queue.Push(stat1)
	queue.Push(stat2)
	queue.Push(stat3)

	assert.Equal(t, 3, queue.Len())

	// 测试弹出元素（按优先级排序）
	v, ok = queue.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(2), v.(*ProcessStat).Id)
	assert.Equal(t, int64(3), v.(*ProcessStat).Priority)

	v, ok = queue.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(3), v.(*ProcessStat).Id)
	assert.Equal(t, int64(2), v.(*ProcessStat).Priority)

	v, ok = queue.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(1), v.(*ProcessStat).Id)
	assert.Equal(t, int64(1), v.(*ProcessStat).Priority)

	// 测试空队列
	assert.Equal(t, 0, queue.Len())
	v, ok = queue.Pop()
	assert.False(t, ok)
	assert.Nil(t, v)
}

// TestSchedulerWithError 测试调度器错误处理
func TestSchedulerWithError(t *testing.T) {
	lg := otelzap.New(zap.NewExample())
	provider := traceutil.DefaultProvider()
	options := NewOptions(lg, provider.Tracer("test"))
	ctx := context.Background()
	sche, err := NewScheduler(ctx, options)
	assert.NoError(t, err)

	// 创建一个无效的流程定义（空字符串）
	process := &types.Process{
		Name:     "test-error",
		Uid:      "error-pa",
		Priority: 1,
		Status:   types.Process_Waiting,
	}

	stat := &ProcessStat{
		Process:     process,
		Definitions: "", // 空的流程定义
		FlowNodes:   []*types.FlowNode{},
	}

	// 执行流程
	err = sche.Execute(stat)
	assert.NoError(t, err)

	// 等待一段时间，让调度器处理错误
	time.Sleep(500 * time.Millisecond)

	// 验证流程阶段是否为完成
	assert.Equal(t, types.Process_Finish, process.Stage)
	// 验证是否有错误消息
	assert.NotEmpty(t, process.ErrMsg)
}

// TestSchedulerWithMultipleProcesses 测试调度器处理多个流程
func TestSchedulerWithMultipleProcesses(t *testing.T) {
	lg := otelzap.New(zap.NewExample())
	provider := traceutil.DefaultProvider()
	options := NewOptions(lg, provider.Tracer("test"))
	ctx := context.Background()
	sche, err := NewScheduler(ctx, options)
	assert.NoError(t, err)

	// 读取测试 BPMN 文件
	content, err := testdata.ReadFile("testdata/task.bpmn")
	assert.NoError(t, err)

	// 创建多个流程
	for i := 0; i < 3; i++ {
		process := &types.Process{
			Name:     "test-multi",
			Uid:      fmt.Sprintf("pa-%d", i),
			Priority: int64(i + 1),
			Args: &types.BpmnArgs{
				Headers:     map[string]string{},
				Properties:  map[string]*types.Value{},
				DataObjects: map[string]*types.Value{},
			},
			DefinitionsUid:     "Definitions_04fu1l0",
			DefinitionsVersion: 1,
			DefinitionsProcess: "test",
			Context: &types.ProcessContext{
				Variables:   map[string]*types.Value{},
				DataObjects: map[string]*types.Value{},
			},
			Attempts: 0,
			Status:   types.Process_Waiting,
		}

		stat := &ProcessStat{
			Process:     process,
			Definitions: string(content),
			FlowNodes:   []*types.FlowNode{},
		}

		err = sche.Execute(stat)
		assert.NoError(t, err)
	}

	// 等待流程完成
	time.Sleep(500 * time.Millisecond)
}

// TestSchedulerWithDifferentModes 测试调度器处理不同模式
func TestSchedulerWithDifferentModes(t *testing.T) {
	lg := otelzap.New(zap.NewExample())
	provider := traceutil.DefaultProvider()
	options := NewOptions(lg, provider.Tracer("test"))
	ctx := context.Background()
	sche, err := NewScheduler(ctx, options)
	assert.NoError(t, err)

	// 读取测试 BPMN 文件
	content, err := testdata.ReadFile("testdata/task.bpmn")
	assert.NoError(t, err)

	// 测试 Simple 模式
	process := &types.Process{
		Name:     "test-simple-mode",
		Uid:      "pa-simple",
		Priority: 1,
		Mode:     types.TransitionMode_Simple,
		Args: &types.BpmnArgs{
			Headers:     map[string]string{},
			Properties:  map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		DefinitionsUid:     "Definitions_04fu1l0",
		DefinitionsVersion: 1,
		DefinitionsProcess: "test",
		Context: &types.ProcessContext{
			Variables:   map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		Attempts: 0,
		Status:   types.Process_Waiting,
	}

	stat := &ProcessStat{
		Process:     process,
		Definitions: string(content),
		FlowNodes:   []*types.FlowNode{},
	}

	err = sche.Execute(stat)
	assert.NoError(t, err)

	// 测试 Transition 模式
	process2 := &types.Process{
		Name:     "test-transition-mode",
		Uid:      "pa-transition",
		Priority: 1,
		Mode:     types.TransitionMode_Transition,
		Args: &types.BpmnArgs{
			Headers:     map[string]string{},
			Properties:  map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		DefinitionsUid:     "Definitions_04fu1l0",
		DefinitionsVersion: 1,
		DefinitionsProcess: "test",
		Context: &types.ProcessContext{
			Variables:   map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		Attempts: 0,
		Status:   types.Process_Waiting,
	}

	stat2 := &ProcessStat{
		Process:     process2,
		Definitions: string(content),
		FlowNodes:   []*types.FlowNode{},
	}

	err = sche.Execute(stat2)
	assert.NoError(t, err)

	// 等待流程完成
	time.Sleep(500 * time.Millisecond)
}

// TestWatcherMechanism 测试观察者机制
func TestWatcherMechanism(t *testing.T) {
	lg := otelzap.New(zap.NewExample())
	provider := traceutil.DefaultProvider()
	options := NewOptions(lg, provider.Tracer("test"))
	ctx := context.Background()
	sche, err := NewScheduler(ctx, options)
	assert.NoError(t, err)

	// 添加观察者
	wch := sche.Watch(ctx, "test-watcher")

	// 读取测试 BPMN 文件
	content, err := testdata.ReadFile("testdata/task.bpmn")
	assert.NoError(t, err)

	// 创建流程
	process := &types.Process{
		Name:     "test-watcher",
		Uid:      "pa-watcher",
		Priority: 1,
		Args: &types.BpmnArgs{
			Headers:     map[string]string{},
			Properties:  map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		DefinitionsUid:     "Definitions_04fu1l0",
		DefinitionsVersion: 1,
		DefinitionsProcess: "test",
		Context: &types.ProcessContext{
			Variables:   map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
		Attempts: 0,
		Status:   types.Process_Waiting,
	}

	stat := &ProcessStat{
		Process:     process,
		Definitions: string(content),
		FlowNodes:   []*types.FlowNode{},
	}

	// 执行流程
	err = sche.Execute(stat)
	assert.NoError(t, err)

	// 接收观察者消息
	processReceived := false

	for i := 0; i < 10; i++ {
		rsp := wch.Recv()
		if rsp.Err != nil {
			break
		}

		if rsp.Process != nil {
			processReceived = true
			if rsp.Process.Stage == types.Process_Finish {
				break
			}
		}
	}

	// 验证是否收到消息
	assert.True(t, processReceived, "Should receive process messages")
}
