package handlers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/internal/test"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	billMock "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var uuidRegExp = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}"

type OrderTestSuite struct {
	suite.Suite
	router *OrderRoute
	caller *test.EchoReqResCaller
}

func Test_Order(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}

func (suite *OrderTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewOrderRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})

	if e != nil {
		panic(e)
	}
}

func (suite *OrderTestSuite) TearDownTest() {}

// Test CreateJson route
func (suite *OrderTestSuite) executeCreateJsonTest(body string, headers map[string]string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.NoAuthGroupPath + orderPath).
		Init(test.ReqInitJSON()).
		SetHeaders(headers).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_CreateJson_Ok_WithPreparedOrder() {
	orderId := uuid.New().String()
	body := fmt.Sprintf(`{"order": "%s"}`, orderId)
	headers := map[string]string{}

	bill := &billMock.BillingService{}
	bill.On("IsOrderCanBePaying", mock2.Anything, mock2.Anything).
		Return(&billing.IsOrderCanBePayingResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: orderId}}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.Contains(suite.T(), res.Body.String(), "payment_form_url")
	assert.Regexp(suite.T(), uuidRegExp, res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_WithPreparedOrder_BillingReturnError() {
	orderId := uuid.New().String()
	body := fmt.Sprintf(`{"order": "%s"}`, orderId)
	headers := map[string]string{}

	bill := &billMock.BillingService{}
	bill.On("IsOrderCanBePaying", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_WithPreparedOrder_BillingResponseStatusError() {
	orderId := uuid.New().String()
	body := fmt.Sprintf(`{"order": "%s"}`, orderId)
	headers := map[string]string{}
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("IsOrderCanBePaying", mock2.Anything, mock2.Anything).
		Return(&billing.IsOrderCanBePayingResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_Ok_WithoutPreparedOrder() {
	body := "{}"
	orderId := uuid.New().String()
	headers := map[string]string{}

	bill := &billMock.BillingService{}
	bill.On("OrderCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: orderId}}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.Contains(suite.T(), res.Body.String(), "payment_form_url")
	assert.Regexp(suite.T(), uuidRegExp, res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_WithoutPreparedOrder_BillingReturnError() {
	body := ""
	headers := map[string]string{}

	bill := &billMock.BillingService{}
	bill.On("OrderCreateProcess", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_WithoutPreparedOrder_BillingResponseStatusError() {
	body := ""
	headers := map[string]string{}
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("OrderCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_BindingError() {
	body := "<some string>"
	headers := map[string]string{}

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_ValidationError() {
	body := `{"currency": "test"}`
	headers := map[string]string{}

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.NewValidationError("field validation for 'Currency' failed on the 'len' tag"), httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_UserWithoutSignatureHeader() {
	body := `{"user": {"id": "1"}}`
	headers := map[string]string{}

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorMessageSignatureHeaderIsEmpty, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_UserWithBadSignature_BillingError() {
	body := `{"user": {"id": "1"}}`
	headers := map[string]string{common.HeaderXApiSignatureHeader: "signature"}

	bill := &billMock.BillingService{}
	bill.On("CheckProjectRequestSignature", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_CreateJson_UserWithBadSignature_BillingReturnErrorStatus() {
	body := `{"user": {"id": "1"}}`
	headers := map[string]string{common.HeaderXApiSignatureHeader: "signature"}

	bill := &billMock.BillingService{}
	bill.On("CheckProjectRequestSignature", mock2.Anything, mock2.Anything).
		Return(&billing.CheckProjectRequestSignatureResponse{Status: billing.ResponseStatusBadData}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeCreateJsonTest(body, headers)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test GetPaymentFormData route
func (suite *OrderTestSuite) executeGetPaymentFormDataTest(orderId string, cookie *http.Cookie) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderIdPath).
		Init(test.ReqInitJSON()).
		AddCookie(cookie).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_GetPaymentFormData_Ok() {
	orderId := uuid.New().String()
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = "ffffffffffffffffffffffff"
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("PaymentFormJsonDataProcess", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormJsonDataResponse{Status: billing.ResponseStatusOk, Cookie: "setcookie"}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentFormDataTest(orderId, cookie)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.NotEmpty(suite.T(), res.Result().Cookies())
}

func (suite *OrderTestSuite) Test_GetPaymentFormData_OrderValidationError() {
	orderId := "some_value"
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = "ffffffffffffffffffffffff"
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	res, err := suite.executeGetPaymentFormDataTest(orderId, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetPaymentFormData_BillingReturnError() {
	orderId := uuid.New().String()
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = "ffffffffffffffffffffffff"
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("PaymentFormJsonDataProcess", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentFormDataTest(orderId, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetPaymentFormData_BillingResponseStatusError() {
	orderId := uuid.New().String()
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = "ffffffffffffffffffffffff"
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("PaymentFormJsonDataProcess", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormJsonDataResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentFormDataTest(orderId, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test RecreateOrder route
func (suite *OrderTestSuite) executeRecreateOrderTest(orderId string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.NoAuthGroupPath + orderReCreatePath).
		Init(test.ReqInitJSON()).
		BodyString(fmt.Sprintf(`{"order_id": "%s"}`, orderId)).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_RecreateOrder_Ok() {
	orderId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("OrderReCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: orderId}}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRecreateOrderTest(orderId)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.Contains(suite.T(), res.Body.String(), "payment_form_url")
	assert.Regexp(suite.T(), uuidRegExp, res.Body.String())
}

func (suite *OrderTestSuite) Test_RecreateOrder_OrderValidationError() {
	orderId := "some_value"

	res, err := suite.executeRecreateOrderTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_RecreateOrder_BillingReturnError() {
	orderId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("OrderReCreateProcess", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRecreateOrderTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_RecreateOrder_BillingResponseStatusError() {
	orderId := uuid.New().String()
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("OrderReCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRecreateOrderTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test ChangeLanguage route
func (suite *OrderTestSuite) executeChangeLanguageTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPatch).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderLanguagePath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_Ok() {
	orderId := uuid.New().String()
	body := `{"lang": "en"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormLanguageChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_OrderIdEmptyError() {
	orderId := ""
	body := `{"lang": "en"}`

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_BindRequestError() {
	orderId := uuid.New().String()
	body := `<datawrong>`

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_OrderValidationError() {
	orderId := "some_value"
	body := `{"lang": "en"}`

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"lang": "en"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormLanguageChanged", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeLanguage_BillingResponseStatusError() {
	orderId := uuid.New().String()
	body := `{"lang": "en"}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("PaymentFormLanguageChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeLanguageTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test ChangeCustomer route
func (suite *OrderTestSuite) executeChangeCustomerTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPatch).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderCustomerPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_Ok() {
	orderId := uuid.New().String()
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPaymentAccountChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_OrderIdEmptyError() {
	orderId := ""
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_BindError() {
	orderId := uuid.New().String()
	body := `<data wrong>`

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_ValidationError() {
	orderId := uuid.New().String()
	body := `{"method_id": "some_value", "account": "4000000000000002"}`

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPaymentAccountChanged", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangeCustomer_BillingResponseStatusError() {
	orderId := uuid.New().String()
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPaymentAccountChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangeCustomerTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test ProcessBillingAddress route
func (suite *OrderTestSuite) executeProcessBillingAddressTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderBillingAddressPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_Ok() {
	orderId := uuid.New().String()
	body := `{"country": "US", "zip": "98001"}`

	bill := &billMock.BillingService{}
	bill.On("ProcessBillingAddress", mock2.Anything, mock2.Anything).
		Return(&billing.ProcessBillingAddressResponse{Status: billing.ResponseStatusOk, Cookie: "setcookie"}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.NotEmpty(suite.T(), res.Result().Cookies())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_OrderIdEmptyError() {
	orderId := ""
	body := `{"country": "US", "zip": "98001"}`

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_BindError() {
	orderId := uuid.New().String()
	body := "<some wrong body>"

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_ValidationError() {
	orderId := uuid.New().String()
	body := `{"country": "some_value", "zip": "98001"}`

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_ValidationZipError() {
	orderId := uuid.New().String()
	body := `{"country": "US", "zip": "00"}`

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(billing.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectZip.Message, msg.Message)
	assert.Regexp(suite.T(), "Zip", msg.Details)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"country": "US", "zip": "98001"}`

	bill := &billMock.BillingService{}
	bill.On("ProcessBillingAddress", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ProcessBillingAddress_BillingResponseStatusError() {
	orderId := uuid.New().String()
	body := `{"country": "US", "zip": "98001"}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("ProcessBillingAddress", mock2.Anything, mock2.Anything).
		Return(&billing.ProcessBillingAddressResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessBillingAddressTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test NotifySale route
func (suite *OrderTestSuite) executeNotifySaleTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderNotifySalesPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_NotifySale_Ok() {
	orderId := uuid.New().String()
	body := `{"email": "test@test.com"}`

	bill := &billMock.BillingService{}
	bill.On("SetUserNotifySales", mock2.Anything, mock2.Anything).Return(nil, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeNotifySaleTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifySale_OrderIdEmptyError() {
	orderId := ""
	body := `{"email": "test@test.com"}`

	res, err := suite.executeNotifySaleTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifySale_BindError() {
	orderId := uuid.New().String()
	body := "<some wrong body>"

	res, err := suite.executeNotifySaleTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifySale_ValidationError() {
	orderId := uuid.New().String()
	body := `{"email": "test"}`

	res, err := suite.executeNotifySaleTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifySale_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"email": "test@test.com"}`

	bill := &billMock.BillingService{}
	bill.On("SetUserNotifySales", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeNotifySaleTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test NotifyNewRegion route
func (suite *OrderTestSuite) executeNotifyNewRegionTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderNotifyNewRegionPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_NotifyNewRegion_Ok() {
	orderId := uuid.New().String()
	body := `{"email": "test@test.com"}`

	bill := &billMock.BillingService{}
	bill.On("SetUserNotifyNewRegion", mock2.Anything, mock2.Anything).Return(nil, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeNotifyNewRegionTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifyNewRegion_OrderIdEmptyError() {
	orderId := ""
	body := `{"email": "test@test.com"}`

	res, err := suite.executeNotifyNewRegionTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifyNewRegion_BindError() {
	orderId := uuid.New().String()
	body := "<some wrong body>"

	res, err := suite.executeNotifyNewRegionTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifyNewRegion_ValidationError() {
	orderId := uuid.New().String()
	body := `{"email": "test"}`

	res, err := suite.executeNotifyNewRegionTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_NotifyNewRegion_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"email": "test@test.com"}`

	bill := &billMock.BillingService{}
	bill.On("SetUserNotifyNewRegion", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeNotifyNewRegionTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test ChangePlatform route
func (suite *OrderTestSuite) executeChangePlatformTest(orderId string, body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + orderPlatformPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_Ok() {
	orderId := uuid.New().String()
	body := `{"platform": "gog"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPlatformChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_OrderIdEmptyError() {
	orderId := ""
	body := `{"platform": "gog"}`

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message.(billing.ResponseErrorMessage).Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_BindError() {
	orderId := uuid.New().String()
	body := "<some wrong body>"

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_ValidationError() {
	orderId := uuid.New().String()
	body := `{"platform": "g"}`

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_BillingReturnError() {
	orderId := uuid.New().String()
	body := `{"platform": "gog"}`

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPlatformChanged", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_ChangePlatformPayment_BillingResponseStatusError() {
	orderId := uuid.New().String()
	body := `{"platform": "gog"}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("PaymentFormPlatformChanged", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentFormDataChangeResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeChangePlatformTest(orderId, body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test GetReceipt route
func (suite *OrderTestSuite) executeGetReceiptTest(orderId string, receiptId string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterOrderId, orderId, ":"+common.RequestParameterReceiptId, receiptId).
		Path(common.NoAuthGroupPath + orderReceiptPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_GetReceipt_Ok() {
	orderId := uuid.New().String()
	receiptId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("OrderReceipt", mock2.Anything, mock2.Anything).
		Return(&billing.OrderReceiptResponse{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetReceiptTest(orderId, receiptId)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetReceipt_ValidationError() {
	orderId := "string"
	receiptId := uuid.New().String()

	res, err := suite.executeGetReceiptTest(orderId, receiptId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetReceipt_BillingReturnError() {
	orderId := uuid.New().String()
	receiptId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("OrderReceipt", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetReceiptTest(orderId, receiptId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetReceipt_BillingResponseStatusError() {
	orderId := uuid.New().String()
	receiptId := uuid.New().String()
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("OrderReceipt", mock2.Anything, mock2.Anything).
		Return(&billing.OrderReceiptResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetReceiptTest(orderId, receiptId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

// Test GetOrderForPaylink route
func (suite *OrderTestSuite) executeGetOrderForPaylinkTest(id string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, id).
		Path(common.NoAuthGroupPath + paylinkIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())
}

func (suite *OrderTestSuite) Test_GetOrderForPaylink_Ok() {
	id := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("IncrPaylinkVisits", mock2.Anything, mock2.Anything).Return(nil, nil)
	bill.On("OrderCreateByPaylink", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: "fbd3036f-0f1c-4e98-b71c-d4cd61213f90"}}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetOrderForPaylinkTest(id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusFound, res.Code)
	assert.Empty(suite.T(), res.Body.String())

	url, err := res.Result().Location()
	assert.NoError(suite.T(), err)
	assert.Regexp(suite.T(), uuidRegExp, url.String())
}

func (suite *OrderTestSuite) Test_GetOrderForPaylink_Ok_WithGORutineIsFalse() {
	id := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("IncrPaylinkVisits", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	bill.On("OrderCreateByPaylink", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: "fbd3036f-0f1c-4e98-b71c-d4cd61213f90"}}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetOrderForPaylinkTest(id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusFound, res.Code)
	assert.Empty(suite.T(), res.Body.String())

	url, err := res.Result().Location()
	assert.NoError(suite.T(), err)
	assert.Regexp(suite.T(), uuidRegExp, url.String())
}

func (suite *OrderTestSuite) Test_GetOrderForPaylink_BillingReturnError() {
	id := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("IncrPaylinkVisits", mock2.Anything, mock2.Anything).Return(nil, nil)
	bill.On("OrderCreateByPaylink", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetOrderForPaylinkTest(id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetOrderForPaylink_BillingResponseStatusError() {
	id := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("IncrPaylinkVisits", mock2.Anything, mock2.Anything).Return(nil, nil)
	bill.On("OrderCreateByPaylink", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusSystemError}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetOrderForPaylinkTest(id)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OrderTestSuite) Test_GetOrderForPaylink_InvalidFormUrlMask() {
	id := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("IncrPaylinkVisits", mock2.Anything, mock2.Anything).Return(nil, nil)
	bill.On("OrderCreateByPaylink", mock2.Anything, mock2.Anything).
		Return(&billing.OrderCreateProcessResponse{Status: billing.ResponseStatusOk, Item: &billing.Order{Uuid: "uuid"}}, nil)
	suite.router.dispatch.Services.Billing = bill

	suite.router.cfg.OrderInlineFormUrlMask = string([]byte{0x7f})

	res, err := suite.executeGetOrderForPaylinkTest(id)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}
