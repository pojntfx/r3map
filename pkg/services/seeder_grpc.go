package services

import (
	"context"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/v1"
)

//go:generate sh -c "rm -rf ../api/proto/v1 && mkdir -p ../api/proto/v1 && protoc --go_out=../api/proto/v1 --go_opt=paths=source_relative --go-grpc_out=../api/proto/v1 --go-grpc_opt=paths=source_relative --proto_path=../../api/proto/v1 ../../api/proto/v1/*.proto"
