package main

import (
	"context"
	"net/smtp"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-geography-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

func main() {
	cfg := config.Get()

	log.Namespace = "dp-frontend-geography-controller"

	router := mux.NewRouter()

	rend := renderer.New(cfg.RendererURL)

	router.StrictSlash(true).Path("/geography/").Methods("GET").HandlerFunc(handlers.GeographyRender(rend))

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
