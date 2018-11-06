package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "Port dla aplikacji.")
	flag.Parse()
	gomniauth.SetSecurityKey("klucz autoryzacyjny")
	gomniauth.WithProviders(
		google.New("452068588331-6ufs8h24ukggttnfi7usuqi2p1sd2mdb.apps.googleusercontent.com",
			"DI4iogPEWCWLASSn8PvnWayQ", "http://localhost:8080/auth/callback/google"),
	)
	r := newRoom()
	// r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()
	log.Println("Uruchamianie serwera WWW na porcie", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {

		log.Fatal("ListenAndServe: ", err)
	}
}
