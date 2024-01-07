package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gocql/gocql"
	"github.com/minio/minio-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/ltsrpc/pkg/rpc"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	backendTypeFile      = "file"
	backendTypeMemory    = "memory"
	backendTypeDirectory = "directory"
	backendTypeLtsrpc    = "ltsrpc"
	backendTypeGrpc      = "grpc"
	backendTypeFrpc      = "frpc"
	backendTypeRedis     = "redis"
	backendTypeS3        = "s3"
	backendTypeCassandra = "cassandra"
)

var (
	knownBackendTypes = []string{backendTypeFile, backendTypeMemory, backendTypeDirectory, backendTypeLtsrpc, backendTypeGrpc, backendTypeFrpc, backendTypeRedis, backendTypeS3, backendTypeCassandra}

	errUnknownBackend     = errors.New("unknown backend")
	errNoPeerFound        = errors.New("no peer found")
	errMissingCredentials = errors.New("missing credentials")
	errMissingPassword    = errors.New("missing password")
)

func main() {
	s := flag.Int64("size", 536870912, "Size of the memory region, file to allocate or to size assume in case of the ltsrpc/gRPC/fRPC remotes")
	chunkSize := flag.Int64("chunk-size", client.MaximumBlockSize, "Chunk size to use")

	remoteBackend := flag.String(
		"remote-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Remote backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	remoteLocation := flag.String("remote-backend-location", filepath.Join(os.TempDir(), "remote"), "Remote backend's remote address (for ltsrpc/gRPC/fRPC, e.g. localhost:1337), URI (for redis, e.g. redis://username:password@localhost:6379/0, or S3, e.g. http://accessKey:secretKey@localhost:9000?bucket=bucket&prefix=prefix or Cassandra/ScyllaDB, e.g. cassandra://username:password@localhost:9042?keyspace=keyspace&table=table&prefix=prefix) or directory (for directory backend)")
	remoteChunking := flag.Bool("remote-backend-chunking", false, "Whether the remote backend requires to be interfaced with in fixed chunks")

	outputBackend := flag.String(
		"output-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Output backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	outputLocation := flag.String("output-backend-location", filepath.Join(os.TempDir(), "output"), "Output backend's output address (for ltsrpc/gRPC/fRPC, e.g. localhost:1337), URI (for redis, e.g. redis://username:password@localhost:6379/0, or S3, e.g. http://accessKey:secretKey@localhost:9000?bucket=bucket&prefix=prefix or Cassandra/ScyllaDB, e.g. cassandra://username:password@localhost:9042?keyspace=keyspace&table=table&prefix=prefix) or directory (for directory backend)")
	outputChunking := flag.Bool("output-backend-chunking", false, "Whether the output backend requires to be interfaced with in fixed chunks")

	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")

	check := flag.Bool("check", true, "Whether to check read and write results against expected data")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		remote backend.Backend
		output backend.Backend
	)
	for _, config := range []struct {
		backendInstance *backend.Backend
		backendType     string
		backendLocation string
		chunking        bool
	}{
		{
			&remote,
			*remoteBackend,
			*remoteLocation,
			*remoteChunking,
		},
		{
			&output,
			*outputBackend,
			*outputLocation,
			*outputChunking,
		},
	} {
		switch config.backendType {
		case backendTypeMemory:
			*config.backendInstance = backend.NewMemoryBackend(make([]byte, *s))

		case backendTypeFile:
			file, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			defer os.RemoveAll(file.Name())

			if err := file.Truncate(*s); err != nil {
				panic(err)
			}

			*config.backendInstance = backend.NewFileBackend(file)

		case backendTypeDirectory:
			if err := os.MkdirAll(config.backendLocation, os.ModePerm); err != nil {
				panic(err)
			}

			*config.backendInstance = lbackend.NewDirectoryBackend(config.backendLocation, *s, *chunkSize, 512, *verbose)

		case backendTypeLtsrpc:
			ready := make(chan struct{})
			registry := rpc.NewRegistry[services.BackendRemote, json.RawMessage](
				struct{}{},

				time.Second*10,
				ctx,
				&rpc.Options{
					OnClientConnect: func(remoteID string) {
						ready <- struct{}{}
					},
				},
			)

			conn, err := net.Dial("tcp", config.backendLocation)
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			go func() {
				encoder := json.NewEncoder(conn)
				decoder := json.NewDecoder(conn)

				if err := registry.LinkStream(
					func(v rpc.Message[json.RawMessage]) error {
						return encoder.Encode(v)
					},
					func(v *rpc.Message[json.RawMessage]) error {
						return decoder.Decode(v)
					},

					func(v any) (json.RawMessage, error) {
						b, err := json.Marshal(v)
						if err != nil {
							return nil, err
						}

						return json.RawMessage(b), nil
					},
					func(data json.RawMessage, v any) error {
						return json.Unmarshal([]byte(data), v)
					},
				); err != nil {
					if !utils.IsClosedErr(err) {
						panic(err)
					}
				}
			}()

			<-ready

			var peer *services.BackendRemote

			_ = registry.ForRemotes(func(remoteID string, remote services.BackendRemote) error {
				peer = &remote

				return nil
			})

			if peer == nil {
				panic(errNoPeerFound)
			}

			*config.backendInstance = lbackend.NewRPCBackend(ctx, peer, *s, *verbose)

		case backendTypeGrpc:
			conn, err := grpc.Dial(config.backendLocation, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			*config.backendInstance = lbackend.NewRPCBackend(ctx, services.NewBackendRemoteGrpc(v1proto.NewBackendClient(conn)), *s, *verbose)

		case backendTypeFrpc:
			client, err := v1frpc.NewClient(nil, nil)
			if err != nil {
				panic(err)
			}

			if err := client.Connect(config.backendLocation); err != nil {
				panic(err)
			}
			defer client.Close()

			*config.backendInstance = lbackend.NewRPCBackend(ctx, services.NewBackendRemoteFrpc(client), *s, *verbose)

		case backendTypeRedis:
			options, err := redis.ParseURL(config.backendLocation)
			if err != nil {
				panic(err)
			}

			client := redis.NewClient(options)
			defer client.Close()

			*config.backendInstance = lbackend.NewRedisBackend(ctx, client, *s, *verbose)

		case backendTypeS3:
			u, err := url.Parse(config.backendLocation)
			if err != nil {
				panic(err)
			}

			user := u.User
			if user == nil {
				panic(errMissingCredentials)
			}

			pw, ok := user.Password()
			if !ok {
				panic(errMissingPassword)
			}

			client, err := minio.New(u.Host, user.Username(), pw, u.Scheme == "https")
			if err != nil {
				panic(err)
			}

			bucketName := u.Query().Get("bucket")

			bucketExists, err := client.BucketExists(bucketName)
			if err != nil {
				panic(err)
			}

			if !bucketExists {
				if err := client.MakeBucket(bucketName, ""); err != nil {
					panic(err)
				}
			}

			*config.backendInstance = lbackend.NewS3Backend(ctx, client, bucketName, u.Query().Get("prefix"), *s, *verbose)

		case backendTypeCassandra:
			u, err := url.Parse(config.backendLocation)
			if err != nil {
				panic(err)
			}

			user := u.User
			if user == nil {
				panic(errMissingCredentials)
			}

			pw, ok := user.Password()
			if !ok {
				panic(errMissingPassword)
			}

			cluster := gocql.NewCluster(u.Host)
			cluster.Consistency = gocql.Quorum
			cluster.Authenticator = gocql.PasswordAuthenticator{
				Username: user.Username(),
				Password: pw,
			}

			if u.Scheme == "cassandrasecure" {
				cluster.SslOpts = &gocql.SslOptions{
					EnableHostVerification: true,
				}
			}

			keyspaceName := u.Query().Get("keyspace")
			{
				setupSession, err := cluster.CreateSession()
				if err != nil {
					panic(err)
				}

				if err := setupSession.Query(`create keyspace if not exists ` + keyspaceName + ` with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`).Exec(); err != nil {
					setupSession.Close()

					panic(err)
				}

				setupSession.Close()
			}
			cluster.Keyspace = keyspaceName

			session, err := cluster.CreateSession()
			if err != nil {
				panic(err)
			}
			defer session.Close()

			tableName := u.Query().Get("table")

			if err := session.Query(`create table if not exists ` + tableName + ` (key blob primary key, data blob)`).Exec(); err != nil {
				panic(err)
			}

			*config.backendInstance = lbackend.NewCassandraBackend(session, tableName, u.Query().Get("prefix"), *s, *verbose)

		default:
			panic(errUnknownBackend)
		}

		if config.chunking {
			*config.backendInstance = lbackend.NewReaderAtBackend(
				chunks.NewArbitraryReadWriterAt(
					*config.backendInstance,
					*chunkSize,
				),
				(*config.backendInstance).Size,
				(*config.backendInstance).Sync,
				false,
			)
		}
	}

	size := (*s / *chunkSize) * *chunkSize

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			remote,
			0), rand.Reader, size); err != nil {
		panic(err)
	}

	devPath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	devFile, err := os.Open(devPath)
	if err != nil {
		panic(err)
	}
	defer devFile.Close()

	beforeOpen := time.Now()

	var (
		mountedReader io.Reader
		m             any
		sync          func() error
	)
	if *slice {
		mnt := mount.NewDirectSliceMount(
			remote,
			devFile,

			nil,
			nil,
		)

		go func() {
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		mountedSlice, err := mnt.Open()
		if err != nil {
			panic(err)
		}

		defer func() {
			beforeClose := time.Now()

			_ = mnt.Close()

			afterClose := time.Since(beforeClose)

			fmt.Printf("Close: %v\n", afterClose)
		}()

		mountedReader, m, sync = bytes.NewReader(mountedSlice), mountedSlice, mnt.Sync
	} else {
		mnt := mount.NewDirectFileMount(
			remote,
			devFile,

			nil,
			nil,
		)

		go func() {
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		mountedFile, err := mnt.Open()
		if err != nil {
			panic(err)
		}

		defer func() {
			beforeClose := time.Now()

			_ = mnt.Close()

			afterClose := time.Since(beforeClose)

			fmt.Printf("Close: %v\n", afterClose)
		}()

		mountedReader, m, sync = mountedFile, mountedFile, mnt.Sync
	}

	afterOpen := time.Since(beforeOpen)

	fmt.Printf("Open: %v\n", afterOpen)

	beforeFirstTwoChunks := time.Now()

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			output,
			0,
		), mountedReader, *chunkSize*2); err != nil {
		panic(err)
	}

	afterFirstTwoChunks := time.Since(beforeFirstTwoChunks)

	fmt.Printf("Latency till first two chunks: %v\n", afterFirstTwoChunks)

	if *slice {
		mountedReader = bytes.NewReader(m.([]byte))
	} else {
		if _, err := m.(*os.File).Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
	}

	beforeRead := time.Now()

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			output,
			0,
		), mountedReader, size); err != nil {
		panic(err)
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

	validate := func(output io.ReaderAt) error {
		if *slice {
			mountedReader = bytes.NewReader(m.([]byte))
		} else {
			if _, err := m.(*os.File).Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
		}

		remoteHash := xxhash.New()
		if _, err := io.Copy(
			remoteHash,
			io.NewSectionReader(
				remote,
				0,
				size,
			),
		); err != nil {
			return err
		}

		outputHash := xxhash.New()
		if _, err := io.Copy(
			outputHash,
			io.NewSectionReader(
				output,
				0,
				size,
			),
		); err != nil {
			return err
		}

		if remoteHash.Sum64() != outputHash.Sum64() {
			return errors.New("remote and output hashes don't match")
		}

		return nil
	}

	if *check {
		if err := validate(output); err != nil {
			panic(err)
		}

		fmt.Println("Read check: Passed")
	}

	beforeWrite := time.Now()

	if *slice {
		if _, err := rand.Read(m.([]byte)); err != nil {
			panic(err)
		}
	} else {
		if _, err := io.CopyN(m.(*os.File), rand.Reader, size); err != nil {
			panic(err)
		}
	}

	afterWrite := time.Since(beforeWrite)

	throughputMB = float64(size) / (1024 * 1024) / afterWrite.Seconds()

	fmt.Printf("Write throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

	if *check {
		if *slice {
			mountedReader = bytes.NewReader(m.([]byte))
		} else {
			if _, err := m.(*os.File).Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
		}

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				output,
				0,
			), mountedReader, size); err != nil {
			panic(err)
		}

		beforeSync := time.Now()

		if err := sync(); err != nil {
			panic(err)
		}

		afterSync := time.Since(beforeSync)

		fmt.Printf("Sync: %v\n", afterSync)

		if err := validate(output); err != nil {
			panic(err)
		}

		fmt.Println("Write check: Passed")
	}
}
