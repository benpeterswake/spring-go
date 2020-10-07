package spring

import (
	"log"
	"net/http"
	"time"

	"github.com/golobby/container"
	internal "github.com/golobby/container/pkg/container"
	"github.com/gorilla/mux"
)

var Container internal.Container

func init() {
	Container = container.NewContainer() // returns container.Container
	Container.Singleton(func() *mux.Router {
		return mux.NewRouter().StrictSlash(true)
	})
	log.Println("SPRING APP INIT")
}

func Run() {
	var r *mux.Router
	Container.Make(&r)
	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server started on port 8000")
	log.Fatal(srv.ListenAndServe())
}
