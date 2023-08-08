package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
)

//go:generate sh -c "mkdir -p ../api/frpc/mount/v1 && protoc --go-frpc_out=../api/frpc/mount/v1 --go-frpc_opt=paths=source_relative --proto_path=../../api/proto/mount/v1 ../../api/proto/mount/v1/*.proto"

type BackendServiceFrpc struct {
	svc *BackendService
}

func NewBackendServiceFrpc(svc *BackendService) *BackendServiceFrpc {
	return &BackendServiceFrpc{svc}
}

func (s *BackendServiceFrpc) ReadAt(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMountV1ReadAtArgs) (*v1.ComPojtingerFelicitasR3MapMountV1ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.Length), args.Off)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMountV1ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *BackendServiceFrpc) WriteAt(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMountV1WriteAtArgs) (*v1.ComPojtingerFelicitasR3MapMountV1WriteAtReply, error) {
	length, err := s.svc.WriteAt(ctx, args.P, args.Off)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMountV1WriteAtReply{
		Length: int32(length),
	}, nil
}

func (s *BackendServiceFrpc) Size(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMountV1SizeArgs) (*v1.ComPojtingerFelicitasR3MapMountV1SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMountV1SizeReply{
		Size: size,
	}, nil
}

func (s *BackendServiceFrpc) Sync(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMountV1SyncArgs) (*v1.ComPojtingerFelicitasR3MapMountV1SyncReply, error) {
	if err := s.svc.Sync(ctx); err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMountV1SyncReply{}, nil
}
