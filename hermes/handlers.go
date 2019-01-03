package hermes

import (
	"github.com/golang/protobuf/proto"
	pb "github.com/jtejido/hermes/hermespb"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	Value string
}

func peerHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPeerHandler(w, r)
		case http.MethodPut:
			putPeerHandler(w, r)
		case http.MethodDelete:
			deletePeerHandler(w, r)
		}
	})
}

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, defaultBasePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	key := r.URL.Path[len(defaultBasePath):]

	if key == "" {
		http.Error(w, "Empty key.", http.StatusBadRequest)
		return
	}

	var ctx Context

	value, err := cache.Get(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.GetResponse{Value: value})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(body)
}

func putPeerHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len(defaultBasePath):]

	if target == "" {
		http.Error(w, "Empty request.", http.StatusBadRequest)
		return
	}

	var ctx Context

	entry, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cache.Set(ctx, target, entry); err != nil {
		body, err_b := proto.Marshal(&pb.SetResponse{
			Error: &pb.Error{
				Message: err.Error(),
				Code:    0,
			},
		})

		if err_b != nil {
			http.Error(w, err_b.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len(defaultBasePath):]

	if target == "" {
		http.Error(w, "Empty key.", http.StatusBadRequest)
		return
	}

	var ctx Context
	if err := cache.Delete(ctx, target); err != nil {
		body, err_b := proto.Marshal(&pb.DeleteResponse{
			Error: &pb.Error{
				Message: err.Error(),
				Code:    0,
			},
		})

		if err_b != nil {
			http.Error(w, err_b.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
