package common

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

const (
	Prefix                   = "internal.dispatcher"
	UnmarshalKey             = "dispatcher"
	UnmarshalGlobalConfigKey = "dispatcher.global"
	NoAuthGroupPath          = "/api/v1"
)

// ExtractRawBodyContext
func ExtractRawBodyContext(ctx echo.Context) []byte {
	if rawBody, ok := ctx.Get("rawBody").([]byte); ok {
		return rawBody
	}
	return nil
}

// ExtractBinderContext
func ExtractBinderContext(ctx echo.Context) echo.Binder {
	if binder, ok := ctx.Get("binder").(echo.Binder); ok {
		return binder
	}
	return nil
}

// SetRawBodyContext
func SetRawBodyContext(ctx echo.Context, rawBody []byte) {
	ctx.Set("rawBody", rawBody)
}

// SetBinder
func SetBinder(ctx echo.Context, binder echo.Binder) {
	ctx.Set("binder", binder)
}

// Groups
type Groups struct {
	Common *echo.Group
}

// Handler
type Handler interface {
	Route(groups *Groups)
}

// Validate
type Validator interface {
	Use(validator *validator.Validate)
}

// Services
type Services struct {
	Billing billing.BillingService
}

// Handlers
type Handlers []Handler

// HandlerSet
type HandlerSet struct {
	Services Services
	Validate *validator.Validate
	AwareSet provider.AwareSet
}

// BindAndValidate
func (h HandlerSet) BindAndValidate(req interface{}, ctx echo.Context) *echo.HTTPError {
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorRequestParamsIncorrect)
	}
	if err := h.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, GetValidationError(err))
	}
	return nil
}

// SrvCallHandler returns error if present, otherwise response as JSON with 200 OK
func (h HandlerSet) SrvCallHandler(req interface{}, err error, name, method string) *echo.HTTPError {
	h.AwareSet.L().Error(billing.ErrorGrpcServiceCallFailed,
		logger.PairArgs(
			ErrorFieldService, name,
			ErrorFieldMethod, method,
		),
		logger.WithPrettyFields(logger.Fields{"err": err, ErrorFieldRequest: req}),
	)
	return echo.NewHTTPError(http.StatusInternalServerError, ErrorInternal)
}
