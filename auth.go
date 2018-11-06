package main

import (
	"github.com/stretchr/objx"
	"fmt"
	"net/http"
	"strings"

	"github.com/stretchr/gomniauth"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd podczas próby pobrania dostawcy %s: %s", provider, err), http.StatusBadRequest)
			return
		}
		loginURL, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd wywołania GetBeginAuthURL dla %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("location", loginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		// log.Println("TO DO: obsługa logowania z użyciem ", provider)

	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd podczas próby pobrania dostawcy %s: %s", provider, err), http.StatusBadRequest)
			return
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd próby dokonania uwierzytelniania dla %s: %s", provider, err), http.StatusInternalServerError)
			return 
		}

		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd podczas próby pobrania użytkownika z %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		authCookieValue := objx.New(map[string]interface{}{
			"name": user.Name(),
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: authCookieValue,
			Path: "/",
		})

		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Akcja autoryzacyjna %s nie jest obsługiwana", action)
	}
}

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")
	if err == http.ErrNoCookie {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.next.ServeHTTP(w, r)
}

// MustAuth ...
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}
