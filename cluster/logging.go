package cluster

import (
	"github.com/jtejido/hermes/config"
	"log"
	"os"
)

var (
	logger Logging
	conf   *config.Config
)

type Logging interface {
	Printf(format string, args ...interface{})
}

type clusterLogger struct {
	logger *log.Logger
}

func init() {
	logger, _ = NewLogger("")
}

func NewLogger(path string) (Logging, error) {
	var lg *log.Logger
	if path == "" {
		lg = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		lg = log.New(f, "", log.LstdFlags)
	}

	logger = &clusterLogger{
		logger: lg,
	}

	return logger, nil
}

func (c *clusterLogger) Printf(format string, args ...interface{}) {
	c.logger.Printf(format, args...)
}
