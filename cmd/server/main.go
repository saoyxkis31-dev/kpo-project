package main

import (
	_ "Confetki/docs"
	"Confetki/internal/cashe"
	"Confetki/internal/database"
	"Confetki/internal/handler"
	"Confetki/internal/middleware"
	"Confetki/internal/repository"
	"Confetki/internal/service"
	"Confetki/pkg/auth"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

// @title Confetki API
// @version 1.0
// @description API бэкенда для магазина десертов
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// читаем файл .env и загружаем переменные в систему
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, берутся переменные системы")
	}

	ctx := context.Background()
	connString := os.Getenv("CONN_STRING")
	if connString == "" {
		log.Fatal("CONN_STRING IS EMPTY")
	}
	auth.Init()

	pool, err := database.Connect(ctx, connString)
	if err != nil {
		fmt.Println("database failed connection: ", err)
		return
	}
	defer pool.Close()
	fmt.Println("database connection")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb, err := cashe.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatal("failed to connect redis: %v", err)
	}
	defer rdb.Close()
	log.Println("connected to redis successfully")

	productRepository := repository.NewProductRepository(pool)
	masterClassRepository := repository.NewMasterClassRepository(pool)
	productService := service.NewProductService(productRepository, rdb)
	masterClassService := service.NewMasterClassService(masterClassRepository)
	orderRepository := repository.NewOrderRepository(pool)
	orderService := service.NewOrderService(orderRepository, productRepository)
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewServiceAuth(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	appHandler := handler.NewHandler(productService, masterClassService, orderService)

	r := chi.NewRouter()
	r.Get("/", handler.Home)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Get("/", appHandler.GetProducts)
			r.Get("/{id}", appHandler.GetProductById)
			r.Get("/{id}/master-class", appHandler.GetMasterClass)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware)
				r.Use(middleware.AdminMiddleware)

				r.Post("/", appHandler.CreateProduct)
				r.Delete("/{id}", appHandler.DeleteProduct)
			})
		})

		r.Route("/master-classes", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.AdminMiddleware)

			r.Post("/", appHandler.CreateMasterClass)
		})

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

	// что тут ваще дальше происходит подробно разобрать *93202422444444444444444444444444444444444444444444444444444444444444444424242
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
