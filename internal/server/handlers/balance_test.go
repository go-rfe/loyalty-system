package handlers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	"github.com/go-rfe/loyalty-system/internal/repository/orders/mocks"
	"github.com/go-rfe/loyalty-system/internal/server/handlers"
	"github.com/golang/mock/gomock"
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
				store.EXPECT().GetBalance(gomock.Any(), "test").Return(&orders.Balance{
					Current:   700.8,
					Withdrawn: 50.4,
				}, nil)
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
				store.EXPECT().GetWithdrawals(gomock.Any(), "test").Return([]orders.Withdraw{
					{
						Order:       "2377225624",
						Sum:         500,
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