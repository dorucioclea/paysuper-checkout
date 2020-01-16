package common

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
)

type Dictionary struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const (
	RequestParameterId        = "id"
	RequestParameterOrderId   = "order_id"
	RequestParameterZipUsa    = "zip_usa"
	RequestParameterReceiptId = "receipt_id"

	QueryParameterNameUtmMedium   = "utm_medium"
	QueryParameterNameUtmCampaign = "utm_campaign"
	QueryParameterNameUtmSource   = "utm_source"

	ErrorMessageMask = "field validation for '%s' failed on the '%s' tag"

	HeaderAcceptLanguage      = "Accept-Language"
	HeaderUserAgent           = "User-Agent"
	HeaderXApiSignatureHeader = "X-API-SIGNATURE"
	HeaderReferer             = "referer"

	CustomerTokenCookiesName = "_ps_ctkn"

	ErrorFieldService = "service"
	ErrorFieldMethod  = "method"
	ErrorFieldRequest = "request"

	InternalErrorTemplate = "internal error"

	ValidationParameterOrderId   = "OrderId"
	ValidationParameterOrderUuid = "OrderUuid"
)

func LogSrvCallFailedGRPC(log logger.Logger, err error, name, method string, req interface{}) {
	log.Error(billing.ErrorGrpcServiceCallFailed,
		logger.PairArgs(
			ErrorFieldService, name,
			ErrorFieldMethod, method,
		),
		logger.WithPrettyFields(logger.Fields{"err": err, ErrorFieldRequest: req}),
	)
}
