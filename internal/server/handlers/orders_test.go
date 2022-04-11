package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	"github.com/go-rfe/loyalty-system/internal/repository/orders/mocks"
	"github.com/go-rfe/loyalty-system/internal/server/handlers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authHeader = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiJ0ZXN0In0." +
		"Gmlw_dPyBS-autswceWkocF9ELiEHKeS86-MHgG8MhY"
)

type wantOrders struct {
	code int
	data string
}

type testOrder struct {
	name       string
	method     string
	url        string
	order      string
	authHeader string
	buildStubs func(store *mocks.MockStore)
	want       wantOrders
}

func TestOrdersHandlers(t *testing.T) {
	testOrders := []testOrder{
		{
			name:       "OK Create new order",
			method:     http.MethodPost,
			url:        "/api/user/orders",
			order:      "267876232367723",
			authHeader: authHeader,
			want: wantOrders{
				code: http.StatusAccepted,
				data: "",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().CreateOrder(gomock.Any(), "test", "267876232367723")
			},
		},
		{
			name:       "OK Create order",
			method:     http.MethodPost,
			url:        "/api/user/orders",
			order:      "267876232367723",
			authHeader: authHeader,
			want: wantOrders{
				code: http.StatusOK,
				data: "",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().CreateOrder(gomock.Any(), "test", "267876232367723").Return(orders.ErrOrderExists)
			},
		},
		{
			name:       "Bad order number",
			method:     http.MethodPost,
			url:        "/api/user/orders",
			order:      "1111",
			authHeader: authHeader,
			want: wantOrders{
				code: http.StatusUnprocessableEntity,
				data: "",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().CreateOrder(gomock.Any(), "test", "1111").Times(0)
			},
		},
		{
			name:       "Bad order number",
			method:     http.MethodPost,
			url:        "/api/user/orders",
			order:      "267876232367723",
			authHeader: authHeader,
			want: wantOrders{
				code: http.StatusConflict,
				data: "",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().CreateOrder(gomock.Any(), "test", "267876232367723").Return(orders.ErrOtherOrderExists)
			},
		},
		{
			name:       "Unauthorized Create order",
			method:     http.MethodPost,
			url:        "/api/user/orders",
			order:      "267876232367723",
			authHeader: "",
			want: wantOrders{
				code: http.StatusUnauthorized,
				data: "",
			},
			buildStubs: func(store *mocks.MockStore) {
				store.EXPECT().CreateOrder(gomock.Any(), "test", "267876232367723").Times(0)
			},
		},
	}

	jwtToken := jwtauth.New("HS256", []byte("test"), []byte("test"))

	mux := chi.NewRouter()
	store := getOrdersStore(t)
	handlers.RegisterPrivateHandlers(mux, store, jwtToken)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	for _, tt := range testOrders {
		t.Run(tt.name, func(t *testing.T) {
			tt.buildStubs(store)
			testOrdersRequest(t, ts, tt)
		})
	}
}

func testOrdersRequest(t *testing.T, ts *httptest.Server, testData testOrder) {
	t.Helper()

	var body bytes.Buffer
	body.WriteString(testData.order)

	req, err := http.NewRequest(testData.method, ts.URL+testData.url, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", testData.authHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, testData.want.code, resp.StatusCode)
}

func getOrdersStore(t *testing.T) *mocks.MockStore {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := mocks.NewMockStore(ctrl)

	return s
}
