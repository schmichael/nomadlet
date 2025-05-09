package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	client "github.com/schmichael/nomadlet/client"
	"github.com/schmichael/nomadlet/internal/structs"
	"github.com/schmichael/nomadlet/version"
)

func main() {
	config := structs.DefaultConfig()

	flag.IntVar(&config.Cores, "cores", config.Cores, "number of cores")
	flag.IntVar(&config.Mhz, "mhz", config.Mhz, "total mhz available")
	flag.IntVar(&config.Mem, "mem", config.Mem, "total memory in MB")
	flag.StringVar(&config.Region, "region", config.Region, "region")
	flag.StringVar(&config.Datacenter, "dc", config.Datacenter, "datacenter")
	flag.StringVar(&config.Server, "server", config.Server, "server address")
	flag.StringVar(&config.StatePath, "state", config.StatePath, "state file path")
	flag.StringVar(&config.Name, "name", config.Name, "node name")
	//TODO tls stuff
	//TODO multi-server handling

	versionFlag := false
	flag.BoolVar(&versionFlag, "version", versionFlag, "print version and exit")

	flag.Parse()

	if versionFlag {
		fmt.Println("nomadlet " + version.Version)
		os.Exit(0)
	}

	if config.Name == "" {
		fmt.Fprintf(os.Stderr, "must specify node name")
		os.Exit(1)
	}

	client, err := client.NewClient(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		client.Run(ctx)
	}()
	<-doneCh
}
