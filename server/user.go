package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/iamsayantan/talky"
	"github.com/iamsayantan/talky/store"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// contextKey type is used hold values to http context.
type contextKey int

const (
	// KeyAuthUser holds the currently authenticatd user to context.
	KeyAuthUser contextKey = 0

	// AuthorizationHeader is the key from where we extract the authentication token.
	AuthorizationHeader = "Authorization"
)

// JwtSigningSecret used for signing and verifying jwt tokens.
const JwtSigningSecret = "secret"

// JWTClaims represents the JWT token payload
type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type userHandler struct {
	userRepo store.UserRepository
}

func NewUserHandler(repo store.UserRepository) WebHandler {
	return &userHandler{userRepo: repo}
}

func (uh *userHandler) Route() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", uh.register)
	r.Post("/login", uh.login)
	r.Group(func(r chi.Router) {
		r.Use(uh.authenticate)
		r.Get("/me", uh.me)
	})

	return r
}

func (uh *userHandler) register(w http.ResponseWriter, r *http.Request) {
	var registrationReq registerRequest

	err := json.NewDecoder(r.Body).Decode(&registrationReq)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	user := &talky.User{
		FirstName: registrationReq.FirstName,
		LastName:  registrationReq.LastName,
		Username:  registrationReq.Username,
		Password:  registrationReq.Password,
	}

	if err := user.IsValid(); err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}
	_, err = uh.userRepo.FindByUsername(registrationReq.Username)
	if err == nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: "Username already taken"}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(registrationReq.Password), bcrypt.DefaultCost)
	if err := user.IsValid(); err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusInternalServerError, errResp)
		return
	}
	user.Password = string(passwordBytes)

	user, err = uh.userRepo.CreateUser(user)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	token, err := uh.generateAuthToken(user)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	resp := struct {
		User        *talky.User `json:"user"`
		AccessToken string      `json:"access_token"`
	}{User: user, AccessToken: token}

	sendResponse(w, http.StatusCreated, resp)
}

func (uh *userHandler) login(w http.ResponseWriter, r *http.Request) {
	var loginReq loginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		errResp := struct {
			Error string `json:"error"`
		}{Error: "Enter both your username and password"}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	user, err := uh.userRepo.FindByUsername(loginReq.Username)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: "invalid password"}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	accessToken, err := uh.generateAuthToken(user)
	if err != nil {
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	resp := struct {
		User        *talky.User `json:"user"`
		AccessToken string      `json:"access_token"`
	}{User: user, AccessToken: accessToken}

	sendResponse(w, http.StatusOK, resp)
}

func (uh *userHandler) me(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(KeyAuthUser).(*talky.User)
	if !ok {
		errResp := struct {
			Error string `json:"error"`
		}{Error: "Invalid access token"}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	resp := struct {
		User *talky.User `json:"user"`
	}{User: authUser}

	sendResponse(w, http.StatusOK, resp)
}

func (uh *userHandler) generateAuthToken(user *talky.User) (string, error) {
	jwtKey := []byte(JwtSigningSecret)
	expirationTime := time.Now().Add(time.Hour * 24 * 365) // valid for one year

	claims := &JWTClaims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (uh *userHandler) verifyAuthToken(token string) (*talky.User, error) {
	jwtKey := []byte(JwtSigningSecret)

	claims := &JWTClaims{}
	tokn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, errors.New("invalid access token")
	}

	if !tokn.Valid {
		return nil, errors.New("invalid access token")
	}

	user, err := uh.userRepo.FindById(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (uh *userHandler) authenticate(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(AuthorizationHeader)
		ctx := r.Context()

		if token == "" {
			errResp := struct {
				Error string `json:"error"`
			}{Error: "Unauthorized Access"}

			sendResponse(w, http.StatusUnauthorized, errResp)
			return
		}

		user, err := uh.verifyAuthToken(token)

		if err != nil {
			errResp := struct {
				Error string `json:"error"`
			}{Error: err.Error()}

			sendResponse(w, http.StatusUnauthorized, errResp)
			return
		}

		ctx = context.WithValue(ctx, KeyAuthUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
