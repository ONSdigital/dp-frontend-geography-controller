package main

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-api-clients-go/codelist"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-geography-controller/handlers"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type unencryptedAuth struct {
	smtp.Auth
}

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
	log.Namespace = "dp-frontend-geography-controller"

	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting configuration", log.Error(err))
		os.Exit(1)
	}
	log.Event(ctx, "config on startup", log.Data{"config": cfg})

	versionInfo, err := health.NewVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.Error(err))
		os.Exit(1)
	}

	cc := codelist.New(cfg.CodeListAPIURL)
	dc := dataset.NewAPIClient(cfg.DatasetAPIURL)
	rend := renderer.New(cfg.RendererURL)

	hc := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	if err = registerCheckers(ctx, &hc, cc, dc, rend); err != nil {
		os.Exit(1)
	}

	router := mux.NewRouter()

	router.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	router.StrictSlash(true).Path("/geography").Methods("GET").HandlerFunc(handlers.HomepageRender(rend, cc, cfg.EnableLoop11))
	router.StrictSlash(true).Path("/geography/{codeListID}").Methods("GET").HandlerFunc(handlers.ListPageRender(rend, cc, cfg.EnableLoop11))
	router.StrictSlash(true).Path("/geography/{codeListID}/{codeID}").Methods("GET").HandlerFunc(handlers.AreaPageRender(rend, cc, dc))

	s := server.New(cfg.BindAddr, router)
	s.HandleOSSignals = false

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Event(ctx, "error starting http server", log.Error(err))
			os.Exit(2)
		}
	}()

	hc.Start(ctx)

	// Block until a fatal error occurs
	select {
	case signal := <-signals:
		log.Event(ctx, "quitting after os signal received", log.Data{"signal": signal})
	}

	log.Event(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout))

	// give the app `Timeout` seconds to close gracefully before killing it.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	go func() {
		log.Event(ctx, "stop health checkers")
		hc.Stop()

		if err := s.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to gracefully shutdown http server", log.Error(err))
		}

		cancel() // stop timer
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "context deadline exceeded", log.Error(ctx.Err()))
	} else {
		log.Event(ctx, "graceful shutdown complete", log.Data{"context": ctx.Err()})
	}

	os.Exit(0)
}

func registerCheckers(ctx context.Context, h *health.HealthCheck, c *codelist.Client, d *dataset.Client, r *renderer.Renderer) (err error) {
	if err = h.AddCheck("codelist API", c.Checker); err != nil {
		log.Event(ctx, "failed to add codelist API checker", log.Error(err))
	}

	if err = h.AddCheck("dataset API", d.Checker); err != nil {
		log.Event(ctx, "failed to add dataset API checker", log.Error(err))
	}

	if err = h.AddCheck("frontend renderer", r.Checker); err != nil {
		log.Event(ctx, "failed to add frontend renderer checker", log.Error(err))
	}

	return
}
