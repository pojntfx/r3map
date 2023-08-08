package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
)

//go:generate sh -c "mkdir -p ../api/frpc/migration/v1 && protoc --go-frpc_out=../api/frpc/migration/v1 --go-frpc_opt=paths=source_relative --proto_path=../../api/proto/migration/v1 ../../api/proto/migration/v1/*.proto"

type SeederServiceFrpc struct {
	svc *SeederService
}

func NewSeederServiceFrpc(svc *SeederService) *SeederServiceFrpc {
	return &SeederServiceFrpc{svc}
}

func (s *SeederServiceFrpc) ReadAt(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMigrationV1ReadAtArgs) (*v1.ComPojtingerFelicitasR3MapMigrationV1ReadAtReply, error) {
	res, err := s.svc.ReadAt(ctx, int(args.Length), args.Off)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMigrationV1ReadAtReply{
		N: int32(res.N),
		P: res.P,
	}, nil
}

func (s *SeederServiceFrpc) Size(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMigrationV1SizeArgs) (*v1.ComPojtingerFelicitasR3MapMigrationV1SizeReply, error) {
	size, err := s.svc.Size(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMigrationV1SizeReply{
		N: size,
	}, nil
}

func (s *SeederServiceFrpc) Track(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMigrationV1TrackArgs) (*v1.ComPojtingerFelicitasR3MapMigrationV1TrackReply, error) {
	if err := s.svc.Track(ctx); err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMigrationV1TrackReply{}, nil
}

func (s *SeederServiceFrpc) Sync(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMigrationV1SyncArgs) (*v1.ComPojtingerFelicitasR3MapMigrationV1SyncReply, error) {
	dirtyOffsets, err := s.svc.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMigrationV1SyncReply{
		DirtyOffsets: dirtyOffsets,
	}, nil
}

func (s *SeederServiceFrpc) Close(ctx context.Context, args *v1.ComPojtingerFelicitasR3MapMigrationV1CloseArgs) (*v1.ComPojtingerFelicitasR3MapMigrationV1CloseReply, error) {
	if err := s.svc.Close(ctx); err != nil {
		return nil, err
	}

	return &v1.ComPojtingerFelicitasR3MapMigrationV1CloseReply{}, nil
}
