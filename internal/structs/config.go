package structs

import "os"

type Config struct {
	Region     string
	Datacenter string
	Cores      int
	Mhz        int
	Mem        int
	Name       string
	Server     string
	StatePath  string
}

func DefaultConfig() *Config {
	n, _ := os.Hostname()
	return &Config{
		Region:     "global",
		Datacenter: "dc1",
		Cores:      2,
		Mhz:        1000,
		Mem:        1000,
		Name:       n,
		Server:     "127.0.0.1:4647",
		StatePath:  "state.json",
	}
}
