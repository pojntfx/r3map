module github.com/pojntfx/r3map

go 1.20

require (
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/edsrzf/mmap-go v1.1.0
	github.com/gocql/gocql v1.4.0
	github.com/loopholelabs/frisbee-go v0.7.1
	github.com/loopholelabs/polyglot-go v0.5.1
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/pojntfx/dudirekta v0.5.0
	github.com/pojntfx/go-nbd v0.1.9
	github.com/redis/go-redis/v9 v9.0.5
	github.com/rs/zerolog v1.29.1
	github.com/schollz/progressbar/v3 v3.13.1
	google.golang.org/grpc v1.56.1
	google.golang.org/protobuf v1.31.0
	storj.io/drpc v0.0.33
)

require (
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/loopholelabs/common v0.4.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pilebones/go-udev v0.9.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/teivah/broadcast v0.1.0 // indirect
	github.com/zeebo/errs v1.3.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	golang.org/x/crypto v0.10.0 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/term v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230626202813-9b080da550b3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)

replace github.com/gocql/gocql => github.com/scylladb/gocql v1.10.0
