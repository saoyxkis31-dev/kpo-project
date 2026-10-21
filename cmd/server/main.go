package main

import (
	"Confetki/internal/database"
	"Confetki/internal/handler"
	"Confetki/internal/middleware"
	"Confetki/internal/repository"
	"Confetki/internal/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()

	connString := os.Getenv("CONN_STRING")

	pool, err := database.Connect(ctx, connString)
	if err != nil {
		fmt.Println("database failed connection: ", err)
		return
	}
	defer pool.Close()
	fmt.Println("database connection")

	productRepository := repository.NewProductRepository(pool)
	masterClassRepository := repository.NewMasterClassRepository(pool)
	productService := service.NewProductService(productRepository)
	masterClassService := service.NewMasterClassService(masterClassRepository)
	orderRepository := repository.NewOrderRepository(pool)
	orderService := service.NewOrderService(orderRepository, productRepository)
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewServiceAuth(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	appHandler := handler.NewHandler(productService, masterClassService, orderService)

	r := chi.NewRouter()
	r.Get("/", handler.Home)

	r.Route("/api", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Get("/", appHandler.GetProducts)
			r.Get("/{id}", appHandler.GetProductById)
			r.Get("/{id}/master-class", appHandler.GetMasterClass)
			r.Post("/", appHandler.CreateProduct)
			r.Delete("/{id}", appHandler.DeleteProduct)
		})

		r.Post("/master-classes", appHandler.CreateMasterClass)

		r.Route("/orders", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Get("/", appHandler.GetAllOrders)
			r.Post("/", appHandler.CreateOrder)
			r.Get("/{id}", appHandler.GetOrder)
			r.Patch("/{id}/status", appHandler.UpdateStatus)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server is starting on port 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefullyc")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	pool.Close()
	log.Println("Server exited properly")
}
