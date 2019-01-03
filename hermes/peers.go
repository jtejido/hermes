package hermes

import (
	pb "github.com/jtejido/hermes/hermespb"
)

var (
	portPicker func() PeerPicker
)

type Context interface{}

type ProtoGetter interface {
	Get(context Context, in *pb.GetRequest, out *pb.GetResponse) error
	Set(context Context, in *pb.SetRequest, out *pb.SetResponse) error
	Delete(context Context, in *pb.DeleteRequest, out *pb.DeleteResponse) error
}

type PeerPicker interface {
	PickPeer(key string) (peer ProtoGetter, ok bool)
	IncrementLoad()
	DecrementLoad()
	GetLoad() uint64
}

type NoPeers struct{}

func (NoPeers) PickPeer(key string) (peer ProtoGetter, ok bool) { return }
func (NoPeers) IncrementLoad()                                  {}
func (NoPeers) DecrementLoad()                                  {}
func (NoPeers) GetLoad() (value uint64)                         { return }

func RegisterPeerPicker(fn func() PeerPicker) {
	if portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	portPicker = func() PeerPicker { return fn() }
}

func getPeers() PeerPicker {
	if portPicker == nil {
		return NoPeers{}
	}
	pk := portPicker()
	if pk == nil {
		pk = NoPeers{}
	}
	return pk
}
