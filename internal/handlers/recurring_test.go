package handlers

import (
	"errors"
	"github.com/globalsign/mgo/bson"
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

type RecurringTestSuite struct {
	suite.Suite
	router *RecurringRoute
	caller *test.EchoReqResCaller
}

func Test_Recurring(t *testing.T) {
	suite.Run(t, new(RecurringTestSuite))
}

func (suite *RecurringTestSuite) SetupTest() {
	var e error

	settings := test.DefaultSettings()
	srv := common.Services{}

	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewRecurringRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})

	if e != nil {
		panic(e)
	}
}

func (suite *RecurringTestSuite) TearDownTest() {}

// Test RemoveSavedCard route
func (suite *RecurringTestSuite) executeRemoveSavedCardTest(body string, cookie *http.Cookie) (*httptest.ResponseRecorder, error) {
	return suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		Init(test.ReqInitJSON()).
		AddCookie(cookie).
		BodyString(body).
		Exec(suite.T())
}

func (suite *RecurringTestSuite) Test_RemoveSavedCard_Ok() {
	body := `{"id": "ffffffffffffffffffffffff"}`
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("DeleteSavedCard", mock2.Anything, mock2.Anything).
		Return(&billing.EmptyResponseWithStatus{Status: billing.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRemoveSavedCardTest(body, cookie)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *RecurringTestSuite) Test_RemoveSavedCard_BindError() {
	body := `<some_string>`
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	res, err := suite.executeRemoveSavedCardTest(body, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *RecurringTestSuite) Test_RemoveSavedCard_ValidationError() {
	body := `{"id": "string"}`
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	res, err := suite.executeRemoveSavedCardTest(body, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("").Message, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *RecurringTestSuite) Test_RemoveSavedCard_BillingReturnError() {
	body := `{"id": "ffffffffffffffffffffffff"}`
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("DeleteSavedCard", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRemoveSavedCardTest(body, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *RecurringTestSuite) Test_RemoveSavedCard_BillingResponseStatusError() {
	body := `{"id": "ffffffffffffffffffffffff"}`
	msg := &billing.ResponseErrorMessage{Message: "error", Code: "code"}
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Duration(suite.router.cfg.CustomerTokenCookiesLifetimeHours) * time.Second)
	cookie.HttpOnly = true

	bill := &billMock.BillingService{}
	bill.On("DeleteSavedCard", mock2.Anything, mock2.Anything).
		Return(&billing.EmptyResponseWithStatus{Status: billing.ResponseStatusBadData, Message: msg}, nil)
	suite.router.dispatch.Services.Billing = bill

	res, err := suite.executeRemoveSavedCardTest(body, cookie)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), msg, httpErr.Message)
	assert.NotEmpty(suite.T(), res.Body.String())
}
