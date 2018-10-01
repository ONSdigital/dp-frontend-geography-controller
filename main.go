package main

import (
	"context"
	"net/smtp"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-geography-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

type unencryptedAuth struct {
	smtp.Auth
}

func main() {
	cfg := config.Get()

	log.Namespace = "dp-frontend-geography-controller"

	cli := codelist.New(cfg.CodeListAPIURL)
	dcli := dataset.NewAPIClient(cfg.DatasetAPIURL, "", "")

	router := mux.NewRouter()

	rend := renderer.New(cfg.RendererURL)

	router.StrictSlash(true).Path("/healthcheck").HandlerFunc(healthcheck.Handler)
	router.StrictSlash(true).Path("/geography").Methods("GET").HandlerFunc(handlers.HomepageRender(rend, cli))
	router.StrictSlash(true).Path("/geography/{codeListID}").Methods("GET").HandlerFunc(handlers.ListPageRender(rend, cli))
	router.StrictSlash(true).Path("/geography/{codeListID}/{codeID}").Methods("GET").HandlerFunc(handlers.AreaPageRender(rend, cli, dcli))

	log.Info("Starting server", log.Data{
		"bind_addr":        cfg.BindAddr,
		"renderer_url":     cfg.RendererURL,
		"codelist_api_url": cfg.CodeListAPIURL,
	})

	s := server.New(cfg.BindAddr, router)
	s.HandleOSSignals = false

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Error(err, nil)
			os.Exit(2)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.InfoCtx(ctx, "shutting service down gracefully", nil)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.ErrorCtx(ctx, err, nil)
	}
}
