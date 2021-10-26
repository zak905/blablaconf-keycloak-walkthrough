package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	templates, err := template.Must(template.New(""), nil).ParseGlob("templates/*.html")

	if err != nil {
		log.Fatal(err.Error())
	}

	conferenceAPIAddr := os.Getenv("CONFERENCE_API_URL")

	if conferenceAPIAddr == "" {
		conferenceAPIAddr = "localhost:8082"
	}

	http.Get("http://" + conferenceAPIAddr)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		res, err := http.Get("http://" + conferenceAPIAddr)
		if err != nil {
			w.Write([]byte(err.Error()))
		}

		var conferences []string

		err = json.NewDecoder(res.Body).Decode(&conferences)

		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := templates.ExecuteTemplate(w, "home", map[string]interface{}{"conferences": conferences}); err != nil {
			w.Write([]byte(err.Error()))
		}

	})
	log.Println("server started")
	http.ListenAndServe(":8081", r)

}
