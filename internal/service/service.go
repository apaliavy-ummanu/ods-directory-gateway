package service

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/ports/http/server"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/config"
	odsAdapter "github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/adapters/ods-fhir"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/queries"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/runtime"
	odsHTTP "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

type Service struct {
	httpServer *http.Server
}

func NewService() (*Service, error) {
	appConfig, err := config.NewConfigFromEnv()
	if err != nil {
		log.Err(err).Msg("error creating config from env")
		return nil, err
	}

	odsAPIClient, err := odsHTTP.NewClientWithResponses(appConfig.ODSConfig.ServerURL)
	if err != nil {
		log.Err(err).Msg("error creating ODS API client")
		return nil, err
	}

	odsAPIAdapter := odsAdapter.NewClient(odsAPIClient)

	odsGatewayServer, err := server.NewODSGateway(app.ODSGatewayApp{
		Queries: app.Queries{
			GetOrganisationByODSCode: queries.NewGetOrganisationByODSCodeQueryHandler(odsAPIAdapter),
			SearchOrganisations:      queries.NewSearchOrganisationsQueryHandler(odsAPIAdapter),
		},
	})
	if err != nil {
		log.Err(err).Msg("error creating ODS Gateway server")
		return nil, err
	}

	handler := runtime.NewHTTPServer(appConfig, odsGatewayServer)

	return &Service{
		httpServer: &http.Server{
			Addr:              ":" + appConfig.HTTPPort,
			Handler:           handler,
			ReadHeaderTimeout: appConfig.RequestTimeout,
		},
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	// cancel on SIGINT/SIGTERM OR when parent ctx is canceled.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)

	go func() {
		log.Info().Msg("listening...")
		err := s.httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()

	// wait for either:
	//  1) ctx cancellation (signal or parent ctx) => attempt graceful shutdown
	//  2) server fails early/runtime => return that error
	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown requested")

		// give graceful shutdown its own deadline.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			_ = s.httpServer.Close()
			log.Err(err).Msg("graceful shutdown failed (forced close issued)")
			<-serveErr
			return err
		}

		if err := <-serveErr; err != nil {
			return err
		}

		log.Info().Msg("server stopped")
		return nil

	case err := <-serveErr:
		return err
	}
}
