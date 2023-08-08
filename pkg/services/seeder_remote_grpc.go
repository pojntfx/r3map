package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
)

type SeederRemoteGrpc struct {
	client v1.SeederClient
}

func NewSeederRemoteGrpc(client v1.SeederClient) *SeederRemote {
	l := &SeederRemoteGrpc{client}

	return &SeederRemote{
		ReadAt: l.ReadAt,
		Size:   l.Size,
		Track:  l.Track,
		Sync:   l.Sync,
		Close:  l.Close,
	}
}

func (l *SeederRemoteGrpc) ReadAt(ctx context.Context, length int, off int64) (r ReadAtResponse, err error) {
	res, err := l.client.ReadAt(ctx, &v1.ReadAtArgs{
		Length: int32(length),
		Off:    off,
	})
	if err != nil {
		return ReadAtResponse{}, err
	}

	return ReadAtResponse{
		N: int(res.GetN()),
		P: res.GetP(),
	}, err
}

func (l *SeederRemoteGrpc) Size(ctx context.Context) (int64, error) {
	res, err := l.client.Size(ctx, &v1.SizeArgs{})
	if err != nil {
		return -1, err
	}

	return res.GetN(), nil
}

func (l *SeederRemoteGrpc) Track(ctx context.Context) error {
	if _, err := l.client.Track(ctx, &v1.TrackArgs{}); err != nil {
		return err
	}

	return nil
}

func (l *SeederRemoteGrpc) Sync(ctx context.Context) ([]int64, error) {
	res, err := l.client.Sync(ctx, &v1.SyncArgs{})
	if err != nil {
		return []int64{}, err
	}

	return res.GetDirtyOffsets(), nil
}

func (l *SeederRemoteGrpc) Close(ctx context.Context) error {
	if _, err := l.client.Close(ctx, &v1.CloseArgs{}); err != nil {
		return err
	}

	return nil
}
