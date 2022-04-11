package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
	"github.com/go-rfe/loyalty-system/internal/server/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type want struct {
	code int
	data string
}

type testUser struct {
	name   string
	method string
	url    string
	user   *users.User
	want   want
}

func TestUserHandlers(t *testing.T) {
	testRegister := []testUser{
		{
			name:   "OK register user",
			method: http.MethodPost,
			url:    "/api/user/register",
			user: &users.User{
				Login:    "test",
				Password: "test",
			},
			want: want{
				code: http.StatusOK,
				data: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.Gmlw_dPyBS-autswceWkocF9ELiEHKeS86-MHgG8MhY",
			},
		},
		{
			name:   "BAD register user",
			method: http.MethodPost,
			url:    "/api/user/register",
			user: &users.User{
				Login:    "",
				Password: "",
			},
			want: want{
				code: http.StatusBadRequest,
				data: "",
			},
		},
	}

	testLogin := []testUser{
		{
			name:   "OK login user",
			method: http.MethodPost,
			url:    "/api/user/login",
			user: &users.User{
				Login:    "test",
				Password: "test",
			},
			want: want{
				code: http.StatusOK,
				data: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.Gmlw_dPyBS-autswceWkocF9ELiEHKeS86-MHgG8MhY",
			},
		},
		{
			name:   "BAD login user",
			method: http.MethodPost,
			url:    "/api/user/login",
			user: &users.User{
				Login:    "",
				Password: "",
			},
			want: want{
				code: http.StatusBadRequest,
				data: "",
			},
		},
		{
			name:   "Unauthorized login user",
			method: http.MethodPost,
			url:    "/api/user/login",
			user: &users.User{
				Login:    "test",
				Password: "baspassword",
			},
			want: want{
				code: http.StatusUnauthorized,
				data: "",
			},
		},
	}

	jwtToken := jwtauth.New("HS256", []byte("test"), []byte("test"))

	mux := chi.NewRouter()

	handlers.RegisterPublicHandlers(mux, users.NewInMemoryStore(), jwtToken)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	for _, tt := range testRegister {
		t.Run(tt.name, func(t *testing.T) {
			testUserRequest(t, ts, tt)
		})
	}

	for _, tt := range testLogin {
		t.Run(tt.name, func(t *testing.T) {
			testUserRequest(t, ts, tt)
		})
	}
}

func testUserRequest(t *testing.T, ts *httptest.Server, testData testUser) {
	t.Helper()

	var body bytes.Buffer
	jsonEncoder := json.NewEncoder(&body)

	require.NoError(t, jsonEncoder.Encode(*testData.user))

	req, err := http.NewRequest(testData.method, ts.URL+testData.url, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, testData.want.code, resp.StatusCode)
	assert.Equal(t, testData.want.data, resp.Header.Get("Authorization"))
}
