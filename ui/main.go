package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Config struct {
	conferenceAPIAddr    string
	keycloakBaseURL      string
	keycloakRealm        string
	keycloakClientID     string
	keycloakClientSecret string
}

func main() {

	templates, err := template.Must(template.New(""), nil).ParseGlob("templates/*.html")

	if err != nil {
		log.Fatal(err.Error())
	}

	config := readConfig()

	oauthConfig := &oauth2.Config{
		ClientID:     config.keycloakClientID,
		ClientSecret: config.keycloakClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.keycloakBaseURL + "/realms/" + config.keycloakRealm + "/protocol/openid-connect/auth",
			TokenURL: config.keycloakBaseURL + "/realms/" + config.keycloakRealm + "/protocol/openid-connect/token",
		},
		RedirectURL: "http://localhost:8081/oidc",
		Scopes:      []string{"openid", "profile", "email"},
	}

	parser := &jwt.Parser{UseJSONNumber: false, SkipClaimsValidation: true}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	sessionStorage := make(map[string]interface{})

	r.Group(func(sub chi.Router) {
		sub.Use(keycloakMiddleware(sessionStorage, oauthConfig, parser))
		sub.Get("/", func(w http.ResponseWriter, r *http.Request) {

			token := r.Context().Value("token").(*oauth2.Token)

			request, _ := http.NewRequest(http.MethodGet, "http://"+config.conferenceAPIAddr, nil)

			//always check erros, this is just a demo

			token.SetAuthHeader(request)

			res, err := http.DefaultClient.Do(request)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			var conferences []string

			if res.StatusCode != 200 {
				conferences = []string{"unable to fetch conferences list, got status: " + res.Status}
			} else {
				err = json.NewDecoder(res.Body).Decode(&conferences)
			}

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			user := r.Context().Value("user").(string)

			token_raw := r.Context().Value("token_raw").(string)

			if err := templates.ExecuteTemplate(w, "home", map[string]interface{}{
				"conferences": conferences,
				"user":        user,
				"token":       token_raw,
				"accountURL":  config.keycloakBaseURL + "/realms/" + config.keycloakRealm + "/account"}); err != nil {
				w.Write([]byte(err.Error()))
			}

		})
	})

	r.Get("/oidc", func(w http.ResponseWriter, r *http.Request) {

		code := r.URL.Query().Get("code")

		token, err := oauthConfig.Exchange(context.Background(), code)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		sessionID := uuid.NewString()

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID,
			// Secure: true,
			HttpOnly: true,
			Path:     "/",
		})

		sessionStorage[sessionID] = token

		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	})

	log.Println("server started")
	http.ListenAndServe(":8081", r)
}

func readConfig() *Config {
	conferenceAPIAddr := os.Getenv("CONFERENCE_API_URL")

	if conferenceAPIAddr == "" {
		conferenceAPIAddr = "localhost:8082"
	}

	keycloakBaseURL := os.Getenv("KEYCLOAK_BASE_URL")

	if keycloakBaseURL == "" {
		keycloakBaseURL = "localhost:8080"
	}

	keycloakRealm := os.Getenv("KEYCLOAK_REALM")

	if keycloakRealm == "" {
		keycloakRealm = "blablaconf"
	}

	keycloakClientID := os.Getenv("KEYCLOAK_CLIENT_ID")

	if keycloakClientID == "" {
		keycloakClientID = "ui"
	}

	keycloakClientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")

	if keycloakClientSecret == "" {
		keycloakClientSecret = ""
	}

	return &Config{
		conferenceAPIAddr:    conferenceAPIAddr,
		keycloakBaseURL:      keycloakBaseURL,
		keycloakRealm:        keycloakRealm,
		keycloakClientID:     keycloakClientID,
		keycloakClientSecret: keycloakClientSecret,
	}
}

func keycloakMiddleware(sessionStorage map[string]interface{}, config *oauth2.Config, parser *jwt.Parser) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			sessionCookie, err := r.Cookie("session")

			authURL := config.AuthCodeURL("any")

			//if cookie is not present, it raises an error as well
			if err != nil {
				http.Redirect(w, r, authURL, http.StatusPermanentRedirect)
				return
			}

			sessionID := sessionCookie.Value

			value, ok := sessionStorage[sessionID]

			if !ok {
				http.Redirect(w, r, authURL, http.StatusPermanentRedirect)
				return
			}

			token := value.(*oauth2.Token)

			idToken := token.Extra("id_token")

			if !token.Valid() {
				http.Redirect(w, r, authURL, http.StatusPermanentRedirect)
			}

			jwtToken, _, err := parser.ParseUnverified(idToken.(string), jwt.MapClaims{})

			//if cookie is not present, it raises an error as well
			if err != nil {
				http.Redirect(w, r, authURL, http.StatusPermanentRedirect)
				return
			}

			tokenMap := make(map[string]interface{})

			tokenMap["access_token"] = token.AccessToken
			tokenMap["id_token"] = idToken
			tokenMap["refresh_token"] = token.RefreshToken
			tokenMap["token_type"] = token.Type()
			tokenMap["expires_in"] = token.Extra("expires_in")

			tokenBytes, _ := json.MarshalIndent(tokenMap, "", "")

			//TODO: check error, this is just a demo

			r = r.WithContext(context.WithValue(r.Context(), "user", jwtToken.Claims.(jwt.MapClaims)["name"]))
			r = r.WithContext(context.WithValue(r.Context(), "token_raw", string(tokenBytes)))
			r = r.WithContext(context.WithValue(r.Context(), "token", token))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
