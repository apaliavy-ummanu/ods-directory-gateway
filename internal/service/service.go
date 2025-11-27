package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

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
		return nil, err
	}

	odsAPIClient, err := odsHTTP.NewClientWithResponses(appConfig.OdsFhirAPIServerURL)
	if err != nil {
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
		return nil, err
	}

	handler := runtime.NewHTTPServer(appConfig, odsGatewayServer)

	return &Service{
		httpServer: &http.Server{
			Addr:              ":" + appConfig.HTTPPort,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// todo pass ctx properly
func (s *Service) Start(ctx context.Context) error {
	go func() {
		log.Printf("listening...")
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// wait for SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Shutting down...")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(timeoutCtx); err != nil {
		return err
	}

	log.Println("Server stopped.")

	return nil
}
