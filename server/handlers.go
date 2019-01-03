package main

import (
	"encoding/json"
	"github.com/jtejido/hermes/hermes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	Name                 = "hermes"
	Description          = "Hermes Cache API Service"
	ApiBasePath          = "/" + Name + "/api/"
	CachePath            = ApiBasePath + "cache/"
	StatsCachePath       = ApiBasePath + "stats"
	ClearCachePath       = ApiBasePath + "clear"
	FilterClearCachePath = ApiBasePath + "filterClear"
	Version              = "1.0.0"
)

// pretty basic, we can include status codes or something too.
type Response struct {
	Value []byte
}

func cacheIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCacheHandler(w, r)
		case http.MethodPut:
			putCacheHandler(w, r)
		case http.MethodDelete:
			deleteCacheHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func statsIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCacheStatsHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func clearIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getClearHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func clearFilterIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getFilterClearHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func getFilterClearHandler(w http.ResponseWriter, r *http.Request) {
	cache.ResetFilter()
	w.WriteHeader(http.StatusOK)
	return
}

func getClearHandler(w http.ResponseWriter, r *http.Request) {
	cache.Clear()
	w.WriteHeader(http.StatusOK)
	return
}

func getCacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	target, err := json.Marshal(cache.GetStats())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(target)
	return
}

func getCacheHandler(w http.ResponseWriter, r *http.Request) {
	var ctx hermes.Context
	target := r.URL.Path[len(CachePath):]
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("can't get a key if there is no key."))
		log.Print("empty request.")
		return
	}
	entry, err := cache.Get(ctx, target)
	if err != nil {
		errMsg := (err).Error()
		if strings.Contains(errMsg, "not found") {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(&Response{Value: entry})
	if err != nil {
		log.Print(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func putCacheHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len(CachePath):]
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("can't put a key if there is no key."))
		log.Print("empty request.")
		return
	}
	var ctx hermes.Context
	entry, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := cache.Set(ctx, target, entry); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("stored \"%s\" in cache.", target)
	w.WriteHeader(http.StatusCreated)
}

func deleteCacheHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len(CachePath):]
	var ctx hermes.Context
	if err := cache.Delete(ctx, target); err != nil {
		if strings.Contains((err).Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("%s not found.", target)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("internal cache error: %s", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
