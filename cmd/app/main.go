package main

import (
	"context"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service"
)

func main() {
	ctx := context.Background()

	svc, err := service.NewService()
	if err != nil {
		panic(err)
	}

	err = svc.Start(ctx)
	if err != nil {
		panic(err)
	}
}
