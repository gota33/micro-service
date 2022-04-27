package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gota33/errors"
	"github.com/gota33/initializr"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"server/internal/service/auth"
	"server/internal/service/entity"
)

const (
	endpointHealth  = "healthz"
	endpointMetrics = "metrics"
	timeout         = 10 * time.Second
)

type Config struct {
	Addr string
	RDS  *sql.DB
}

func Run(ctx context.Context, c Config) (err error) {
	srv := fiber.New(fiber.Config{
		IdleTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		ErrorHandler: handleError,
	})

	srv.Use(logger.New())
	srv.Use(initUserContext)
	srv.Use(initAuthContext)

	srv.Get(endpointHealth, health())
	srv.Get(endpointMetrics, metrics())

	r := router{Router: srv, config: c}
	r.setup()

	listen := func() error {
		return srv.Listen(c.Addr)
	}

	shutdown := func() {
		if shutdownErr := srv.Shutdown(); shutdownErr != nil {
			logrus.WithError(shutdownErr).Warn("Shutdown server error")
		}
	}

	return initializr.Run(ctx, listen, shutdown)
}

func initUserContext(c *fiber.Ctx) (err error) {
	c.SetUserContext(c.Context())
	return c.Next()
}

func initAuthContext(c *fiber.Ctx) (err error) {
	if token := c.Get(fiber.HeaderAuthorization); token != "" {
		var user auth.User
		if err = user.FromJWT(token); err != nil {
			return errors.Annotate(err, errors.Unauthenticated)
		}
		c.SetUserContext(user.WithContext(c.UserContext()))
	}
	return c.Next()
}

func health() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}
}

func metrics() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

func handleError(c *fiber.Ctx, cause error) error {
	var (
		err          error
		fiberErr     *fiber.Error
		validateErrs validator.ValidationErrors
	)
	switch {
	case errors.As(cause, &fiberErr):
		var status errors.StatusCode
		switch fiberErr.Code {
		case http.StatusBadRequest:
			status = errors.InvalidArgument
		case http.StatusNotFound:
			status = errors.NotFound
		// TODO: Handle more codes...
		default:
			status = errors.Unknown
		}
		err = errors.Annotate(fiberErr, status)
	case errors.As(cause, &validateErrs):
		details := errors.BadRequest{
			FieldViolations: make([]errors.FieldViolation, len(validateErrs)),
		}
		for i, subErr := range validateErrs {
			details.FieldViolations[i] = errors.FieldViolation{
				Field:       subErr.Field(),
				Description: subErr.Error(),
			}
		}
		err = errors.WithBadRequest(cause, details)
	default:
		err = cause
	}

	buf := &bytes.Buffer{}
	enc := errors.NewEncoder(json.NewEncoder(buf))
	if encErr := enc.Encode(err); encErr != nil {
		return fiber.DefaultErrorHandler(c, encErr)
	}

	return c.
		Status(errors.Code(err).Http()).
		JSON(json.RawMessage(buf.Bytes()))
}

func handler[Request any, Response any](h func(context.Context, Request) (Response, error)) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			req Request
			res Response
		)
		if err = c.BodyParser(&req); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return
		}
		if err = c.QueryParser(&req); err != nil {
			return
		}
		if err = paramParser(c, &req); err != nil {
			return
		}
		// TODO: More parsers here...
		if err = doValidate(c, req); err != nil {
			return
		}
		if res, err = h(c.UserContext(), req); err != nil {
			return
		}

		var temp any = res
		switch v := temp.(type) {
		case int:
			return c.SendStatus(v)
		default:
			return c.JSON(res)
		}
	}
}

func doValidate(c *fiber.Ctx, req any) (err error) {
	if c.Method() != http.MethodPatch {
		return entity.Validate.Struct(req)
	}
	return
}

func paramParser(c *fiber.Ctx, req any) (err error) {
	const tagName = "param"

	keys := make([]string, 0)
	t := reflect.TypeOf(req).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if v := f.Tag.Get(tagName); v != "" {
			keys = append(keys, v)
		}
	}

	if len(keys) == 0 {
		return
	}

	var (
		decoder *mapstructure.Decoder
		config  = &mapstructure.DecoderConfig{
			TagName: tagName,
			Result:  req,
		}
	)
	if decoder, err = mapstructure.NewDecoder(config); err != nil {
		return
	}

	params := make(map[string]any, len(keys))
	for _, key := range keys {
		params[key] = c.Params(key)
	}
	return decoder.Decode(params)
}
