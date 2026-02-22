package main

import (
	"log"
	"net/http"
	"recipe-api/config"
	"recipe-api/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize MongoDB
	config.ConnectDB()

	r := mux.NewRouter()

	// Root Route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Welcome to the Recipe API! Use /recipes to access the data."}`))
	}).Methods("GET")

	// Recipe Endpoints
	r.HandleFunc("/recipes", handlers.CreateRecipe).Methods("POST")
	r.HandleFunc("/recipes", handlers.GetAllRecipes).Methods("GET")
	r.HandleFunc("/recipes/top", handlers.GetTopRecipes).Methods("GET")    // Must be before /recipes/{id} to avoid regex matching conflict
	r.HandleFunc("/recipes/search", handlers.SearchRecipes).Methods("GET") // Same here
	r.HandleFunc("/recipes/{id}", handlers.GetRecipeByID).Methods("GET")
	r.HandleFunc("/recipes/{id}/rate", handlers.RateRecipe).Methods("POST")

	// Serve uploaded images statically
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	log.Println("Server started on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
