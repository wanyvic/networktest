package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
)

func main() {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	var configFilePath string
	flag.StringVar(&configFilePath, "c", "", "config file")
	flag.Parse()
	defer glog.Flush()
	config, err := LoadConfigFile(configFilePath)
	if err != nil {
		glog.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	for _, list := range config.AddrList {
		go RunPoolChecker(ctx, &wg, list.Pool, list.Mode)
	}

	// Block until a signal is received.
	<-c
	cancel()
	wg.Wait()
}
