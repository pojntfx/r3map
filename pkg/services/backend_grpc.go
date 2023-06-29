package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
)

//go:generate sh -c "mkdir -p ../api/proto/mount/v1 && protoc --go_out=../api/proto/mount/v1 --go_opt=paths=source_relative --go-grpc_out=../api/proto/mount/v1 --go-grpc_opt=paths=source_relative --proto_path=../../api/proto/mount/v1 ../../api/proto/mount/v1/*.proto"

type BackendGrpc struct {
	v1.UnimplementedBackendServer

	svc *Backend
}

func NewBackendGrpc(svc *Backend) *BackendGrpc {
	return &BackendGrpc{v1.UnimplementedBackendServer{}, svc}
}

func (s *BackendGrpc) ReadAt(ctx context.Context, args *v1.ReadAtArgs) (*v1.ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.GetLength()), args.GetOff())
	if err != nil {
		return nil, err
	}

	return &v1.ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *BackendGrpc) WriteAt(ctx context.Context, args *v1.WriteAtArgs) (*v1.WriteAtReply, error) {
	length, err := s.svc.WriteAt(ctx, args.GetP(), args.GetOff())
	if err != nil {
		return nil, err
	}

	return &v1.WriteAtReply{
		Length: int32(length),
	}, nil
}

func (s *BackendGrpc) Size(ctx context.Context, args *v1.SizeArgs) (*v1.SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SizeReply{
		Size: size,
	}, nil
}

func (s *BackendGrpc) Sync(ctx context.Context, args *v1.SyncArgs) (*v1.SyncReply, error) {
	if err := s.svc.Sync(ctx); err != nil {
		return nil, err
	}

	return &v1.SyncReply{}, nil
}
