package handlers

import (
	"errors"
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
)

type CountryTestSuite struct {
	suite.Suite
	router *CountryRoute
	caller *test.EchoReqResCaller
}

func Test_Country(t *testing.T) {
	suite.Run(t, new(CountryTestSuite))
}

func (suite *CountryTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewCountryRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})

	if e != nil {
		panic(e)
	}
}

func (suite *CountryTestSuite) TearDownTest() {}

// Test GetPaymentCountries route
func (suite *CountryTestSuite) executeGetPaymentCountriesTest(orderId string) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterOrderId, orderId).
		Path(common.NoAuthGroupPath + paymentCountriesOrderIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())
}

func (suite *CountryTestSuite) Test_GetPaymentCountries_Ok() {
	orderId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("GetCountriesListForOrder", mock2.Anything, mock2.Anything).
		Return(&billing.GetCountriesListForOrderResponse{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentCountriesTest(orderId)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *CountryTestSuite) Test_GetPaymentCountries_OrderValidationError() {
	orderId := "some_value"

	res, err := suite.executeGetPaymentCountriesTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorIncorrectOrderId.Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *CountryTestSuite) Test_GetPaymentCountries_BillingReturnError() {
	orderId := uuid.New().String()

	bill := &billMock.BillingService{}
	bill.On("GetCountriesListForOrder", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentCountriesTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *CountryTestSuite) Test_GetPaymentCountries_BillingResponseStatusError() {
	orderId := uuid.New().String()
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}

	bill := &billMock.BillingService{}
	bill.On("GetCountriesListForOrder", mock2.Anything, mock2.Anything).
		Return(&billing.GetCountriesListForOrderResponse{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeGetPaymentCountriesTest(orderId)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}
