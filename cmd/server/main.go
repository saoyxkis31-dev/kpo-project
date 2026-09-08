package main

import (
	"Confetki/internal/handler"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", handler.Home)

	r.Route("/api", func(r chi.Router) {
		r.Route("/products", func(r chi.Router){
			r.Get("/", handler.GetProducts)
			r.Get("/{id}", handler.GetProduct)
			r.Get("/{id}/master-class", handler.GetMasterClass)
		})

	})
	

	fmt.Printf("server is running\n")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Print(err)
	}
}
