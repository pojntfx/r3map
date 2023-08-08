package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
)

type LeecherFrpc struct {
	client *v1.Client
}

func NewLeecherFrpc(client *v1.Client) *SeederRemote {
	l := &LeecherFrpc{client}

	return &SeederRemote{
		ReadAt: l.ReadAt,
		Size:   l.Size,
		Track:  l.Track,
		Sync:   l.Sync,
		Close:  l.Close,
	}
}

func (l *LeecherFrpc) ReadAt(ctx context.Context, length int, off int64) (r ReadAtResponse, err error) {
	res, err := l.client.Seeder.ReadAt(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1ReadAtArgs{
		Length: int32(length),
		Off:    off,
	})
	if err != nil {
		return ReadAtResponse{}, err
	}

	return ReadAtResponse{
		N: int(res.N),
		P: res.P,
	}, err
}

func (l *LeecherFrpc) Size(ctx context.Context) (int64, error) {
	res, err := l.client.Seeder.Size(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1SizeArgs{})
	if err != nil {
		return -1, err
	}

	return res.N, nil
}

func (l *LeecherFrpc) Track(ctx context.Context) error {
	if _, err := l.client.Seeder.Track(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1TrackArgs{}); err != nil {
		return err
	}

	return nil
}

func (l *LeecherFrpc) Sync(ctx context.Context) ([]int64, error) {
	res, err := l.client.Seeder.Sync(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1SyncArgs{})
	if err != nil {
		return []int64{}, err
	}

	return res.DirtyOffsets, nil
}

func (l *LeecherFrpc) Close(ctx context.Context) error {
	if _, err := l.client.Seeder.Close(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1CloseArgs{}); err != nil {
		return err
	}

	return nil
}
