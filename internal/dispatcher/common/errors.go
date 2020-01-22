package common

import (
	grpc "github.com/paysuper/paysuper-proto/go/billingpb"
)

// NewManagementApiResponseError
func NewManagementApiResponseError(code, msg string, details ...string) grpc.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return grpc.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

// NewValidationError
func NewValidationError(details string) grpc.ResponseErrorMessage {
	return NewManagementApiResponseError(ErrorValidationFailed.Code, ErrorValidationFailed.Message, details)
}

var (
	ErrorUnknown                       = NewManagementApiResponseError("co000001", "unknown error. try request later")
	ErrorValidationFailed              = NewManagementApiResponseError("co000002", "validation failed")
	ErrorInternal                      = NewManagementApiResponseError("co000003", InternalErrorTemplate)
	ErrorIncorrectOrderId              = NewManagementApiResponseError("co000004", "incorrect order identifier")
	ErrorMessageSignatureHeaderIsEmpty = NewManagementApiResponseError("co000005", "header with request signature can't be empty")
	ErrorRequestParamsIncorrect        = NewManagementApiResponseError("co000006", "incorrect request parameters")
	ErrorRequestDataInvalid            = NewManagementApiResponseError("co000007", "request data invalid")
	ErrorMessageIncorrectZip           = NewManagementApiResponseError("co000008", "incorrect zip code")

	ValidationErrors = map[string]grpc.ResponseErrorMessage{
		ValidationParameterOrderId:   ErrorIncorrectOrderId,
		ValidationParameterOrderUuid: ErrorIncorrectOrderId,
	}
)
