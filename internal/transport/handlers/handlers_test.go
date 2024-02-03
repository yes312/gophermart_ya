package transport

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	db "gophermart/internal/database"
	"gophermart/internal/mocks"
	"gophermart/internal/models"
	jwtpackage "gophermart/pkg/jwt"
	"gophermart/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	client *resty.Client
	server *httptest.Server
}

func (suite *HandlerTestSuite) SetupSuite() {
	suite.client = resty.New()
}

func TestSuiteTestJSONHandler(t *testing.T) {
	suite.Run(t, &HandlerTestSuite{})
}

func (suite *HandlerTestSuite) TearDownSuite() {
	suite.server.Close()
}

// строка для создания файла с моками
// mockgen -destination="internal/mocks/mock_store.go" -package=mocks "gophermart/internal/database" StoragerDB
func (suite *HandlerTestSuite) TestGetUsers() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	suite.server = httptest.NewServer(http.HandlerFunc(h.Registration))

	type authUserData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	type badAuthUserData struct {
		Login     string `json:"loging"`
		BadString string `json:"badstring"`
	}

	type testCase struct {
		name               string
		login              string
		body               interface{}
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
	}

	tests := []testCase{
		{
			name:  "Логин занят",
			login: "Jhon",
			body: authUserData{
				Login:    "Jhon",
				Password: "12345",
			},
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 409,
			useMock:            true,
		},
		{
			name:  "Ошибка при добавлении пользователя",
			login: "Jhon",
			body: authUserData{
				Login:    "Jhon",
				Password: "12345",
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка при добавлении пользователя"),
			expectedStatusCode: 500,
			useMock:            true,
		},
		{
			name:  "400 — неверный формат запроса;",
			login: "Jhon",
			body: badAuthUserData{
				Login:     "Jhon",
				BadString: "12345",
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка при добавлении пользователя"),
			expectedStatusCode: 400,
			useMock:            false,
		},
	}
	url := "/api/user/register"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().GetUser(h.ctx, test.login).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetBody(test.body).
			SetHeader("Content-Type", "application/json").
			Post(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
	}

	test := testCase{
		name:  "успех",
		login: "Jhon",
		body: authUserData{
			Login:    "Jhon",
			Password: "12345",
		},
		ReturnInterface:    nil,
		ReturnErr:          sql.ErrNoRows,
		expectedStatusCode: 200,
		useMock:            true,
	}

	if test.useMock {
		var mockedDBOperation db.DBOperation
		m.EXPECT().GetUser(h.ctx, test.login).Return(mockedDBOperation)
		m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		hash := "b13f28188ed29c08e6b0a220822e76c2c557a69c480f91924e1a8084004d4c55"
		m.EXPECT().AddUser(h.ctx, test.login, hash).Return(mockedDBOperation)
		m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, nil)

	}

	resp, err := suite.client.R().
		SetBody(test.body).
		SetHeader("Content-Type", "application/json").
		Post(suite.server.URL + url)

	suite.NoError(err)
	suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
	suite.Equal("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDI2NzI5ODYsIlVzZXJJRCI6Ikpob24ifQ.eNyPRFCYvNftB1tsVWfc2jFjUpyiRgzDDnyNeufCyGY"[:30],
		resp.Header().Get("Authorization")[:30])
}

func (suite *HandlerTestSuite) TestLogin() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(http.HandlerFunc(h.Login))
	// 200 — пользователь успешно аутентифицирован;
	// 400 — неверный формат запроса;
	// 401 — неверная пара логин/пароль;
	// 500 — внутренняя ошибка сервера.

	type authUserData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	type badAuthUserData struct {
		Login     string `json:"loging"`
		BadString string `json:"badstring"`
	}

	type testCase struct {
		name               string
		login              string
		body               interface{}
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
	}

	tests := []testCase{
		{
			name:  "неверный логин -пароль(401)",
			login: "Jhon",
			body: authUserData{
				Login:    "Jhon",
				Password: "12345",
			},
			ReturnInterface:    nil,
			ReturnErr:          sql.ErrNoRows,
			expectedStatusCode: 401,
			useMock:            true,
		},
		{
			name:  "Ошибка 500",
			login: "Jhon",
			body: authUserData{
				Login:    "Jhon",
				Password: "12345",
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка при полученнии пользователя"),
			expectedStatusCode: 500,
			useMock:            true,
		},
		{
			name:  "400 — неверный формат запроса;",
			login: "Jhon",
			body: badAuthUserData{
				Login:     "Jhon",
				BadString: "",
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка при добавлении пользователя"),
			expectedStatusCode: 400,
			useMock:            false,
		},
		{
			name:  "200 — пользователь успешно аутентифицирован;",
			login: "Jhon",
			body: authUserData{
				Login:    "Jhon",
				Password: "12345",
			},
			ReturnInterface: models.User{
				Login: "Jhon",
				Hash:  "b13f28188ed29c08e6b0a220822e76c2c557a69c480f91924e1a8084004d4c55",
			},
			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
		},
	}

	url := "/api/user/login"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().GetUser(h.ctx, test.login).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetBody(test.body).
			SetHeader("Content-Type", "application/json").
			SetHeader("authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDI2NzI5ODYsIlVzZXJJRCI6Ikpob24ifQ.eNyPRFCYvNftB1tsVWfc2jFjUpyiRgzDDnyNeufCyGY").
			Post(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
	}

}

func (suite *HandlerTestSuite) TestUploadOrder() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(h.AuthMiddleware(h.UploadOrders))

	// + 200 — номер заказа уже был загружен этим пользователем;
	// + 202 — новый номер заказа принят в обработку;
	// + 400 — неверный формат запроса;
	// + 401 — пользователь не аутентифицирован;
	// 409 — номер заказа уже был загружен другим пользователем;
	// + 422 — неверный формат номера заказа;
	// + 500 — внутренняя ошибка сервера.

	type testCase struct {
		name               string
		orderNumber        string
		userID             string
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
		token              string
	}
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUjhVY"
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUj000"
	tests := []testCase{
		{
			name:               "невалидный номер заказа",
			orderNumber:        "123",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка"),
			expectedStatusCode: 422,
			useMock:            false,
			token:              validToken,
		},
		{
			name:               "неверный номер заказа 2",
			orderNumber:        "12w3",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка"),
			expectedStatusCode: 400,
			useMock:            false,
			token:              validToken,
		},
		{
			name:               "500",
			orderNumber:        "4539148803436467",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 500,
			useMock:            true,
			token:              validToken,
		},
		{
			name:               "401 — пользователь не аутентифицирован",
			orderNumber:        "4539148803436467",
			userID:             "JhonJhon",
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 401,
			useMock:            false,
			token:              invalidToken,
		},
		{
			name:               "202 — новый номер заказа принят в обработку;",
			orderNumber:        "4539148803436467",
			userID:             "Jhon",
			ReturnInterface:    models.OrderUserID{},
			ReturnErr:          nil,
			expectedStatusCode: 202,
			useMock:            true,
			token:              validToken,
		},
		{
			name:        "200 — номер заказа уже был загружен этим пользователем;",
			orderNumber: "4539148803436467",
			userID:      "Jhon",
			ReturnInterface: models.OrderUserID{
				OrderNumber: "4539148803436467",
				UserID:      "Jhon",
			},
			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
			token:              validToken,
		},
		{
			name:        "409 — номер заказа уже был загружен другим пользователем",
			orderNumber: "4539148803436467",
			userID:      "Jhon",
			ReturnInterface: models.OrderUserID{
				OrderNumber: "4539148803436467",
				UserID:      "JhonJhon",
			},
			ReturnErr:          nil,
			expectedStatusCode: 409,
			useMock:            true,
			token:              validToken,
		},
	}

	url := "/api/user/orders"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().AddOrder(h.ctx, test.orderNumber, test.userID).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetBody(test.orderNumber).
			SetHeader("authorization", test.token).
			SetHeader("Content-Type", "application/json").
			Post(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
	}

}

func (suite *HandlerTestSuite) TestGetOrders() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(h.AuthMiddleware(h.GetUploadedOrders))

	// + 200 — успешная обработка запроса.
	// + 204 — нет данных для ответа.
	// + 401 — пользователь не авторизован.
	// + 500 — внутренняя ошибка сервера.

	type testCase struct {
		name               string
		userID             string
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
		token              string
	}
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUjhVY"
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUj000"

	tests := []testCase{
		{
			name:               "401 — пользователь не аутентифицирован",
			userID:             "JhonJhon",
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 401,
			useMock:            false,
			token:              invalidToken,
		},
		{
			name:               "500",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 500,
			useMock:            true,
			token:              validToken,
		},
		{
			name:               "204 — нет данных для ответа.",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          sql.ErrNoRows,
			expectedStatusCode: 204,
			useMock:            true,
			token:              validToken,
		},
		{
			name:   "200 — успешная обработка запроса.",
			userID: "Jhon",
			ReturnInterface: []models.OrderStatus{{
				Number:     "4539148803436467",
				Status:     "NEW",
				Accrual:    0,
				UploadedAt: time.Time{},
			}},
			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
			token:              validToken,
		},
	}

	url := "/api/user/orders"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().GetOrders(h.ctx, test.userID).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetHeader("authorization", test.token).
			SetHeader("Content-Type", "application/json").
			Get(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
		if test.ReturnInterface != nil {

			var order []models.OrderStatus
			err = json.Unmarshal(resp.Body(), &order)
			suite.NoError(err)
			suite.Equal(test.ReturnInterface, order)

		}

	}
}

func (suite *HandlerTestSuite) TestGetBalance() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(h.AuthMiddleware(h.GetBalance))

	// 200 — успешная обработка запроса.
	// 401 — пользователь не авторизован.
	// 500 — внутренняя ошибка сервера.

	type testCase struct {
		name               string
		userID             string
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
		token              string
	}
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUjhVY"
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUj000"

	tests := []testCase{
		{
			name:               "401 — пользователь не аутентифицирован",
			userID:             "JhonJhon",
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 401,
			useMock:            false,
			token:              invalidToken,
		},
		{
			name:               "500",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 500,
			useMock:            true,
			token:              validToken,
		},
		{
			name:   "200 — успешная обработка запроса",
			userID: "Jhon",
			ReturnInterface: models.Balance{
				Current:  100,
				Withdraw: 5,
			},

			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
			token:              validToken,
		},
	}

	url := "/api/user/balance"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().GetBalance(h.ctx, test.userID).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetHeader("authorization", test.token).
			SetHeader("Content-Type", "application/json").
			Get(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
		if test.ReturnInterface != nil {

			var balance models.Balance
			err = json.Unmarshal(resp.Body(), &balance)
			suite.NoError(err)
			suite.Equal(test.ReturnInterface, balance)

		}
	}
}

func (suite *HandlerTestSuite) TestWithdrawBalance() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(h.AuthMiddleware(h.WithdrawBalance))

	// 200 — успешная обработка запроса;
	// +401 — пользователь не авторизован;
	// +402 — на счету недостаточно средств;
	// +422 — неверный номер заказа;
	// +500 — внутренняя ошибка сервера.

	type testCase struct {
		name               string
		userID             string
		data               models.OrderSum
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
		token              string
	}
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUjhVY"
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUj000"

	tests := []testCase{
		{
			name:               "401 — пользователь не аутентифицирован",
			userID:             "JhonJhon",
			data:               models.OrderSum{},
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 401,
			useMock:            false,
			token:              invalidToken,
		},
		{
			name:   "500",
			userID: "Jhon",
			data: models.OrderSum{
				OrderNumber: "4539148803436467",
				Sum:         100,
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 500,
			useMock:            true,
			token:              validToken,
		},
		{
			name:   "422 — неверный номер заказа;",
			userID: "Jhon",
			data: models.OrderSum{
				OrderNumber: "123",
				Sum:         100,
			},
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 422,
			useMock:            false,
			token:              validToken,
		},
		{
			name:   "402 — на счету недостаточно средств",
			userID: "Jhon",
			data: models.OrderSum{
				OrderNumber: "4539148803436467",
				Sum:         100,
			},
			ReturnInterface:    nil,
			ReturnErr:          db.ErrNotEnoughFunds,
			expectedStatusCode: 402,
			useMock:            true,
			token:              validToken,
		},
		{
			name:   "200 — успешная обработка запроса;",
			userID: "Jhon",
			data: models.OrderSum{
				OrderNumber: "4539148803436467",
				Sum:         100,
			},
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
			token:              validToken,
		},
	}

	url := "/api/user/balance/withdraw"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().WithdrawBalance(h.ctx, test.userID, test.data).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetHeader("authorization", test.token).
			SetHeader("Content-Type", "application/json").
			SetBody(test.data).
			Post(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)

	}
}

func (suite *HandlerTestSuite) TestGetWithdrawals() {

	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStoragerDB(ctrl)
	logger, err := logger.NewLogger("Info")
	suite.NoError(err)
	h := New(ctx, m, logger)
	h.AuthToken = *jwtpackage.NewToken(time.Duration(999*time.Hour), "secret")
	suite.server = httptest.NewServer(h.AuthMiddleware(h.GetWithdrawals))

	// 200 — успешная обработка запроса.
	// 204 — нет ни одного списания.
	// 401 — пользователь не авторизован.
	// 500 — внутренняя ошибка сервера.

	type testCase struct {
		name               string
		userID             string
		ReturnInterface    interface{}
		ReturnErr          error
		expectedStatusCode int
		useMock            bool
		token              string
	}
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUjhVY"
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzMTUyOTUsIlVzZXJJRCI6Ikpob24ifQ.Wx3EpITNO9gjrNLJil9vm38zYDNCFYuNNT9d1DUj000"

	tests := []testCase{
		{
			name:               "401 — пользователь не аутентифицирован",
			userID:             "JhonJhon",
			ReturnInterface:    nil,
			ReturnErr:          nil,
			expectedStatusCode: 401,
			useMock:            false,
			token:              invalidToken,
		},
		{
			name:               "500",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          errors.New("ошибка 500"),
			expectedStatusCode: 500,
			useMock:            true,
			token:              validToken,
		},
		{
			name:               "204 — нет ни одного списания.",
			userID:             "Jhon",
			ReturnInterface:    nil,
			ReturnErr:          sql.ErrNoRows,
			expectedStatusCode: 204,
			useMock:            true,
			token:              validToken,
		},
		{
			name:   "200 — успешная обработка запроса.",
			userID: "Jhon",
			ReturnInterface: []models.Withdrawal{
				{
					OrderNumber: "4539148803436467",
					Sum:         100,
					ProcessedAt: time.Time{},
				},
			},
			ReturnErr:          nil,
			expectedStatusCode: 200,
			useMock:            true,
			token:              validToken,
		},
	}

	url := "/api/user/withdrawals"
	for _, test := range tests {

		if test.useMock {
			var mockedDBOperation db.DBOperation
			m.EXPECT().GetWithdrawals(h.ctx, test.userID).Return(mockedDBOperation)
			m.EXPECT().WithRetry(h.ctx, mockedDBOperation).Return(test.ReturnInterface, test.ReturnErr)
		}

		resp, err := suite.client.R().
			SetHeader("authorization", test.token).
			SetHeader("Content-Type", "application/json").
			Get(suite.server.URL + url)

		suite.NoError(err)
		suite.Equal(test.expectedStatusCode, resp.StatusCode(), test.name)
		if test.ReturnInterface != nil {

			var withdrawals []models.Withdrawal
			err = json.Unmarshal(resp.Body(), &withdrawals)
			suite.NoError(err)
			suite.Equal(test.ReturnInterface, withdrawals)

		}
	}
}
