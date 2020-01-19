package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	_ "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"net/http"
)

const (
	paymentCountriesOrderIdPath = "/payment_countries/:order_id"
)

type CountryRoute struct {
	dispatch common.HandlerSet
	cfg      *common.Config
	provider.LMT
}

func NewCountryRoute(set common.HandlerSet, cfg *common.Config) *CountryRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "CountryRoute"})
	return &CountryRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

func (h *CountryRoute) Route(groups *common.Groups) {
	groups.Common.GET(paymentCountriesOrderIdPath, h.getPaymentCountries)
}

// @summary Get information about countries
// @desc Get a list of available countries for this order
// @id paymentCountriesOrderIdPathGetPaymentCountries
// @tag Country
// @accept application/json
// @produce application/json
// @success 200 {object} billing.CountriesList Returns a list of available countries and their data including tax rates
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/payment_countries/{order_id} [get]
func (h *CountryRoute) getPaymentCountries(ctx echo.Context) error {
	req := &grpc.GetCountriesListForOrderRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetCountriesListForOrder(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetCountriesListForOrder")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
