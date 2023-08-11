// Code generated by fRPC Go v0.7.1, DO NOT EDIT.
// source: backend.proto

package v1

import (
	"errors"
	"github.com/loopholelabs/polyglot-go"
	"net"

	"context"
	"crypto/tls"
	"github.com/loopholelabs/frisbee-go"
	"github.com/loopholelabs/frisbee-go/pkg/packet"
	"github.com/rs/zerolog"

	"sync"
)

var (
	NilDecode = errors.New("cannot decode into a nil root struct")
)

type ComPojtingerFelicitasR3MapMountV1ReadAtArgs struct {
	error error
	flags uint8

	Length int32
	Off    int64
}

func NewComPojtingerFelicitasR3MapMountV1ReadAtArgs() *ComPojtingerFelicitasR3MapMountV1ReadAtArgs {
	return &ComPojtingerFelicitasR3MapMountV1ReadAtArgs{}
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
		polyglot.Encoder(b).Int32(x.Length).Int64(x.Off)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	x.Length, err = d.Int32()
	if err != nil {
		return err
	}
	x.Off, err = d.Int64()
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapMountV1ReadAtReply struct {
	error error
	flags uint8

	N int32
	P []byte
}

func NewComPojtingerFelicitasR3MapMountV1ReadAtReply() *ComPojtingerFelicitasR3MapMountV1ReadAtReply {
	return &ComPojtingerFelicitasR3MapMountV1ReadAtReply{}
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtReply) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
		polyglot.Encoder(b).Int32(x.N).Bytes(x.P)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1ReadAtReply) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	x.N, err = d.Int32()
	if err != nil {
		return err
	}
	x.P, err = d.Bytes(nil)
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapMountV1WriteAtArgs struct {
	error error
	flags uint8

	Off int64
	P   []byte
}

func NewComPojtingerFelicitasR3MapMountV1WriteAtArgs() *ComPojtingerFelicitasR3MapMountV1WriteAtArgs {
	return &ComPojtingerFelicitasR3MapMountV1WriteAtArgs{}
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
		polyglot.Encoder(b).Int64(x.Off).Bytes(x.P)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	x.Off, err = d.Int64()
	if err != nil {
		return err
	}
	x.P, err = d.Bytes(nil)
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapMountV1WriteAtReply struct {
	error error
	flags uint8

	Length int32
}

func NewComPojtingerFelicitasR3MapMountV1WriteAtReply() *ComPojtingerFelicitasR3MapMountV1WriteAtReply {
	return &ComPojtingerFelicitasR3MapMountV1WriteAtReply{}
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtReply) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
		polyglot.Encoder(b).Int32(x.Length)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1WriteAtReply) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	x.Length, err = d.Int32()
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapMountV1SyncArgs struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapMountV1SyncArgs() *ComPojtingerFelicitasR3MapMountV1SyncArgs {
	return &ComPojtingerFelicitasR3MapMountV1SyncArgs{}
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncArgs) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncArgs) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapMountV1SyncReply struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapMountV1SyncReply() *ComPojtingerFelicitasR3MapMountV1SyncReply {
	return &ComPojtingerFelicitasR3MapMountV1SyncReply{}
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncReply) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
	}
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapMountV1SyncReply) decode(d *polyglot.Decoder) error {
	if d.Nil() {
		return nil
	}

	var err error
	x.error, err = d.Error()
	if err == nil {
		return nil
	}
	x.flags, err = d.Uint8()
	if err != nil {
		return err
	}
	return nil
}

type Backend interface {
	ReadAt(context.Context, *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) (*ComPojtingerFelicitasR3MapMountV1ReadAtReply, error)
	WriteAt(context.Context, *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) (*ComPojtingerFelicitasR3MapMountV1WriteAtReply, error)
	Sync(context.Context, *ComPojtingerFelicitasR3MapMountV1SyncArgs) (*ComPojtingerFelicitasR3MapMountV1SyncReply, error)
}

const connectionContextKey int = 1000

func SetErrorFlag(flags uint8, error bool) uint8 {
	return flags | 0x2
}
func HasErrorFlag(flags uint8) bool {
	return flags&(1<<1) == 1
}

type Server struct {
	*frisbee.Server
	onClosed func(*frisbee.Async, error)
}

func NewServer(backend Backend, tlsConfig *tls.Config, logger *zerolog.Logger) (*Server, error) {
	var s *Server
	table := make(frisbee.HandlerTable)

	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapMountV1ReadAtArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapMountV1ReadAtReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = backend.ReadAt(ctx, req)
			if err != nil {
				if _, ok := err.(CloseError); ok {
					action = frisbee.CLOSE
				}
				res.Error(outgoing.Content, err)
			} else {
				res.Encode(outgoing.Content)
			}
			outgoing.Metadata.ContentLength = uint32(len(*outgoing.Content))
		}
		return
	}
	table[11] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapMountV1WriteAtArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapMountV1WriteAtReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = backend.WriteAt(ctx, req)
			if err != nil {
				if _, ok := err.(CloseError); ok {
					action = frisbee.CLOSE
				}
				res.Error(outgoing.Content, err)
			} else {
				res.Encode(outgoing.Content)
			}
			outgoing.Metadata.ContentLength = uint32(len(*outgoing.Content))
		}
		return
	}
	table[12] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapMountV1SyncArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapMountV1SyncReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = backend.Sync(ctx, req)
			if err != nil {
				if _, ok := err.(CloseError); ok {
					action = frisbee.CLOSE
				}
				res.Error(outgoing.Content, err)
			} else {
				res.Encode(outgoing.Content)
			}
			outgoing.Metadata.ContentLength = uint32(len(*outgoing.Content))
		}
		return
	}
	var fsrv *frisbee.Server
	var err error
	if tlsConfig != nil {
		fsrv, err = frisbee.NewServer(table, frisbee.WithTLS(tlsConfig), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	} else {
		fsrv, err = frisbee.NewServer(table, frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	}

	fsrv.ConnContext = func(ctx context.Context, conn *frisbee.Async) context.Context {
		return context.WithValue(ctx, connectionContextKey, conn)
	}
	s, err = &Server{
		Server: fsrv,
	}, nil

	fsrv.SetOnClosed(func(async *frisbee.Async, err error) {
		if s.onClosed != nil {
			s.onClosed(async, err)
		}
	})
	return s, err
}

func (s *Server) SetOnClosed(f func(*frisbee.Async, error)) error {
	if f == nil {
		return frisbee.OnClosedNil
	}
	s.onClosed = f
	return nil
}

type subBackendClient struct {
	client            *frisbee.Client
	nextReadAt        uint16
	nextReadAtMu      sync.RWMutex
	inflightReadAt    map[uint16]chan *ComPojtingerFelicitasR3MapMountV1ReadAtReply
	inflightReadAtMu  sync.RWMutex
	nextWriteAt       uint16
	nextWriteAtMu     sync.RWMutex
	inflightWriteAt   map[uint16]chan *ComPojtingerFelicitasR3MapMountV1WriteAtReply
	inflightWriteAtMu sync.RWMutex
	nextSync          uint16
	nextSyncMu        sync.RWMutex
	inflightSync      map[uint16]chan *ComPojtingerFelicitasR3MapMountV1SyncReply
	inflightSyncMu    sync.RWMutex
	nextStreamingID   uint16
	nextStreamingIDMu sync.RWMutex
}
type Client struct {
	*frisbee.Client
	Backend *subBackendClient
}

func NewClient(tlsConfig *tls.Config, logger *zerolog.Logger) (*Client, error) {
	c := new(Client)
	table := make(frisbee.HandlerTable)

	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Backend.inflightReadAtMu.RLock()
		if ch, ok := c.Backend.inflightReadAt[incoming.Metadata.Id]; ok {
			c.Backend.inflightReadAtMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapMountV1ReadAtReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Backend.inflightReadAtMu.RUnlock()
		}
		return
	}
	table[11] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Backend.inflightWriteAtMu.RLock()
		if ch, ok := c.Backend.inflightWriteAt[incoming.Metadata.Id]; ok {
			c.Backend.inflightWriteAtMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapMountV1WriteAtReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Backend.inflightWriteAtMu.RUnlock()
		}
		return
	}
	table[12] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Backend.inflightSyncMu.RLock()
		if ch, ok := c.Backend.inflightSync[incoming.Metadata.Id]; ok {
			c.Backend.inflightSyncMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapMountV1SyncReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Backend.inflightSyncMu.RUnlock()
		}
		return
	}
	var err error
	if tlsConfig != nil {
		c.Client, err = frisbee.NewClient(table, context.Background(), frisbee.WithTLS(tlsConfig), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	} else {
		c.Client, err = frisbee.NewClient(table, context.Background(), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	}

	c.Backend = new(subBackendClient)
	c.Backend.client = c.Client
	c.Backend.nextReadAtMu.Lock()
	c.Backend.nextReadAt = 0
	c.Backend.nextReadAtMu.Unlock()
	c.Backend.inflightReadAt = make(map[uint16]chan *ComPojtingerFelicitasR3MapMountV1ReadAtReply)
	c.Backend.nextWriteAtMu.Lock()
	c.Backend.nextWriteAt = 0
	c.Backend.nextWriteAtMu.Unlock()
	c.Backend.inflightWriteAt = make(map[uint16]chan *ComPojtingerFelicitasR3MapMountV1WriteAtReply)
	c.Backend.nextSyncMu.Lock()
	c.Backend.nextSync = 0
	c.Backend.nextSyncMu.Unlock()
	c.Backend.inflightSync = make(map[uint16]chan *ComPojtingerFelicitasR3MapMountV1SyncReply)
	return c, nil
}

func (c *Client) Connect(addr string, streamHandler ...frisbee.NewStreamHandler) error {
	return c.Client.Connect(addr, func(stream *frisbee.Stream) {})
}

func (c *Client) FromConn(conn net.Conn, streamHandler ...frisbee.NewStreamHandler) error {
	return c.Client.FromConn(conn, func(stream *frisbee.Stream) {})
}

func (c *subBackendClient) ReadAt(ctx context.Context, req *ComPojtingerFelicitasR3MapMountV1ReadAtArgs) (res *ComPojtingerFelicitasR3MapMountV1ReadAtReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapMountV1ReadAtReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 10

	c.nextReadAtMu.Lock()
	c.nextReadAt += 1
	id := c.nextReadAt
	c.nextReadAtMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightReadAtMu.Lock()
	c.inflightReadAt[id] = ch
	c.inflightReadAtMu.Unlock()
	err = c.client.WritePacket(p)
	if err != nil {
		packet.Put(p)
		return
	}
	select {
	case res = <-ch:
		err = res.error
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.inflightReadAtMu.Lock()
	delete(c.inflightReadAt, id)
	c.inflightReadAtMu.Unlock()
	packet.Put(p)
	return
}

func (c *subBackendClient) WriteAt(ctx context.Context, req *ComPojtingerFelicitasR3MapMountV1WriteAtArgs) (res *ComPojtingerFelicitasR3MapMountV1WriteAtReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapMountV1WriteAtReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 11

	c.nextWriteAtMu.Lock()
	c.nextWriteAt += 1
	id := c.nextWriteAt
	c.nextWriteAtMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightWriteAtMu.Lock()
	c.inflightWriteAt[id] = ch
	c.inflightWriteAtMu.Unlock()
	err = c.client.WritePacket(p)
	if err != nil {
		packet.Put(p)
		return
	}
	select {
	case res = <-ch:
		err = res.error
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.inflightWriteAtMu.Lock()
	delete(c.inflightWriteAt, id)
	c.inflightWriteAtMu.Unlock()
	packet.Put(p)
	return
}

func (c *subBackendClient) Sync(ctx context.Context, req *ComPojtingerFelicitasR3MapMountV1SyncArgs) (res *ComPojtingerFelicitasR3MapMountV1SyncReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapMountV1SyncReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 12

	c.nextSyncMu.Lock()
	c.nextSync += 1
	id := c.nextSync
	c.nextSyncMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightSyncMu.Lock()
	c.inflightSync[id] = ch
	c.inflightSyncMu.Unlock()
	err = c.client.WritePacket(p)
	if err != nil {
		packet.Put(p)
		return
	}
	select {
	case res = <-ch:
		err = res.error
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.inflightSyncMu.Lock()
	delete(c.inflightSync, id)
	c.inflightSyncMu.Unlock()
	packet.Put(p)
	return
}

type CloseError struct {
	err error
}

func NewCloseError(err error) CloseError {
	return CloseError{err: err}
}

func (e CloseError) Error() string {
	return e.err.Error()
}
