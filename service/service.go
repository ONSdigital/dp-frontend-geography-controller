package service

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/codelist"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-geography-controller/handlers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the frontend homepage controller
type Service struct {
	Config             *config.Config
	routerHealthClient *health.Client
	HealthCheck        HealthChecker
	Server             HTTPServer
	CodelistClient     *codelist.Client
	DatasetClient      *dataset.Client
	RendererClient     *renderer.Renderer
	ServiceList        *ExternalServiceList
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	log.Event(ctx, "running service", log.INFO)

	// Initialise Service struct
	svc = &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}

	// Get health client for api router
	svc.routerHealthClient = serviceList.GetHealthClient("api-router", cfg.APIRouterURL)

	// Initialise clients
	svc.CodelistClient = codelist.NewWithHealthClient(svc.routerHealthClient)
	svc.DatasetClient = dataset.NewWithHealthClient(svc.routerHealthClient)
	svc.RendererClient = renderer.New(cfg.RendererURL)

	// Get healthcheck with checkers
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "failed to create health check", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := svc.registerCheckers(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	// Initialise router
	router := mux.NewRouter()
	router.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	router.StrictSlash(true).Path("/geography").Methods("GET").HandlerFunc(handlers.HomepageRender(svc.RendererClient, svc.CodelistClient))
	router.StrictSlash(true).Path("/geography/{codeListID}").Methods("GET").HandlerFunc(handlers.ListPageRender(svc.RendererClient, svc.CodelistClient))
	router.StrictSlash(true).Path("/geography/{codeListID}/{codeID}").Methods("GET").HandlerFunc(handlers.AreaPageRender(svc.RendererClient, svc.CodelistClient, svc.DatasetClient))

	svc.Server = serviceList.GetHTTPServer(cfg.BindAddr, router)

	// Start Healthcheck and HTTP Server
	svc.HealthCheck.Start(ctx)
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return svc, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context, cfg *config.Config) (err error) {

	hasErrors := false

	if err = svc.HealthCheck.AddCheck("API router", svc.routerHealthClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add api router checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("frontend renderer", svc.RendererClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add frontend renderer checker", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
