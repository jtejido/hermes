package hermes

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jtejido/hermes/consistenthash"
	pb "github.com/jtejido/hermes/hermespb"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type HTTPPool struct {
	Context     func(*http.Request) Context
	Transport   func(Context) http.RoundTripper
	self        string
	opts        HTTPPoolOptions
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
	mux         *http.ServeMux
}

type HTTPPoolOptions struct {
	BasePath string
	Replicas int
}

func NewHTTPPool(self string) *HTTPPool {
	p := NewHTTPPoolOpts(self, nil)
	http.Handle(p.opts.BasePath, p)
	return p
}

var httpPoolMade bool
var maxFileDescriptors = 100

func NewHTTPPoolOpts(self string, o *HTTPPoolOptions) *HTTPPool {
	if httpPoolMade {
		panic("NewHTTPPool must be called only once")
	}
	httpPoolMade = true

	p := &HTTPPool{
		self:        self,
		httpGetters: make(map[string]*httpGetter),
	}

	if o != nil {
		p.opts = *o
	}

	if p.opts.BasePath == "" {
		p.opts.BasePath = defaultBasePath
	}

	if p.opts.Replicas == 0 {
		p.opts.Replicas = defaultReplicas
	}

	p.mux = http.NewServeMux()
	p.peers = consistenthash.New(p.opts.Replicas)

	p.mux.Handle(defaultBasePath, loader(peerHandler()))

	RegisterPeerPicker(func() PeerPicker { return p })
	return p
}

func (p *HTTPPool) Set(peerList ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(p.opts.Replicas)
	p.peers.Add(peerList...)
	p.httpGetters = make(map[string]*httpGetter, len(peerList))
	for _, peer := range peerList {
		p.httpGetters[peer] = &httpGetter{transport: p.Transport, baseURL: peer + p.opts.BasePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (ProtoGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.peers.IsEmpty() {
		return nil, false
	}

	if peer := p.peers.Get(key); peer != p.self {
		return p.httpGetters[peer], true
	}

	return nil, false
}

func (p *HTTPPool) IncrementLoad() {
	p.peers.Increment(p.self)
}

func (p *HTTPPool) DecrementLoad() {
	p.peers.Decrement(p.self)
}

func (p *HTTPPool) GetLoad() uint64 {
	return p.peers.GetLoad(p.self)
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

type httpGetter struct {
	transport func(Context) http.RoundTripper
	baseURL   string
}

var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func (h *httpGetter) Get(context Context, in *pb.GetRequest, out *pb.GetResponse) error {
	u := fmt.Sprintf(
		"%v%v",
		h.baseURL,
		url.QueryEscape(in.GetKey()),
	)

	req, err := http.NewRequest("GET", u, nil)

	if err != nil {
		return err
	}
	tr := http.DefaultTransport
	if h.transport != nil {
		tr = h.transport(context)
	}
	res, err := tr.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}
	b := bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	defer bufferPool.Put(b)
	_, err = io.Copy(b, res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	err = proto.Unmarshal(b.Bytes(), out)
	if err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

func (h *httpGetter) Set(context Context, in *pb.SetRequest, out *pb.SetResponse) error {
	u := fmt.Sprintf(
		"%v%v",
		h.baseURL,
		url.QueryEscape(in.GetKey()),
	)

	req, err := http.NewRequest("PUT", u, bytes.NewBuffer(in.GetValue()))

	if err != nil {
		return err
	}
	tr := http.DefaultTransport
	if h.transport != nil {
		tr = h.transport(context)
	}
	res, err := tr.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b := bufferPool.Get().(*bytes.Buffer)
		b.Reset()
		defer bufferPool.Put(b)
		_, err = io.Copy(b, res.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %v", err)
		}
		err = proto.Unmarshal(b.Bytes(), out)
		if err != nil {
			return fmt.Errorf("decoding response body: %v", err)
		}

		return nil
	}

	return nil
}

func (h *httpGetter) Delete(context Context, in *pb.DeleteRequest, out *pb.DeleteResponse) error {
	u := fmt.Sprintf(
		"%v%v",
		h.baseURL,
		url.QueryEscape(in.GetKey()),
	)

	req, err := http.NewRequest("DELETE", u, nil)

	if err != nil {
		return err
	}
	tr := http.DefaultTransport
	if h.transport != nil {
		tr = h.transport(context)
	}
	res, err := tr.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b := bufferPool.Get().(*bytes.Buffer)
		b.Reset()
		defer bufferPool.Put(b)
		_, err = io.Copy(b, res.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %v", err)
		}
		err = proto.Unmarshal(b.Bytes(), out)
		if err != nil {
			return fmt.Errorf("decoding response body: %v", err)
		}

		return nil
	}

	return nil
}
