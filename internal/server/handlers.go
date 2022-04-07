package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

const (
	requestTimeout = 1 * time.Second
)

func RegisterPublicHandlers(mux *chi.Mux, userStore users.Store, auth *jwtauth.JWTAuth) {
	mux.Group(func(r chi.Router) {
		r.Route("/api/user/register", UserRegisterHandler(userStore, auth))
		r.Route("/api/user/login", UserLoginHandler(userStore, auth))
	})
}

func RegisterPrivateHandlers(mux *chi.Mux, userStore users.Store, auth *jwtauth.JWTAuth) {
	mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)

		r.Route("/api/user/orders", OrdersHandler(userStore, auth))
	})
}

func UserRegisterHandler(userStore users.Store, auth *jwtauth.JWTAuth) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", userRegisterHandler(userStore, auth))
	}
}

func UserLoginHandler(userStore users.Store, auth *jwtauth.JWTAuth) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", userLoginHandler(userStore, auth))
	}
}

func userRegisterHandler(store users.Store, auth *jwtauth.JWTAuth) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		var user users.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		err = user.Validate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		err = store.CreateUser(requestContext, user.Login, user.Password)
		if errors.Is(err, users.ErrUserExists) {
			http.Error(
				w,
				err.Error(),
				http.StatusConflict,
			)

			return
		}
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something went wrong during user create: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		userToken, err := getUserToken(user.Login, auth)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something went wrong during user create: %q", err),
				http.StatusInternalServerError,
			)

			return
		}
		w.Header().Set("Authorization", "Bearer "+userToken)
		w.WriteHeader(http.StatusOK)
	}
}

func userLoginHandler(userStore users.Store, authToken *jwtauth.JWTAuth) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		var user users.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		err = user.Validate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		err = userStore.ValidateUser(requestContext, user.Login, user.Password)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Unauthorized: %q", err),
				http.StatusUnauthorized,
			)

			return
		}

		userToken, err := getUserToken(user.Login, authToken)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something went wrong during user login: %q", err),
				http.StatusInternalServerError,
			)

			return
		}
		w.Header().Set("Authorization", "Bearer "+userToken)
		w.WriteHeader(http.StatusOK)
	}
}

func getUserToken(login string, authToken *jwtauth.JWTAuth) (string, error) {
	_, tokenString, err := authToken.Encode(map[string]interface{}{"login": login})

	return tokenString, err
}

func OrdersHandler(userStore users.Store, auth *jwtauth.JWTAuth) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["login"])))
		})
	}
}
