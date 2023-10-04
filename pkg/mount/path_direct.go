package mount

import (
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/utils"
)

type DirectPathMount struct {
	e *server.Export
	f *os.File

	serverOptions *server.Options
	clientOptions *client.Options

	scs []net.Conn

	ccs []net.Conn

	errs chan error
}

func NewDirectPathMount(
	b backend.Backend,
	f *os.File,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *DirectPathMount {
	return &DirectPathMount{
		e: &server.Export{
			Name:    "default",
			Backend: b,
		},
		f: f,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		ccs: []net.Conn{},

		errs: make(chan error),
	}
}

func (d *DirectPathMount) Wait() error {
	for err := range d.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DirectPathMount) Open() error {
	// fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	// if err != nil {
	// 	return err
	// }

	addrDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	addr := filepath.Join(addrDir, "r3map.sock")

	var listenerWg sync.WaitGroup
	listenerWg.Add(1)

	go func() {
		lis, err := net.Listen("unix", addr)
		if err != nil {
			d.errs <- err

			return
		}
		// TODO: Add lis to struct and close it on struct.Close

		listenerWg.Done()

		for {
			c, err := lis.Accept()
			if err != nil {
				// TODO: We should not return here, but continue instead, with some sort of logging since its no longer 1:1
				if !utils.IsClosedErr(err) {
					d.errs <- err
				}

				return
			}
			d.scs = append(d.scs, c)

			go func() {
				if err := server.Handle(
					c,
					[]*server.Export{d.e},
					d.serverOptions,
				); err != nil {
					if !utils.IsClosedErr(err) {
						d.errs <- err
					}

					return
				}
			}()
		}
	}()

	ready := make(chan struct{})

	go func() {
		listenerWg.Wait()

		for i := 0; i < 2; i++ {
			c, err := net.Dial("unix", addr)
			if err != nil {
				d.errs <- err

				return
			}
			d.ccs = append(d.ccs, c)
		}

		if d.clientOptions == nil {
			d.clientOptions = &client.Options{}
		}

		d.clientOptions.OnConnected = func() {
			ready <- struct{}{}
		}

		if err := client.Connect(d.ccs, d.f, d.clientOptions); err != nil {
			if !utils.IsClosedErr(err) {
				d.errs <- err
			}

			return
		}
	}()

	<-ready

	return nil
}

func (d *DirectPathMount) Close() error {
	_ = client.Disconnect(d.f)

	for _, cc := range d.ccs {
		_ = cc.Close()
	}
	d.ccs = []net.Conn{}

	// if d.cf != nil {
	// 	_ = d.cf.Close()
	// }

	for _, sc := range d.scs {
		_ = sc.Close()
	}
	d.scs = []net.Conn{}

	// TODO: Also close the listener

	// if d.sf != nil {
	// 	_ = d.sf.Close()
	// }

	if d.errs != nil {
		close(d.errs)

		d.errs = nil
	}

	return nil
}

func (d *DirectPathMount) Sync() error {
	return nil
}

func (d *DirectPathMount) SwapBackend(b backend.Backend) {
	d.e.Backend = b
}
