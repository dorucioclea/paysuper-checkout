package common

import (
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/labstack/echo/v4"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

// CheckProjectAuthRequestSignature
func CheckProjectAuthRequestSignature(dispatch HandlerSet, ctx echo.Context, projectId string) error {
	signature := ctx.Request().Header.Get(HeaderXApiSignatureHeader)

	if signature == "" {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessageSignatureHeaderIsEmpty)
	}

	req := &billing.CheckProjectRequestSignatureRequest{
		Body:      string(ExtractRawBodyContext(ctx)),
		ProjectId: projectId,
		Signature: signature,
	}
	rsp, err := dispatch.Services.Billing.CheckProjectRequestSignature(ctx.Request().Context(), req)

	if err != nil {
		dispatch.AwareSet.L().Error(InternalErrorTemplate, logger.Args("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrorUnknown)
	}

	if rsp.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return nil
}

// GetValidationError
func GetValidationError(err error) (rspErr billing.ResponseErrorMessage) {
	vErr := err.(validator.ValidationErrors)[0] // TODO: possible out of range
	val, ok := ValidationErrors[vErr.Field()]

	if ok {
		rspErr = val
	} else {
		if vErr.Tag() == RequestParameterZipUsa {
			rspErr = ErrorMessageIncorrectZip
		} else {
			rspErr = ErrorValidationFailed
		}
	}

	rspErr.Details = fmt.Sprintf(ErrorMessageMask, vErr.Field(), vErr.Tag())
	return rspErr
}
