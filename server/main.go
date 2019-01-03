package main

import (
	"context"
	"flag"
	"github.com/jtejido/hermes/cluster"
	"github.com/jtejido/hermes/config"
	"github.com/jtejido/hermes/hermes"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	cache  *hermes.Cache
	ver    bool
	conf   *config.Config
	logger cluster.Logging
)

func init() {

	conf, _ = config.LoadConfig("")

	flag.IntVar(&conf.Cache.ShardCount, "shards", conf.Cache.ShardCount, "Number of shards for the cache.")
	flag.IntVar(&conf.Http.Host, "host", conf.Http.Host, "The front-end local port.")
	flag.IntVar(&conf.Peers.Listen, "listen", conf.Peers.Listen, "The port for peers to listen on.")
	flag.IntVar(&conf.Cache.Size, "maxmemory", conf.Cache.Size, "Maximum amount of data in the cache in MB.")
	flag.Float64Var(&conf.Cache.Lambda, "lambda", conf.Cache.Lambda, "Lambda used for LRFU.")
	flag.BoolVar(&conf.Filter.Enabled, "filter", conf.Filter.Enabled, "Bloom Filter enabled?")
	flag.UintVar(&conf.Filter.FilterItemCount, "filter-items", conf.Filter.FilterItemCount, "Maximum number of items to be stored in filter.")
	flag.StringVar(&conf.Http.AccessLog, "logfile", conf.Http.AccessLog, "Location of the logfile.")
	flag.BoolVar(&ver, "version", false, "Hermes version.")
}

func main() {

	flag.Parse()

	logger, _ = cluster.NewLogger(conf.Http.AccessLog)

	logger.Printf("cache initialised.")

	cachePeers, _ := cluster.New(conf)

	frontend := ":" + strconv.Itoa(conf.Http.Host)

	cache = hermes.NewCache(conf)

	peers := &http.Server{Addr: cachePeers.ListenOn(), Handler: cachePeers}

	s := NewServer(func(s *server) {
		s.logger = logger
	})

	server := &http.Server{Addr: frontend, Handler: s}

	logger.Printf("starting peer listening on " + frontend)

	go func() {
		if err := peers.ListenAndServe(); err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	logger.Printf("starting http listening on " + frontend)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	select {}

	graceful(server, peers, 5*time.Second)

}

func graceful(hs *http.Server, hs_p *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Printf("shutdown with timeout: %s", timeout)

	if err := hs.Shutdown(ctx); err != nil {
		logger.Printf("server Error: %v\n", err)
	} else {
		logger.Printf("http listening stopped. \n")
	}

	if err := hs_p.Shutdown(ctx); err != nil {
		logger.Printf("listener Error: %v\n", err)
	} else {
		logger.Printf("peer listening stopped. \n")
	}
}
