package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/elog"
)

func main() {
	ctx := context.Background()
	elog.Init()

	svc, err := service.NewService()
	if err != nil {
		log.Panic().Err(err).Msg("could not create service instance")
	}

	err = svc.Start(ctx)
	if err != nil {
		log.Panic().Err(err).Msg("could not start service")
	}
}
