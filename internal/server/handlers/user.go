package handlers

import (
	"fmt"
	"net/http"

	"context"
	"encoding/json"
	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/models"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

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

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		err = user.ValidateFields()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		err = user.SetPassword(user.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

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

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		err = user.ValidateFields()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		err = ValidateUser(requestContext, user, userStore)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %q", err), http.StatusUnauthorized)

			return
		}

		userToken, err := getUserToken(user.Login, authToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Something went wrong during user login: %q", err), http.StatusInternalServerError)

			return
		}
		w.Header().Set("Authorization", "Bearer "+userToken)
		w.WriteHeader(http.StatusOK)
	}
}

func ValidateUser(ctx context.Context, user models.User, userStore users.Store) error {
	existingUser, err := userStore.GetUser(ctx, user.Login)
	if err != nil {
		return err
	}

	return existingUser.CheckPassword(user.Password)
}

func getUserToken(login string, auth *jwtauth.JWTAuth) (string, error) {
	_, tokenString, err := auth.Encode(map[string]interface{}{"sub": login})

	return tokenString, err
}
