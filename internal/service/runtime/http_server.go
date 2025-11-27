package runtime

import (
	"github.com/labstack/echo/v4"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/ports/http"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/ports/http/server"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/config"
)

func NewHTTPServer(config config.Config, server *server.ODSGatewayServer) *echo.Echo {
	e := echo.New()

	http.RegisterHandlers(e, server)
	return e
}
