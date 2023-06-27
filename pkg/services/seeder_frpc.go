package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/v1"
)

//go:generate sh -c "mkdir -p ../api/proto/v1 && protoc --go_out=../api/proto/v1 --go_opt=paths=source_relative --go-frpc_out=../api/proto/v1 --go-frpc_opt=paths=source_relative --proto_path=../../api/proto/v1 ../../api/proto/v1/*.proto"

type SeederFrpc struct {
	svc *Seeder
}

func NewSeederFrpc(svc *Seeder) *SeederFrpc {
	return &SeederFrpc{svc}
}

func (s *SeederFrpc) ReadAt(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapV1ReadAtArgs) (*v1.ComPojtingerFelicitasR3MapV1ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.Length), args.Off)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapV1ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *SeederFrpc) Size(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapV1SizeArgs) (*v1.ComPojtingerFelicitasR3MapV1SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapV1SizeReply{
		N: size,
	}, nil
}

func (s *SeederFrpc) Track(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapV1TrackArgs) (*v1.ComPojtingerFelicitasR3MapV1TrackReply, error) {
	if err := s.svc.Track(ctx); err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapV1TrackReply{}, nil
}

func (s *SeederFrpc) Sync(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapV1SyncArgs) (*v1.ComPojtingerFelicitasR3MapV1SyncReply, error) {
	dirtyOffsets, err := s.svc.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapV1SyncReply{
		DirtyOffsets: dirtyOffsets,
	}, nil
}

func (s *SeederFrpc) Close(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapV1CloseArgs) (*v1.ComPojtingerFelicitasR3MapV1CloseReply, error) {
	if err := s.svc.Close(ctx); err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapV1CloseReply{}, nil
}
