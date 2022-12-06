package main

import (
	"os"

	"github.com/estrys/estrys/cmd"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/worker"
)

func main() {
	globalContext, _, err := cmd.Bootstrap()
	log := dic.GetService[logger.Logger]()
	if err != nil {
		if log != nil {
			log.WithError(err).Error("unable to start application")
			os.Exit(1)
		}
		panic(err)
	}

	err = worker.StartBroker(globalContext)
	if err != nil {
		log.WithError(err).Error("worker failed")
		os.Exit(1)
	}
	log.Info("worker stopped")
	os.Exit(0)
}
