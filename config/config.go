package config

import "github.com/BurntSushi/toml"
import "fmt"

// This is the base Config type for hermes. Extend as needed.
type Config struct {
	Title  string
	Cache  CacheConfig  `toml:"cache"`
	Filter FilterConfig `toml:"filter"`
	Http   HermesHTTP   `toml:"http"`
	Peers  PeerConfig   `toml:"peers"`
}

type CacheConfig struct {
	ShardCount int `toml:"shards"`
	Size       int
	Lambda     float64
}

type FilterConfig struct {
	FilterItemCount uint `toml:"default_filter_count"`
	Enabled         bool
}

type HermesHTTP struct {
	Host      int
	AccessLog string `toml:"access_log_location"`
}

type PeerConfig struct {
	Listen                    int
	PeerListenPort            int `toml:"peer_listen_port"`
	Heartbeat                 int
	MulticastEnabled          bool   `toml:"multicast_enabled"`
	MulticastClusterName      string `toml:"multicast_cluster_name"`
	MulticastAddress          string `toml:"multicast_address"`
	MulticastPort             int    `toml:"multicast_port"`
	MulticastAnnounceInterval int    `toml:"multicast_announce_interval"`
	Nodes                     []string
}

func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = "config.toml"
	}

	var c Config

	if _, err := toml.DecodeFile(filename, &c); err != nil {
		return nil, err
	}

	fmt.Printf("Successfully decoded config file: %s \n", c.Title)

	return &c, nil
}
