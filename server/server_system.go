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
	"errors"
	"io"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/server/config"
	"github.com/olive-io/gflow/server/dao"
	"github.com/olive-io/gflow/server/dispatch"
)

var _ pb.SystemRPCServer = (*systemGRPCServer)(nil)

type systemGRPCServer struct {
	pb.UnimplementedSystemRPCServer

	ctx context.Context
	lg  *zap.Logger
	cfg *config.Config

	runnerDao   *dao.RunnerDao
	endpointDao *dao.EndpointDao

	dispatcher *dispatch.Dispatcher
}

func newSystemGRPCServer(
	ctx context.Context, lg *zap.Logger, cfg *config.Config,
	runnerDao *dao.RunnerDao, endpointDao *dao.EndpointDao,
	dispatcher *dispatch.Dispatcher) *systemGRPCServer {

	server := &systemGRPCServer{
		ctx: ctx,
		lg:  lg,
		cfg: cfg,

		runnerDao:   runnerDao,
		endpointDao: endpointDao,
		dispatcher:  dispatcher,
	}

	return server
}

func (sgs *systemGRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Reply: "OK"}, nil
}

func (sgs *systemGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	runner := req.GetRunner()
	runner.Online = 1

	value, _ := sgs.runnerDao.Get(ctx, runner.Id, runner.Uid)
	if value != nil {
		runner.Id = value.Id
		runner.OnlineTimestamp = time.Now().UnixMilli()

		sgs.lg.Info("runner online",
			zap.String("uid", runner.Uid),
			zap.String("transport", runner.Transport.String()),
			zap.String("listen", runner.ListenUrl))

		err := sgs.runnerDao.Update(ctx, runner.Id, runner)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		_, err := sgs.runnerDao.Create(ctx, runner)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		sgs.lg.Info("register new runner",
			zap.String("uid", runner.Uid),
			zap.String("transport", runner.Transport.String()),
			zap.String("listen", runner.ListenUrl))
	}

	rsp := &pb.RegisterResponse{
		Runner: runner,
	}
	return rsp, nil
}

func (sgs *systemGRPCServer) Disregister(ctx context.Context, req *pb.DisregisterRequest) (*pb.DisregisterResponse, error) {
	runner, err := sgs.runnerDao.Get(ctx, 0, req.Id)
	if err != nil {
		if dao.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	runner.OfflineTimestamp = time.Now().UnixMilli()
	runner.Online = 0

	err = sgs.runnerDao.Update(ctx, runner.Id, runner)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rsp := &pb.DisregisterResponse{
		Runner: runner,
	}
	return rsp, nil
}

func (sgs *systemGRPCServer) ListRunners(ctx context.Context, req *pb.ListRunnersRequest) (*pb.ListRunnersResponse, error) {
	page, size := int(req.Page), int(req.Size)
	list, total, err := sgs.runnerDao.PageList(ctx, page, size, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rsp := &pb.ListRunnersResponse{
		Runners: list,
		Total:   total,
	}
	return rsp, nil
}

func (sgs *systemGRPCServer) GetRunner(ctx context.Context, req *pb.GetRunnerRequest) (*pb.GetRunnerResponse, error) {
	runner, err := sgs.runnerDao.Get(ctx, req.Id, req.Uid)
	if err != nil {
		if dao.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	rsp := &pb.GetRunnerResponse{Runner: runner}
	return rsp, nil
}

func (sgs *systemGRPCServer) ListEndpoints(ctx context.Context, req *pb.ListEndpointsRequest) (*pb.ListEndpointsResponse, error) {
	page, size := int(req.Page), int(req.Size)
	list, total, err := sgs.endpointDao.PageList(ctx, page, size, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rsp := &pb.ListEndpointsResponse{
		Endpoints: list,
		Total:     total,
	}
	return rsp, nil
}

func (sgs *systemGRPCServer) AddEndpoints(ctx context.Context, req *pb.AddEndpointsRequest) (*pb.AddEndpointsResponse, error) {
	for _, endpoint := range req.Endpoints {
		if err := sgs.addEndpoint(ctx, endpoint, req.Target); err != nil {
			return nil, err
		}
	}

	resp := &pb.AddEndpointsResponse{}
	return resp, nil
}

func (sgs *systemGRPCServer) addEndpoint(ctx context.Context, endpoint *types.Endpoint, target string) error {
	cond := map[string]string{"name": endpoint.Name}
	value, err := sgs.endpointDao.First(ctx, cond)
	if err != nil {
		if !dao.IsNotFound(err) {
			return status.Error(codes.Internal, err.Error())
		}

		endpoint.Targets = []string{target}
		_, err = sgs.endpointDao.Create(ctx, endpoint)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		found := false
		for _, item := range value.Targets {
			if item == target {
				found = true
				break
			}
		}
		if !found {
			value.Targets = append(value.Targets, target)
			err = sgs.endpointDao.Update(ctx, value.Id, value)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	return nil
}

func (sgs *systemGRPCServer) RunnerDispatch(stream pb.SystemRPC_RunnerDispatchServer) error {

	var listenURL string
	lg := sgs.lg
	ctx := stream.Context()
	peerInfo, ok := peer.FromContext(ctx)
	if ok {
		listenURL = "grpcstream://" + peerInfo.Addr.String()
	}

	rsp, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	handshake := rsp.Handshake
	if handshake == nil || handshake.Runner == nil {
		return status.Errorf(codes.InvalidArgument, "missing handshake message")
	}

	runnerMsg := handshake.Runner
	runner, err := sgs.runnerDao.Get(ctx, 0, runnerMsg.Uid)
	if err != nil {
		if !dao.IsNotFound(err) {
			return status.Error(codes.Internal, err.Error())
		}
		runner = runnerMsg
		runner.ListenUrl = listenURL
		_, _ = sgs.runnerDao.Create(ctx, runner)
	} else {
		runner.ListenUrl = listenURL
		_ = sgs.runnerDao.Update(ctx, runner.Id, runner)
	}

	serverID := sgs.cfg.Server.ID
	reply := &pb.RunnerDispatchResponse{
		Handshake: &pb.HandshakeResponse{
			ServerID: serverID,
			Runner:   runner,
		},
	}
	if err = stream.Send(reply); err != nil {
		lg.Error("sends handshake response", zap.Error(err))
		return err
	}

	lg.Info("add runner dispatcher",
		zap.String("uid", runner.Uid),
		zap.String("listen-url", runner.ListenUrl),
		zap.String("hostname", runner.Hostname),
	)

	input, err := sgs.dispatcher.Register(ctx, runner.Uid)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case req, allowed := <-input:
				if !allowed {
					return
				}
				if in := req.CallTask; in != nil {
					msg := &pb.RunnerDispatchResponse{CallTask: in}
					if serr := stream.Send(msg); serr != nil {
						result := &types.CallTaskResponse{SeqId: in.SeqId, Error: serr.Error()}
						sgs.dispatcher.Reply(&dispatch.Response{CallTask: result})
						lg.Error("sends call task response", zap.Error(serr))
					}
				}
			}
		}
	}()

LOOP:
	for {
		recv, rerr := stream.Recv()
		if rerr != nil {
			if rerr == io.EOF || errors.Is(rerr, context.Canceled) {
				lg.Error("gflow runner dispatcher disconnected", zap.Error(rerr))
			} else {
				lg.Error("receives runner message", zap.Error(rerr))
			}
			break LOOP
		}

		switch {
		case recv.Heartbeat != nil:
			msg := &pb.RunnerDispatchResponse{HeartBeat: &pb.HeartBeatResponse{}}
			if serr := stream.Send(msg); serr != nil {
				lg.Error("sends heartbeat response", zap.Error(serr))
			}
		case recv.CallTask != nil:
			sgs.dispatcher.Reply(&dispatch.Response{CallTask: recv.CallTask})
		}
	}

	lg.Info("remove runner dispatcher",
		zap.String("uid", runner.Uid),
		zap.String("listen-url", runner.ListenUrl),
		zap.String("hostname", runner.Hostname),
	)

	return nil
}
