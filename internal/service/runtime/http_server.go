package runtime

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	svcHTTP "github.com/Cleo-Systems/ods-fhir-gateway/internal/service/common/ports/http"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/common/ports/http/server"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/config"
)

func NewHTTPServer(config config.Config, server *server.ODSGatewayServer) *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2M"))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: getOrigins(config.Account),
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "X-API-Key", "x-correlation-id"},
	}))

	e.GET("/liveness", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	e.GET("/readiness", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	api := e.Group("")
	api.Use(APIKeyMiddleware(APIKeyConfig{
		Header:   "X-API-Key",
		Expected: config.APIKey,
		Skips:    []string{"/liveness", "/readiness"},
	}))

	e.Server.ReadTimeout = config.RequestTimeout
	e.Server.WriteTimeout = config.RequestTimeout
	e.Server.IdleTimeout = config.RequestTimeout

	svcHTTP.RegisterHandlers(api, server)

	return e
}

type APIKeyConfig struct {
	Header   string
	Expected string
	Skips    []string
	Allowed  []string
}

func APIKeyMiddleware(c APIKeyConfig) echo.MiddlewareFunc {
	header := c.Header
	if header == "" {
		header = "X-API-Key"
	}

	allowed := make([]string, 0, 1+len(c.Allowed))
	if c.Expected != "" {
		allowed = append(allowed, c.Expected)
	}
	allowed = append(allowed, c.Allowed...)

	skips := make(map[string]struct{}, len(c.Skips))
	for _, p := range c.Skips {
		skips[p] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if _, ok := skips[ctx.Request().URL.Path]; ok {
				return next(ctx)
			}

			got := strings.TrimSpace(ctx.Request().Header.Get(header))
			if got == "" || len(allowed) == 0 {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing api key")
			}

			for _, want := range allowed {
				if want == "" {
					continue
				}
				if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
					return next(ctx)
				}
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
		}
	}
}

func getOrigins(account string) []string {
	origins := []string{
		"http://localhost:*",
	}
	if account == "dev" {
		origins = append(origins, []string{
			"https://*.elevate-dev.cleosystems.com",
			"https://*.elevate-dev.cleosystems.com:*",
			"https://integration.elevate-dev.cleosystems.com",
		}...)
	}
	if account == "staging" {
		origins = append(origins, []string{
			"https://*.elevate-stg.cleosystems.com",
			"https://*.elevate-stg.cleosystems.com:*",
			"https://integration.elevate-stg.cleosystems.com",
		}...)
	}
	if account == "production" {
		origins = append(origins, []string{
			"https://*.elevate.cleosystems.com",
			"https://*.elevate.cleosystems.com:*",
			"https://integration.elevate.cleosystems.com",
		}...)
	}
	return origins
}
