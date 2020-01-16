package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
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
	// Has a true value if it needs to redirect by a link.
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

// @summary Create a payment and return a redirect URL
// @desc Create a payment using the order data and return the redirect URL to the payment authorisation
// @id paymentPathProcessCreatePayment
// @tag Payment
// @accept application/json
// @produce application/json
// @body grpc.PaymentCreateRequest
// @success 200 {object} RedirectResponse Returns a redirect URL and a boolean value whether a redirect needs
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @router /api/v1/payment [post]
func (h *PaymentRoute) processCreatePayment(ctx echo.Context) error {
	data := make(map[string]string)
	err := (&common.PaymentCreateProcessBinder{}).Bind(data, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req := &billing.PaymentCreateRequest{
		Data:           data,
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}
	res, err := h.dispatch.Services.Billing.PaymentCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "PaymentCreateProcess")
	}

	if res.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	body := map[string]interface{}{
		"redirect_url":  res.RedirectUrl,
		"need_redirect": res.NeedRedirect,
	}

	return ctx.JSON(http.StatusOK, body)
}
