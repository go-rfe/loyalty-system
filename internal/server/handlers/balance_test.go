package handlers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/models"
	"github.com/go-rfe/loyalty-system/internal/repository/orders/mocks"
	"github.com/go-rfe/loyalty-system/internal/server/handlers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantBalance struct {
	code int
	data string
}

type testBalance struct {
	name       string
	method     string
	url        string
	authHeader string
	buildStubs func(store *mocks.MockStore)
	want       wantBalance
}

func TestBalanceHandlers(t *testing.T) {
	accrualOne := decimal.NewFromFloat(300.2)
	accrualTwo := decimal.NewFromFloat(451.0)
	processedOrders := []models.Order{
		{
			Accrual: &accrualOne,
		},
		{
			Accrual: &accrualTwo,
		},
	}
	withdrawals := []models.Withdraw{
		{
			Sum: decimal.NewFromFloat(25.1),
		},
		{
			Sum: decimal.NewFromFloat(25.3),
		},
	}
	testOrders := []testBalance{
		{
			name:       "Get Balance",
			method:     http.MethodGet,
			url:        "/api/user/balance",
			authHeader: authHeader,
			want: wantBalance{
				code: http.StatusOK,
				data: "{\"current\":700.8,\"withdrawn\":50.4}\n",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().GetProcessedOrders(gomock.Any(), "test").Return(processedOrders, nil).Times(1)
				store.EXPECT().GetWithdrawals(gomock.Any(), "test").Return(withdrawals, nil).Times(1)
			},
		},
		{
			name:       "Get Withdrawals",
			method:     http.MethodGet,
			url:        "/api/user/balance/withdrawals",
			authHeader: authHeader,
			want: wantBalance{
				code: http.StatusOK,
				data: "[{\"order\":\"2377225624\",\"sum\":500, \"processed_at\":\"2014-11-12T11:45:26.371Z\"}]\n",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().GetWithdrawals(gomock.Any(), "test").Return([]models.Withdraw{
					{
						Order:       "2377225624",
						Sum:         decimal.NewFromFloat(500),
						ProcessedAt: getDate(),
					},
				}, nil)
			},
		},
	}

	jwtToken := jwtauth.New("HS256", []byte("test"), []byte("test"))

	mux := chi.NewRouter()
	store := getBalanceStore(t)
	handlers.RegisterPrivateHandlers(mux, store, jwtToken)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	for _, tt := range testOrders {
		t.Run(tt.name, func(t *testing.T) {
			tt.buildStubs(store)
			testBalanceRequest(t, ts, tt)
		})
	}
}

func testBalanceRequest(t *testing.T, ts *httptest.Server, testData testBalance) {
	t.Helper()

	req, err := http.NewRequest(testData.method, ts.URL+testData.url, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", testData.authHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, testData.want.code, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.JSONEq(t, testData.want.data, string(respBody))

	require.NoError(t, err)

	err = resp.Body.Close()
	if err != nil {
		return
	}
}

func getBalanceStore(t *testing.T) *mocks.MockStore {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := mocks.NewMockStore(ctrl)

	return s
}

func getDate() time.Time {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	t, _ := time.Parse(layout, str)

	return t
}
