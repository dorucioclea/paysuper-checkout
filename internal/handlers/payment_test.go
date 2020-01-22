package handlers

import (
	"errors"
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
)

type PaymentTestSuite struct {
	suite.Suite
	router *PaymentRoute
	caller *test.EchoReqResCaller
}

func Test_Payment(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}

func (suite *PaymentTestSuite) SetupTest() {
	var e error

	settings := test.DefaultSettings()
	srv := common.Services{}

	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewPaymentRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})

	if e != nil {
		panic(e)
	}
}

func (suite *PaymentTestSuite) TearDownTest() {}

// Test ProcessCreatePayment route
func (suite *PaymentTestSuite) executeProcessCreatePaymentTest(body string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.NoAuthGroupPath + paymentPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())
}

func (suite *PaymentTestSuite) Test_ProcessCreatePayment_Ok() {
	body := `{"test": "test", "one": true, "two": false}`

	bill := &billMock.BillingService{}
	bill.On("PaymentCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentCreateResponse{Status: billing.ResponseStatusOk, RedirectUrl: "url", NeedRedirect: true}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessCreatePaymentTest(body)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.Regexp(suite.T(), "redirect_url", res.Body.String())
}

func (suite *PaymentTestSuite) Test_ProcessCreatePayment_BindError() {
	body := `<some_string>`

	res, err := suite.executeProcessCreatePaymentTest(body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestDataInvalid, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentTestSuite) Test_ProcessCreatePayment_BillingReturnError() {
	body := `{}`

	bill := &billMock.BillingService{}
	bill.On("PaymentCreateProcess", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessCreatePaymentTest(body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentTestSuite) Test_ProcessCreatePayment_BillingResponseStatusError() {
	body := `{}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("PaymentCreateProcess", mock2.Anything, mock2.Anything).
		Return(&billing.PaymentCreateResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeProcessCreatePaymentTest(body)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}
