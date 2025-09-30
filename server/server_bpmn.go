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

package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/olive-io/bpmn/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/server/dao"
	"github.com/olive-io/gflow/server/scheduler"
)

var _ pb.BpmnRPCServer = (*bpmnGRPCServer)(nil)

type bpmnGRPCServer struct {
	pb.UnimplementedBpmnRPCServer

	ctx context.Context
	lg  *zap.Logger

	sch *scheduler.Scheduler

	definitionsDao *dao.DefinitionsDao
	processDao     *dao.ProcessDao
}

func newBpmnServer(ctx context.Context, lg *zap.Logger, sch *scheduler.Scheduler, definitionsDao *dao.DefinitionsDao, processDao *dao.ProcessDao) *bpmnGRPCServer {
	wch := sch.Watch(ctx, "bpmn-grpc-rpc")
	server := &bpmnGRPCServer{
		ctx:            ctx,
		lg:             lg,
		sch:            sch,
		definitionsDao: definitionsDao,
		processDao:     processDao,
	}
	go server.process(wch)
	return server
}

func (bgs *bpmnGRPCServer) DeployDefinition(ctx context.Context, req *pb.DeployDefinitionsRequest) (*pb.DeployDefinitionsResponse, error) {
	bpmnDef, err := schema.Parse(req.Content)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	idPtr, _ := bpmnDef.Id()
	if idPtr == nil || *idPtr == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Definitions id is required"))
	}
	uid := *idPtr

	executed := false
	for i := range *bpmnDef.Processes() {
		process := (*bpmnDef.Processes())[i]
		executePtr, _ := process.IsExecutable()
		executed = executePtr
		if executed {
			break
		}
	}
	if !executed {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("no executable process found"))
	}

	definitions, err := bgs.definitionsDao.GetDefinitions(ctx, 0, uid)
	if err != nil {
		if !dao.IsNotFound(err) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		definitions = &types.Definitions{
			Uid:         uid,
			Description: req.Description,
			Metadata:    req.Metadata,
			Content:     string(req.Content),
			Version:     1,
			IsExecute:   true,
		}
		id, err := bgs.definitionsDao.CreateDefinitions(ctx, definitions)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if definitions.Id == 0 {
			definitions.Id = id
		}

		bgs.lg.Info("deploy definition",
			zap.String("uid", uid),
			zap.Uint64("version", definitions.Version),
		)

		return &pb.DeployDefinitionsResponse{Definitions: definitions}, nil
	}

	if req.Metadata != nil {
		definitions.Metadata = req.Metadata
	}
	if req.Description != "" {
		definitions.Description = req.Description
	}
	if len(req.Content) != 0 {
		definitions.Content = string(req.Content)
		definitions.Version = definitions.Version + 1
	}
	if err = bgs.definitionsDao.UpdateDefinitions(ctx, definitions); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	bgs.lg.Info("deploy definition",
		zap.String("uid", uid),
		zap.Uint64("version", definitions.Version),
	)

	rsp := &pb.DeployDefinitionsResponse{Definitions: definitions}
	return rsp, nil
}

func (bgs *bpmnGRPCServer) ListDefinitions(ctx context.Context, req *pb.ListDefinitionsRequest) (*pb.ListDefinitionsResponse, error) {
	page, size := req.Page, req.Size
	list, total, err := bgs.definitionsDao.ListDefinitions(ctx, page, size)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &pb.ListDefinitionsResponse{
		DefinitionsList: list,
		Total:           total,
	}
	return rsp, nil
}

func (bgs *bpmnGRPCServer) GetDefinitions(ctx context.Context, req *pb.GetDefinitionsRequest) (*pb.GetDefinitionsResponse, error) {
	definitions, err := bgs.definitionsDao.GetDefinitionsWithVersion(ctx, 0, req.Uid, req.Version)
	if err != nil {
		if dao.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &pb.GetDefinitionsResponse{
		Definitions: definitions,
	}
	return rsp, nil
}

func (bgs *bpmnGRPCServer) RemoveDefinitions(ctx context.Context, req *pb.RemoveDefinitionsRequest) (*pb.RemoveDefinitionsResponse, error) {
	return &pb.RemoveDefinitionsResponse{}, nil
}

func (bgs *bpmnGRPCServer) ExecuteProcess(ctx context.Context, req *pb.ExecuteProcessRequest) (*pb.ExecuteProcessResponse, error) {
	definitionUID := req.DefinitionsUid
	version := req.DefinitionsVersion

	definitions, err := bgs.definitionsDao.GetDefinitionsWithVersion(ctx, 0, definitionUID, version)
	if err != nil {
		if dao.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	process := &types.Process{
		Name:     req.Name,
		Uid:      uuid.New().String(),
		Metadata: map[string]string{},
		Priority: req.Priority,
		Args: &types.BpmnArgs{
			Headers:     req.Headers,
			Properties:  req.Properties,
			DataObjects: req.DataObjects,
		},
		DefinitionsUid:     definitionUID,
		DefinitionsVersion: version,
		Context: &types.ProcessContext{
			Variables:   map[string]*types.Value{},
			DataObjects: map[string]*types.Value{},
		},
	}

	if err = bgs.processDao.CreateProcess(ctx, process); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	processStat := &scheduler.ProcessStat{
		Process:     process,
		Definitions: definitions.Content,
		FlowNodes:   make([]*types.FlowNode, 0),
	}

	err = bgs.sch.Execute(processStat)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ExecuteProcessResponse{Process: process}, nil
}

func (bgs *bpmnGRPCServer) ListProcess(ctx context.Context, req *pb.ListProcessRequest) (*pb.ListProcessResponse, error) {
	page, size := req.Page, req.Size
	options := dao.NewListProcessOptions(req.DefinitionsUid, req.DefinitionsVersion)
	processes, total, err := bgs.processDao.ListProcesses(ctx, page, size, options)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &pb.ListProcessResponse{
		Processes: processes,
		Total:     total,
	}
	return rsp, nil
}

func (bgs *bpmnGRPCServer) GetProcess(ctx context.Context, req *pb.GetProcessRequest) (*pb.GetProcessResponse, error) {
	process, err := bgs.processDao.GetProcess(ctx, req.Id)
	if err != nil {
		if dao.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	nodes, err := bgs.processDao.ListFlowNodes(ctx, process.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &pb.GetProcessResponse{
		Process:    process,
		Activities: nodes,
	}
	return rsp, nil
}

func (bgs *bpmnGRPCServer) process(wch *scheduler.WatchChan) {
	lg := bgs.lg
	ctx := bgs.ctx
	for {
		select {
		case <-ctx.Done():
			wch.Close()
			return
		default:
		}

		rsp := wch.Next()
		if rsp.Err != nil {
			continue
		}
		if p := rsp.Process; p != nil {
			err := bgs.processDao.UpdateProcess(ctx, p)
			if err != nil {
				lg.Error("update process",
					zap.Int64("id", p.Id),
					zap.String("definitions", p.DefinitionsUid),
					zap.Uint64("version", p.DefinitionsVersion),
					zap.Error(err))
			}
		}
		if node := rsp.FlowNode; node != nil {
			err := bgs.processDao.SaveFlowNode(ctx, node)
			if err != nil {
				lg.Error("save flow node",
					zap.Int64("process", node.ProcessId),
					zap.String("flow", node.Name),
					zap.Error(err))
			}
		}
	}
}
