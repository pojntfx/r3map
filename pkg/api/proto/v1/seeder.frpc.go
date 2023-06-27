// Code generated by fRPC Go v0.7.1, DO NOT EDIT.
// source: seeder.proto

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

type ComPojtingerFelicitasR3MapV1ReadAtArgs struct {
	error error
	flags uint8

	Length int32
	Off    int64
}

func NewComPojtingerFelicitasR3MapV1ReadAtArgs() *ComPojtingerFelicitasR3MapV1ReadAtArgs {
	return &ComPojtingerFelicitasR3MapV1ReadAtArgs{}
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtArgs) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1ReadAtArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtArgs) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1ReadAtReply struct {
	error error
	flags uint8

	N int32
	P []byte
}

func NewComPojtingerFelicitasR3MapV1ReadAtReply() *ComPojtingerFelicitasR3MapV1ReadAtReply {
	return &ComPojtingerFelicitasR3MapV1ReadAtReply{}
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtReply) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1ReadAtReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1ReadAtReply) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1SizeArgs struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1SizeArgs() *ComPojtingerFelicitasR3MapV1SizeArgs {
	return &ComPojtingerFelicitasR3MapV1SizeArgs{}
}

func (x *ComPojtingerFelicitasR3MapV1SizeArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1SizeArgs) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1SizeArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1SizeArgs) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1SizeReply struct {
	error error
	flags uint8

	N int64
}

func NewComPojtingerFelicitasR3MapV1SizeReply() *ComPojtingerFelicitasR3MapV1SizeReply {
	return &ComPojtingerFelicitasR3MapV1SizeReply{}
}

func (x *ComPojtingerFelicitasR3MapV1SizeReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1SizeReply) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)
		polyglot.Encoder(b).Int64(x.N)
	}
}

func (x *ComPojtingerFelicitasR3MapV1SizeReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1SizeReply) decode(d *polyglot.Decoder) error {
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
	x.N, err = d.Int64()
	if err != nil {
		return err
	}
	return nil
}

type ComPojtingerFelicitasR3MapV1TrackArgs struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1TrackArgs() *ComPojtingerFelicitasR3MapV1TrackArgs {
	return &ComPojtingerFelicitasR3MapV1TrackArgs{}
}

func (x *ComPojtingerFelicitasR3MapV1TrackArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1TrackArgs) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1TrackArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1TrackArgs) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1TrackReply struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1TrackReply() *ComPojtingerFelicitasR3MapV1TrackReply {
	return &ComPojtingerFelicitasR3MapV1TrackReply{}
}

func (x *ComPojtingerFelicitasR3MapV1TrackReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1TrackReply) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1TrackReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1TrackReply) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1SyncArgs struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1SyncArgs() *ComPojtingerFelicitasR3MapV1SyncArgs {
	return &ComPojtingerFelicitasR3MapV1SyncArgs{}
}

func (x *ComPojtingerFelicitasR3MapV1SyncArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1SyncArgs) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1SyncArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1SyncArgs) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1SyncReply struct {
	error error
	flags uint8

	DirtyOffsets []int64
}

func NewComPojtingerFelicitasR3MapV1SyncReply() *ComPojtingerFelicitasR3MapV1SyncReply {
	return &ComPojtingerFelicitasR3MapV1SyncReply{}
}

func (x *ComPojtingerFelicitasR3MapV1SyncReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1SyncReply) Encode(b *polyglot.Buffer) {
	if x == nil {
		polyglot.Encoder(b).Nil()
	} else {
		if x.error != nil {
			polyglot.Encoder(b).Error(x.error)
			return
		}
		polyglot.Encoder(b).Uint8(x.flags)

		polyglot.Encoder(b).Slice(uint32(len(x.DirtyOffsets)), polyglot.Int64Kind)
		for _, v := range x.DirtyOffsets {
			polyglot.Encoder(b).Int64(v)
		}
	}
}

func (x *ComPojtingerFelicitasR3MapV1SyncReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1SyncReply) decode(d *polyglot.Decoder) error {
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
	var sliceSize uint32
	sliceSize, err = d.Slice(polyglot.Int64Kind)
	if err != nil {
		return err
	}
	if uint32(len(x.DirtyOffsets)) != sliceSize {
		x.DirtyOffsets = make([]int64, sliceSize)
	}
	for i := uint32(0); i < sliceSize; i++ {
		x.DirtyOffsets[i], err = d.Int64()
		if err != nil {
			return err
		}
	}
	return nil
}

type ComPojtingerFelicitasR3MapV1CloseArgs struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1CloseArgs() *ComPojtingerFelicitasR3MapV1CloseArgs {
	return &ComPojtingerFelicitasR3MapV1CloseArgs{}
}

func (x *ComPojtingerFelicitasR3MapV1CloseArgs) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1CloseArgs) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1CloseArgs) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1CloseArgs) decode(d *polyglot.Decoder) error {
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

type ComPojtingerFelicitasR3MapV1CloseReply struct {
	error error
	flags uint8
}

func NewComPojtingerFelicitasR3MapV1CloseReply() *ComPojtingerFelicitasR3MapV1CloseReply {
	return &ComPojtingerFelicitasR3MapV1CloseReply{}
}

func (x *ComPojtingerFelicitasR3MapV1CloseReply) Error(b *polyglot.Buffer, err error) {
	polyglot.Encoder(b).Error(err)
}

func (x *ComPojtingerFelicitasR3MapV1CloseReply) Encode(b *polyglot.Buffer) {
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

func (x *ComPojtingerFelicitasR3MapV1CloseReply) Decode(b []byte) error {
	if x == nil {
		return NilDecode
	}
	d := polyglot.GetDecoder(b)
	defer d.Return()
	return x.decode(d)
}

func (x *ComPojtingerFelicitasR3MapV1CloseReply) decode(d *polyglot.Decoder) error {
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

type Seeder interface {
	ReadAt(context.Context, *ComPojtingerFelicitasR3MapV1ReadAtArgs) (*ComPojtingerFelicitasR3MapV1ReadAtReply, error)
	Size(context.Context, *ComPojtingerFelicitasR3MapV1SizeArgs) (*ComPojtingerFelicitasR3MapV1SizeReply, error)
	Track(context.Context, *ComPojtingerFelicitasR3MapV1TrackArgs) (*ComPojtingerFelicitasR3MapV1TrackReply, error)
	Sync(context.Context, *ComPojtingerFelicitasR3MapV1SyncArgs) (*ComPojtingerFelicitasR3MapV1SyncReply, error)
	Close(context.Context, *ComPojtingerFelicitasR3MapV1CloseArgs) (*ComPojtingerFelicitasR3MapV1CloseReply, error)
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

func NewServer(seeder Seeder, tlsConfig *tls.Config, logger *zerolog.Logger) (*Server, error) {
	var s *Server
	table := make(frisbee.HandlerTable)

	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapV1ReadAtArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapV1ReadAtReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = seeder.ReadAt(ctx, req)
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
		req := NewComPojtingerFelicitasR3MapV1SizeArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapV1SizeReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = seeder.Size(ctx, req)
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
		req := NewComPojtingerFelicitasR3MapV1TrackArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapV1TrackReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = seeder.Track(ctx, req)
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
	table[13] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapV1SyncArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapV1SyncReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = seeder.Sync(ctx, req)
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
	table[14] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewComPojtingerFelicitasR3MapV1CloseArgs()
		err := req.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
		if err == nil {
			var res *ComPojtingerFelicitasR3MapV1CloseReply
			outgoing = incoming
			outgoing.Content.Reset()
			res, err = seeder.Close(ctx, req)
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

type subSeederClient struct {
	client            *frisbee.Client
	nextReadAt        uint16
	nextReadAtMu      sync.RWMutex
	inflightReadAt    map[uint16]chan *ComPojtingerFelicitasR3MapV1ReadAtReply
	inflightReadAtMu  sync.RWMutex
	nextSize          uint16
	nextSizeMu        sync.RWMutex
	inflightSize      map[uint16]chan *ComPojtingerFelicitasR3MapV1SizeReply
	inflightSizeMu    sync.RWMutex
	nextTrack         uint16
	nextTrackMu       sync.RWMutex
	inflightTrack     map[uint16]chan *ComPojtingerFelicitasR3MapV1TrackReply
	inflightTrackMu   sync.RWMutex
	nextSync          uint16
	nextSyncMu        sync.RWMutex
	inflightSync      map[uint16]chan *ComPojtingerFelicitasR3MapV1SyncReply
	inflightSyncMu    sync.RWMutex
	nextClose         uint16
	nextCloseMu       sync.RWMutex
	inflightClose     map[uint16]chan *ComPojtingerFelicitasR3MapV1CloseReply
	inflightCloseMu   sync.RWMutex
	nextStreamingID   uint16
	nextStreamingIDMu sync.RWMutex
}
type Client struct {
	*frisbee.Client
	Seeder *subSeederClient
}

func NewClient(tlsConfig *tls.Config, logger *zerolog.Logger) (*Client, error) {
	c := new(Client)
	table := make(frisbee.HandlerTable)

	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Seeder.inflightReadAtMu.RLock()
		if ch, ok := c.Seeder.inflightReadAt[incoming.Metadata.Id]; ok {
			c.Seeder.inflightReadAtMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapV1ReadAtReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Seeder.inflightReadAtMu.RUnlock()
		}
		return
	}
	table[11] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Seeder.inflightSizeMu.RLock()
		if ch, ok := c.Seeder.inflightSize[incoming.Metadata.Id]; ok {
			c.Seeder.inflightSizeMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapV1SizeReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Seeder.inflightSizeMu.RUnlock()
		}
		return
	}
	table[12] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Seeder.inflightTrackMu.RLock()
		if ch, ok := c.Seeder.inflightTrack[incoming.Metadata.Id]; ok {
			c.Seeder.inflightTrackMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapV1TrackReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Seeder.inflightTrackMu.RUnlock()
		}
		return
	}
	table[13] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Seeder.inflightSyncMu.RLock()
		if ch, ok := c.Seeder.inflightSync[incoming.Metadata.Id]; ok {
			c.Seeder.inflightSyncMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapV1SyncReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Seeder.inflightSyncMu.RUnlock()
		}
		return
	}
	table[14] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.Seeder.inflightCloseMu.RLock()
		if ch, ok := c.Seeder.inflightClose[incoming.Metadata.Id]; ok {
			c.Seeder.inflightCloseMu.RUnlock()
			res := NewComPojtingerFelicitasR3MapV1CloseReply()
			res.Decode((*incoming.Content)[:incoming.Metadata.ContentLength])
			ch <- res
		} else {
			c.Seeder.inflightCloseMu.RUnlock()
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

	c.Seeder = new(subSeederClient)
	c.Seeder.client = c.Client
	c.Seeder.nextReadAtMu.Lock()
	c.Seeder.nextReadAt = 0
	c.Seeder.nextReadAtMu.Unlock()
	c.Seeder.inflightReadAt = make(map[uint16]chan *ComPojtingerFelicitasR3MapV1ReadAtReply)
	c.Seeder.nextSizeMu.Lock()
	c.Seeder.nextSize = 0
	c.Seeder.nextSizeMu.Unlock()
	c.Seeder.inflightSize = make(map[uint16]chan *ComPojtingerFelicitasR3MapV1SizeReply)
	c.Seeder.nextTrackMu.Lock()
	c.Seeder.nextTrack = 0
	c.Seeder.nextTrackMu.Unlock()
	c.Seeder.inflightTrack = make(map[uint16]chan *ComPojtingerFelicitasR3MapV1TrackReply)
	c.Seeder.nextSyncMu.Lock()
	c.Seeder.nextSync = 0
	c.Seeder.nextSyncMu.Unlock()
	c.Seeder.inflightSync = make(map[uint16]chan *ComPojtingerFelicitasR3MapV1SyncReply)
	c.Seeder.nextCloseMu.Lock()
	c.Seeder.nextClose = 0
	c.Seeder.nextCloseMu.Unlock()
	c.Seeder.inflightClose = make(map[uint16]chan *ComPojtingerFelicitasR3MapV1CloseReply)
	return c, nil
}

func (c *Client) Connect(addr string, streamHandler ...frisbee.NewStreamHandler) error {
	return c.Client.Connect(addr, func(stream *frisbee.Stream) {})
}

func (c *Client) FromConn(conn net.Conn, streamHandler ...frisbee.NewStreamHandler) error {
	return c.Client.FromConn(conn, func(stream *frisbee.Stream) {})
}

func (c *subSeederClient) ReadAt(ctx context.Context, req *ComPojtingerFelicitasR3MapV1ReadAtArgs) (res *ComPojtingerFelicitasR3MapV1ReadAtReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapV1ReadAtReply, 1)
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

func (c *subSeederClient) Size(ctx context.Context, req *ComPojtingerFelicitasR3MapV1SizeArgs) (res *ComPojtingerFelicitasR3MapV1SizeReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapV1SizeReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 11

	c.nextSizeMu.Lock()
	c.nextSize += 1
	id := c.nextSize
	c.nextSizeMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightSizeMu.Lock()
	c.inflightSize[id] = ch
	c.inflightSizeMu.Unlock()
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
	c.inflightSizeMu.Lock()
	delete(c.inflightSize, id)
	c.inflightSizeMu.Unlock()
	packet.Put(p)
	return
}

func (c *subSeederClient) Track(ctx context.Context, req *ComPojtingerFelicitasR3MapV1TrackArgs) (res *ComPojtingerFelicitasR3MapV1TrackReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapV1TrackReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 12

	c.nextTrackMu.Lock()
	c.nextTrack += 1
	id := c.nextTrack
	c.nextTrackMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightTrackMu.Lock()
	c.inflightTrack[id] = ch
	c.inflightTrackMu.Unlock()
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
	c.inflightTrackMu.Lock()
	delete(c.inflightTrack, id)
	c.inflightTrackMu.Unlock()
	packet.Put(p)
	return
}

func (c *subSeederClient) Sync(ctx context.Context, req *ComPojtingerFelicitasR3MapV1SyncArgs) (res *ComPojtingerFelicitasR3MapV1SyncReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapV1SyncReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 13

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

func (c *subSeederClient) Close(ctx context.Context, req *ComPojtingerFelicitasR3MapV1CloseArgs) (res *ComPojtingerFelicitasR3MapV1CloseReply, err error) {
	ch := make(chan *ComPojtingerFelicitasR3MapV1CloseReply, 1)
	p := packet.Get()
	p.Metadata.Operation = 14

	c.nextCloseMu.Lock()
	c.nextClose += 1
	id := c.nextClose
	c.nextCloseMu.Unlock()
	p.Metadata.Id = id

	req.Encode(p.Content)
	p.Metadata.ContentLength = uint32(len(*p.Content))
	c.inflightCloseMu.Lock()
	c.inflightClose[id] = ch
	c.inflightCloseMu.Unlock()
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
	c.inflightCloseMu.Lock()
	delete(c.inflightClose, id)
	c.inflightCloseMu.Unlock()
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
