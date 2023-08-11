package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
)

type BackendRemoteFrpc struct {
	client *v1.Client
}

func NewBackendRemoteFrpc(client *v1.Client) *BackendRemote {
	l := &BackendRemoteFrpc{client}

	return &BackendRemote{
		ReadAt:  l.ReadAt,
		WriteAt: l.WriteAt,
		Sync:    l.Sync,
	}
}

func (l *BackendRemoteFrpc) ReadAt(ctx context.Context, length int, off int64) (r ReadAtResponse, err error) {
	res, err := l.client.Backend.ReadAt(ctx, &v1.ComPojtingerFelicitasR3MapMountV1ReadAtArgs{
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

func (l *BackendRemoteFrpc) WriteAt(ctx context.Context, p []byte, off int64) (n int, err error) {
	res, err := l.client.Backend.WriteAt(ctx, &v1.ComPojtingerFelicitasR3MapMountV1WriteAtArgs{
		Off: off,
		P:   p,
	})
	if err != nil {
		return 0, err
	}

	return int(res.Length), nil
}

func (l *BackendRemoteFrpc) Sync(ctx context.Context) error {
	if _, err := l.client.Backend.Sync(ctx, &v1.ComPojtingerFelicitasR3MapMountV1SyncArgs{}); err != nil {
		return err
	}

	return nil
}
