package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-geography-controller/service"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

const serviceName = "dp-frontend-geography-controller"

// App version informaton retrieved on runtime
var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "unable to run application", log.Error(err), log.FATAL)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create service initialiser and an error channel for fatal errors
	svcErrors := make(chan error, 1)
	svcList := service.NewServiceList(&service.Init{})

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting configuration", log.FATAL, log.Error(err))
		return err
	}
	log.Event(ctx, "config on startup", log.INFO, log.Data{"config": cfg})

	// Start service
	svc, err := service.Run(ctx, cfg, svcList, BuildTime, GitCommit, Version, svcErrors)
	if err != nil {
		return errors.Wrap(err, "running service failed")
	}

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Event(ctx, "service error received", log.ERROR, log.Error(err))
	case sig := <-signals:
		log.Event(ctx, "os signal received", log.Data{"signal": sig}, log.INFO)
	}
	return svc.Close(ctx)
}
