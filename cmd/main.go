package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"numerisTask/api"
	"numerisTask/models"
	"os"
)

func registerAPI(router *chi.Mux) {
	router.Route("/api/v1", func(apiRouter chi.Router) {
		//Invoice API
		apiRouter.Get("/invoices", api.GetInvoices)
		apiRouter.Get("/invoice/{invoiceId}", api.GetInvoiceByInvoiceId)

	})
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading environment variables")
	}

	_, err = models.Init()

	if err != nil {
		fmt.Println(err)
		log.Fatal("Error connecting to database ...")
	}

	//Base Router
	router := chi.NewRouter()
	//Router Middleware mount LOGGER
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.AllowContentType("application/json"))

	//NOTFOUND HANDLER
	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		notFoundResponse := map[string]string{"detail": "not found"}
		jsonResponse, _ := json.Marshal(notFoundResponse)
		writer.Write(jsonResponse)
		// Log the 404 error manually
		log.Printf("404 - Not Found: %s %s\n", request.Method, request.RequestURI)
	})

	//NOT FOUND METHOD NOT ALLOWED
	router.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusMethodNotAllowed)
		notMethodResponse := map[string]string{"detail": "method not allowed"}
		jsonResponse, _ := json.Marshal(notMethodResponse)
		writer.Write(jsonResponse)
		// Log the 405 error manually
		log.Printf("405 - Method Not Allowed: %s %s\n", request.Method, request.RequestURI)
	})

	port := os.Getenv("PORT")

	log.Printf("Listening on port :%s", port)

	//REGISTERING APIs and mounting it on the base Router
	registerAPI(router)
	http.ListenAndServe(fmt.Sprintf(":%v", port), router)
}
