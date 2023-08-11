package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
)

type SeederRemoteFrpc struct {
	client *v1.Client
}

func NewSeederRemoteFrpc(client *v1.Client) *SeederRemote {
	l := &SeederRemoteFrpc{client}

	return &SeederRemote{
		ReadAt: l.ReadAt,
		Track:  l.Track,
		Sync:   l.Sync,
		Close:  l.Close,
	}
}

func (l *SeederRemoteFrpc) ReadAt(ctx context.Context, length int, off int64) (r ReadAtResponse, err error) {
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

func (l *SeederRemoteFrpc) Track(ctx context.Context) error {
	if _, err := l.client.Seeder.Track(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1TrackArgs{}); err != nil {
		return err
	}

	return nil
}

func (l *SeederRemoteFrpc) Sync(ctx context.Context) ([]int64, error) {
	res, err := l.client.Seeder.Sync(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1SyncArgs{})
	if err != nil {
		return []int64{}, err
	}

	return res.DirtyOffsets, nil
}

func (l *SeederRemoteFrpc) Close(ctx context.Context) error {
	if _, err := l.client.Seeder.Close(ctx, &v1.ComPojtingerFelicitasR3MapMigrationV1CloseArgs{}); err != nil {
		return err
	}

	return nil
}
