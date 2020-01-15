package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"net/http"
)

const (
	paymentPath = "/payment"
)

type PaymentRoute struct {
	dispatch common.HandlerSet
	cfg      *common.Config
	provider.LMT
}

type RedirectResponse struct {
	// The redirection URL.
	RedirectUrl string `json:"redirect_url"`
	// Has a true value if it's need to perform redirection by link.
	NeedRedirect bool `json:"need_redirect"`
}

func NewPaymentRoute(set common.HandlerSet, cfg *common.Config) *PaymentRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PaymentRoute"})
	return &PaymentRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

func (h *PaymentRoute) Route(groups *common.Groups) {
	groups.Common.POST(paymentPath, h.processCreatePayment)
}

// @summary Get a redirect URL
// @desc Get a redirect URL after a processed payment
// @id paymentPathProcessCreatePayment
// @tag Payment
// @accept application/json
// @produce application/json
// @body grpc.PaymentCreateRequest
// @success 200 {object} RedirectResponse OK
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @router /api/v1/payment [post]
func (h *PaymentRoute) processCreatePayment(ctx echo.Context) error {
	data := make(map[string]string)
	err := (&common.PaymentCreateProcessBinder{}).Bind(data, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req := &grpc.PaymentCreateRequest{
		Data:           data,
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}
	res, err := h.dispatch.Services.Billing.PaymentCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentCreateProcess")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	body := map[string]interface{}{
		"redirect_url":  res.RedirectUrl,
		"need_redirect": res.NeedRedirect,
	}

	return ctx.JSON(http.StatusOK, body)
}
