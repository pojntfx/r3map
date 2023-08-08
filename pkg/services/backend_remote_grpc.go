package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
)

type BackendRemoteGrpc struct {
	client v1.BackendClient
}

func NewBackendRemoteGrpc(client v1.BackendClient) *BackendRemote {
	l := &BackendRemoteGrpc{client}

	return &BackendRemote{
		ReadAt:  l.ReadAt,
		WriteAt: l.WriteAt,
		Size:    l.Size,
		Sync:    l.Sync,
	}
}

func (l *BackendRemoteGrpc) ReadAt(ctx context.Context, length int, off int64) (r ReadAtResponse, err error) {
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

func (l *BackendRemoteGrpc) WriteAt(ctx context.Context, p []byte, off int64) (n int, err error) {
	res, err := l.client.WriteAt(ctx, &v1.WriteAtArgs{
		Off: off,
		P:   p,
	})
	if err != nil {
		return 0, err
	}

	return int(res.GetLength()), nil
}

func (l *BackendRemoteGrpc) Size(ctx context.Context) (int64, error) {
	res, err := l.client.Size(ctx, &v1.SizeArgs{})
	if err != nil {
		return 0, err
	}

	return res.GetSize(), nil
}

func (l *BackendRemoteGrpc) Sync(ctx context.Context) error {
	if _, err := l.client.Sync(ctx, &v1.SyncArgs{}); err != nil {
		return err
	}

	return nil
}
