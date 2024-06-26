package httphandler

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	prettylogger "github.com/rdbell/echo-pretty-logger"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func NewRouter(timeout time.Duration, port int) *echo.Echo {
	e := echo.New()

	e.Use(prettylogger.Logger)
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			return true, nil
		},
		AllowMethods:     []string{"GET", "OPTIONS", "POST"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	e.Validator = &customValidator{validator: validator.New()}

	e.GET("/api/healthcheck", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})

	return e
}
