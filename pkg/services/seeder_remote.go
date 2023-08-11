package services

import "context"

type SeederRemote struct {
	ReadAt func(context context.Context, length int, off int64) (r ReadAtResponse, err error)
	Track  func(context context.Context) error
	Sync   func(context context.Context) ([]int64, error)
	Close  func(context context.Context) error
}
