package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"log"

	"github.com/t4ke0/gbs"
)

func main() {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	cancel := make(chan struct{})

	go func() {
		<-quit
		cancel <- struct{}{}
	}()

	if err := gbs.LiveBuild("../", func(...gbs.BuildFuncOpt) error {
		fmt.Println("build ....")
		return nil
	}, cancel); err != nil {
		log.Fatal(err)
	}
}
