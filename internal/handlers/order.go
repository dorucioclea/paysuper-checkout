package handlers

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	u "github.com/PuerkitoBio/purell"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/internal/helpers"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"net/http"
	"time"
)

const (
	orderPath                = "/order"
	orderIdPath              = "/order/:order_id"
	orderReCreatePath        = "/order/recreate"
	orderLanguagePath        = "/orders/:order_id/language"
	orderCustomerPath        = "/orders/:order_id/customer"
	orderBillingAddressPath  = "/orders/:order_id/billing_address"
	orderNotifySalesPath     = "/orders/:order_id/notify_sale"
	orderNotifyNewRegionPath = "/orders/:order_id/notify_new_region"
	orderPlatformPath        = "/orders/:order_id/platform"
	orderReceiptPath         = "/orders/receipt/:receipt_id/:order_id"
	paylinkIdPath            = "/paylink/:id"
)

const (
	errorTemplateName = "error.html"
)

type CreateOrderJsonProjectResponse struct {
	// The unique identifier for the order.
	Id string `json:"id"`
	// The URL of the PaySuper-hosted payment form.
	PaymentFormUrl string `json:"payment_form_url"`
}

type ReCreateOrderRequest struct {
	// The unique identifier for the order.
	Id string `json:"order_id"`
}

type ListOrdersRequest struct {
	MerchantId    string   `json:"merchant_id" validate:"required,hexadecimal,len=24"`
	FileType      string   `json:"file_type" validate:"required"`
	Template      string   `json:"template" validate:"omitempty,hexadecimal"`
	Id            string   `json:"id" validate:"omitempty,uuid"`
	Project       []string `json:"project" validate:"omitempty,dive,hexadecimal,len=24"`
	PaymentMethod []string `json:"payment_method" validate:"omitempty,dive,hexadecimal,len=24"`
	Country       []string `json:"country" validate:"omitempty,dive,alpha,len=2"`
	Status        []string `json:"status," validate:"omitempty,dive,alpha,oneof=created processed canceled rejected refunded chargeback pending"`
	PmDateFrom    int64    `json:"pm_date_from" validate:"omitempty,numeric,gt=0"`
	PmDateTo      int64    `json:"pm_date_to" validate:"omitempty,numeric,gt=0"`
}

type OrderRoute struct {
	dispatch common.HandlerSet
	cfg      *common.Config
	provider.LMT
}

func NewOrderRoute(set common.HandlerSet, cfg *common.Config) *OrderRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "OrderRoute"})
	return &OrderRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

func (h *OrderRoute) Route(groups *common.Groups) {
	groups.Common.POST(orderPath, h.createJson)
	groups.Common.GET(orderIdPath, h.getPaymentFormData)
	groups.Common.POST(orderReCreatePath, h.recreateOrder)
	groups.Common.PATCH(orderLanguagePath, h.changeLanguage)
	groups.Common.PATCH(orderCustomerPath, h.changeCustomer)
	groups.Common.POST(orderBillingAddressPath, h.processBillingAddress)
	groups.Common.POST(orderNotifySalesPath, h.notifySale)
	groups.Common.POST(orderNotifyNewRegionPath, h.notifyNewRegion)
	groups.Common.POST(orderPlatformPath, h.changePlatform)
	groups.Common.GET(orderReceiptPath, h.getReceipt)
	groups.Common.GET(paylinkIdPath, h.getOrderForPaylink)
}

// @summary Create a payment order
// @desc Create a payment order with a customer and order data
// @id orderPathСreateJson
// @tag Payment Order
// @accept application/json
// @produce application/json
// @body billing.OrderCreateRequest
// @success 200 {object} CreateOrderJsonProjectResponse Returns the order ID and payment form URL
// @failure 400 {object} grpc.ResponseErrorMessage The error code and message with the error details
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @router /api/v1/order [post]
func (h *OrderRoute) createJson(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{}

	if err := (&common.OrderJsonBinder{}).Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Cookie = helpers.GetRequestCookie(ctx, h.cfg.CookieName)

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	// If request contain user object then paysuper must check request signature
	if req.User != nil {
		httpErr := common.CheckProjectAuthRequestSignature(h.dispatch, ctx, req.ProjectId)

		if httpErr != nil {
			return httpErr
		}
	}

	ctxReq := ctx.Request().Context()
	req.IssuerUrl = ctx.Request().Header.Get(common.HeaderReferer)

	var (
		order *billing.Order
	)

	// If request contain prepared order identifier than try to get order by this identifier
	if req.PspOrderUuid != "" {
		req := &billing.IsOrderCanBePayingRequest{
			OrderId:   req.PspOrderUuid,
			ProjectId: req.ProjectId,
		}
		rsp, err := h.dispatch.Services.Billing.IsOrderCanBePaying(ctxReq, req)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "IsOrderCanBePaying")
		}

		if rsp.Status != billing.ResponseStatusOk {
			return echo.NewHTTPError(int(rsp.Status), rsp.Message)
		}

		order = rsp.Item
	} else {
		rsp, err := h.dispatch.Services.Billing.OrderCreateProcess(ctxReq, req)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "OrderCreateProcess")
		}

		if rsp.Status != http.StatusOK {
			return echo.NewHTTPError(int(rsp.Status), rsp.Message)
		}

		order = rsp.Item
	}

	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: h.cfg.OrderInlineFormUrlMask + order.Uuid,
	}

	return ctx.JSON(http.StatusOK, response)
}

// @summary Get the order data
// @desc Get the order data to render a payment form
// @id orderIdPathGetPaymentFormData
// @tag Order
// @accept application/json
// @produce application/json
// @success 200 {object} grpc.PaymentFormJsonData Returns the order data to render a payment form
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/order/{order_id} [get]
func (h *OrderRoute) getPaymentFormData(ctx echo.Context) error {
	req := &billing.PaymentFormJsonDataRequest{
		Locale:  ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		Ip:      ctx.RealIP(),
		Referer: ctx.Request().Header.Get(common.HeaderReferer),
		Cookie:  helpers.GetRequestCookie(ctx, h.cfg.CookieName),
	}

	h.dispatch.AwareSet.L().Info(
		"debug_token",
		logger.WithPrettyFields(logger.Fields{"cookie": req.Cookie, "ip": req.Ip}),
	)

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormJsonDataProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "PaymentFormJsonDataProcess")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	expire := time.Now().Add(time.Duration(h.cfg.CustomerTokenCookiesLifetimeHours) * time.Hour)
	helpers.SetResponseCookie(ctx, h.cfg.CookieName, res.Cookie, h.cfg.CookieDomain, expire)

	return ctx.JSON(http.StatusOK, res.Item)
}

// @summary Recreate a payment order
// @desc Recreate a payment order using the order UUID for the old order
// @id orderReCreatePathRecreateOrder
// @tag Order
// @accept application/json
// @produce application/json
// @body ReCreateOrderRequest
// @success 200 {object} CreateOrderJsonProjectResponse Returns the order ID and payment form URL
// @failure 400 {object} grpc.ResponseErrorMessage The error code and message with the error details
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @router /api/v1/order/recreate [post]
func (h *OrderRoute) recreateOrder(ctx echo.Context) error {
	req := &billing.OrderReCreateProcessRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.OrderReCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "OrderReCreateProcess")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	order := res.Item
	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: h.cfg.OrderInlineFormUrlMask + order.Uuid,
	}

	return ctx.JSON(http.StatusOK, response)
}

// @summary Change the language
// @desc Change the language using the order ID
// @id orderLanguagePathChangeLanguage
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.PaymentFormUserChangeLangRequest
// @success 200 {object} billing.PaymentFormDataChangeResponseItem Returns the order data with a new language
// @failure 400 {object} grpc.ResponseErrorMessage The error code and message with the error details
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/language [patch]
func (h *OrderRoute) changeLanguage(ctx echo.Context) error {
	req := &billing.PaymentFormUserChangeLangRequest{
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormLanguageChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "PaymentFormLanguageChanged")
	}

	if res.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @summary Change a customer
// @desc Change a customer using the order ID
// @id orderCustomerPathChangeCustomer
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.PaymentFormUserChangePaymentAccountRequest
// @success 200 {object} billing.PaymentFormDataChangeResponseItem Returns an order data with a new customers's data
// @failure 400 {object} grpc.ResponseErrorMessage The error code and message with the error details
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/customer [patch]
func (h *OrderRoute) changeCustomer(ctx echo.Context) error {
	req := &billing.PaymentFormUserChangePaymentAccountRequest{
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPaymentAccountChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "PaymentFormPaymentAccountChanged")
	}

	if res.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @summary Change the billing address for the order
// @desc Change the billing address for the order using the order's unique identifier
// @id orderBillingAddressPathProcessBillingAddress
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.ProcessBillingAddressRequest
// @success 200 {object} grpc.ProcessBillingAddressResponseItem Returns the order data with the new billing address
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 403 {object} grpc.ResponseErrorMessage Payments are disallowed for this country
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/billing_address [post]
func (h *OrderRoute) processBillingAddress(ctx echo.Context) error {
	req := &billing.ProcessBillingAddressRequest{
		Cookie: helpers.GetRequestCookie(ctx, h.cfg.CookieName),
		Ip:     ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ProcessBillingAddress(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "ProcessBillingAddress")
	}

	if res.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	expire := time.Now().Add(time.Duration(h.cfg.CustomerTokenCookiesLifetimeHours) * time.Hour)
	helpers.SetResponseCookie(ctx, h.cfg.CookieName, res.Cookie, h.cfg.CookieDomain, expire)

	return ctx.JSON(http.StatusOK, res.Item)
}

// @summary Subscribe to sales and discounts notifications
// @desc Subscribe to sales and discounts notifications using the order ID
// @id orderNotifySalesPathNotifySale
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.SetUserNotifyRequest
// @success 200 {string} Returns an empty response body if the customer has successfully subscribed
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/notify_sale [post]
func (h *OrderRoute) notifySale(ctx echo.Context) error {
	req := &billing.SetUserNotifyRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	_, err := h.dispatch.Services.Billing.SetUserNotifySales(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "SetUserNotifySales")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @summary Subscribe to notifications about new regions
// @desc Subscribe to get an email notification when a new region becomes available at PaySuper to receive payments
// @id orderNotifyNewRegionPathNotifyNewRegion
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.SetUserNotifyRequest
// @success 200 {string} Returns an empty response body if the customer has successfully subscribed
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/notify_new_region [post]
func (h *OrderRoute) notifyNewRegion(ctx echo.Context) error {
	req := &billing.SetUserNotifyRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	_, err := h.dispatch.Services.Billing.SetUserNotifyNewRegion(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "SetUserNotifyNewRegion")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @summary Change an order platform
// @desc Change an order platform by the order ID
// @id orderPlatformPathChangePlatform
// @tag Order
// @accept application/json
// @produce application/json
// @body grpc.PaymentFormUserChangePlatformRequest
// @success 200 {object} billing.PaymentFormDataChangeResponseItem Returns an order data with a new platform name for a payment form rendering
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/{order_id}/platform [post]
func (h *OrderRoute) changePlatform(ctx echo.Context) error {
	req := &billing.PaymentFormUserChangePlatformRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPlatformChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "PaymentFormPlatformChanged")
	}

	if res.Status != billing.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @summary Getting a payment receipt
// @desc Getting a payment receipt data for rendering
// @id orderReceiptPathGetReceipt
// @tag Order
// @accept application/json
// @produce application/json
// @success 200 {object} billing.OrderReceipt Returns a payment receipt data
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param receipt_id path {string} true The unique identifier for the receipt
// @param order_id path {string} true The unique identifier for the order
// @router /api/v1/orders/receipt/{receipt_id}/{order_id} [get]
func (h *OrderRoute) getReceipt(ctx echo.Context) error {
	req := &billing.OrderReceiptRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.OrderReceipt(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "OrderReceipt")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Receipt)
}

// @summary Create a payment link
// @desc Create a payment link by a paylink ID
// @id paylinkIdPathGetOrderForPaylink
// @tag Order
// @accept application/json
// @produce application/json
// @success 200 {string} Returns a payment link URL
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @param id path {string} true The unique identifier for the paylink
// @router /api/v1/paylink/{id} [get]
func (h *OrderRoute) getOrderForPaylink(ctx echo.Context) error {
	paylinkId := ctx.Param(common.RequestParameterId)

	go func() {
		req := &billing.PaylinkRequestById{Id: paylinkId}
		// call with background context to prevent request abandoning when redirect will bw returned in response below
		_, err := h.dispatch.Services.Billing.IncrPaylinkVisits(context.Background(), req)

		if err != nil {
			common.LogSrvCallFailedGRPC(h.L(), err, billing.ServiceName, "IncrPaylinkVisits", req)
		}
	}()

	qParams := ctx.QueryParams()

	req := &billing.OrderCreateByPaylink{
		PaylinkId:   paylinkId,
		PayerIp:     ctx.RealIP(),
		IssuerUrl:   ctx.Request().Header.Get(common.HeaderReferer),
		UtmSource:   qParams.Get(common.QueryParameterNameUtmSource),
		UtmMedium:   qParams.Get(common.QueryParameterNameUtmMedium),
		UtmCampaign: qParams.Get(common.QueryParameterNameUtmCampaign),
		IsEmbedded:  false,
		Cookie:      helpers.GetRequestCookie(ctx, h.cfg.CookieName),
	}

	res, err := h.dispatch.Services.Billing.OrderCreateByPaylink(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, billing.ServiceName, "OrderCreateByPaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	inlineFormRedirectUrl, err := u.NormalizeURLString(
		h.cfg.OrderInlineFormUrlMask+res.Item.Uuid+"?"+qParams.Encode(),
		u.FlagsUsuallySafeGreedy|u.FlagRemoveDuplicateSlashes,
	)

	if err != nil {
		h.L().Error("NormalizeURLString failed", logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.Redirect(http.StatusFound, inlineFormRedirectUrl)
}
