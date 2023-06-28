package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/v1"
)

type SeederDrpc struct {
	v1.DRPCSeederUnimplementedServer

	svc *Seeder
}

func NewSeederDrpc(svc *Seeder) *SeederDrpc {
	return &SeederDrpc{v1.DRPCSeederUnimplementedServer{}, svc}
}

func (s *SeederDrpc) ReadAt(ctx context.Context, args *v1.ReadAtArgs) (*v1.ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.GetLength()), args.GetOff())
	if err != nil {
		return nil, err
	}

	return &v1.ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *SeederDrpc) Size(ctx context.Context, args *v1.SizeArgs) (*v1.SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SizeReply{
		N: size,
	}, nil
}

func (s *SeederDrpc) Track(ctx context.Context, args *v1.TrackArgs) (*v1.TrackReply, error) {
	if err := s.svc.Track(ctx); err != nil {
		return nil, err
	}

	return &v1.TrackReply{}, nil
}

func (s *SeederDrpc) Sync(ctx context.Context, args *v1.SyncArgs) (*v1.SyncReply, error) {
	dirtyOffsets, err := s.svc.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SyncReply{
		DirtyOffsets: dirtyOffsets,
	}, nil
}

func (s *SeederDrpc) Close(ctx context.Context, args *v1.CloseArgs) (*v1.CloseReply, error) {
	if err := s.svc.Close(ctx); err != nil {
		return nil, err
	}

	return &v1.CloseReply{}, nil
}
