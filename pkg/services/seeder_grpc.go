package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/v1"
)

//go:generate sh -c "mkdir -p ../api/proto/v1 && protoc --go_out=../api/proto/v1 --go_opt=paths=source_relative --go-grpc_out=../api/proto/v1 --go-grpc_opt=paths=source_relative --go-drpc_out=../api/proto/v1 --go-drpc_opt=paths=source_relative --proto_path=../../api/proto/v1 ../../api/proto/v1/*.proto"

type SeederGrpc struct {
	v1.UnimplementedSeederServer

	svc *Seeder
}

func NewSeederGrpc(svc *Seeder) *SeederGrpc {
	return &SeederGrpc{v1.UnimplementedSeederServer{}, svc}
}

func (s *SeederGrpc) ReadAt(ctx context.Context, args *v1.ReadAtArgs) (*v1.ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.GetLength()), args.GetOff())
	if err != nil {
		return nil, err
	}

	return &v1.ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *SeederGrpc) Size(ctx context.Context, args *v1.SizeArgs) (*v1.SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SizeReply{
		N: size,
	}, nil
}

func (s *SeederGrpc) Track(ctx context.Context, args *v1.TrackArgs) (*v1.TrackReply, error) {
	if err := s.svc.Track(ctx); err != nil {
		return nil, err
	}

	return &v1.TrackReply{}, nil
}

func (s *SeederGrpc) Sync(ctx context.Context, args *v1.SyncArgs) (*v1.SyncReply, error) {
	dirtyOffsets, err := s.svc.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SyncReply{
		DirtyOffsets: dirtyOffsets,
	}, nil
}

func (s *SeederGrpc) Close(ctx context.Context, args *v1.CloseArgs) (*v1.CloseReply, error) {
	if err := s.svc.Close(ctx); err != nil {
		return nil, err
	}

	return &v1.CloseReply{}, nil
}
