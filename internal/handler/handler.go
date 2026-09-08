package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)


func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, " server is running")
}


func GetProducts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "список продуктов")
}


func GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fmt.Fprintln(w, "товар с id: ", id)
}


func GetMasterClass(w http.ResponseWriter, r *http.Request) {
	ProductID := chi.URLParam(r, "id")

	fmt.Fprintln(w, "мастеркласс товара с id: ", ProductID)
}