package structs

type Config struct {
	Region     string
	Datacenter string
	Cores      int
	Mhz        int
	Mem        int
	Server     string
	StatePath  string
}

func DefaultConfig() *Config {
	return &Config{
		Region:     "global",
		Datacenter: "dc1",
		Cores:      2,
		Mhz:        1000,
		Mem:        1000,
		Server:     "127.0.0.1:4647",
		StatePath:  "state.json",
	}
}
