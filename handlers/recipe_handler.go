package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"recipe-api/config"
	"recipe-api/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/image/draw"
)

var recipeCache sync.Map

// Helper for JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Helper for Error responses
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Resize image to max width 800 and compress
func processImage(file io.Reader, filename string) (string, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > 800 {
		ratio := float64(800) / float64(width)
		width = 800
		height = int(float64(height) * ratio)
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	os.MkdirAll("uploads", os.ModePerm)

	outFilename := fmt.Sprintf("%d_%s.jpg", time.Now().UnixNano(), strings.TrimSuffix(filename, filepath.Ext(filename)))
	outPath := filepath.Join("uploads", strings.ReplaceAll(outFilename, " ", "_"))
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	var opt jpeg.Options
	opt.Quality = 75
	err = jpeg.Encode(out, dst, &opt)
	if err != nil {
		return "", err
	}

	return "/" + outPath, nil
}

// 1. Create Recipe
func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	// 2MB max
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "File size exceeds 2MB or invalid multipart form")
		return
	}

	title := strings.TrimSpace(strings.Trim(r.FormValue("title"), `"' \`))
	description := strings.TrimSpace(strings.Trim(r.FormValue("description"), `"' \`))
	ingredientsStr := strings.TrimSpace(strings.Trim(r.FormValue("ingredients"), `"' \`))

	if title == "" || description == "" || ingredientsStr == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	var ingredients []string
	for _, ing := range strings.Split(ingredientsStr, ",") {
		trimmed := strings.TrimSpace(strings.Trim(ing, `"' \`))
		if trimmed != "" {
			ingredients = append(ingredients, trimmed)
		}
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Image file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		respondError(w, http.StatusBadRequest, "Only jpg/png allowed")
		return
	}

	imagePath, err := processImage(file, header.Filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process image")
		return
	}

	recipe := models.Recipe{
		Title:       title,
		Description: description,
		Ingredients: ingredients,
		ImagePath:   imagePath,
		Ratings:     []int{},
		CreatedAt:   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := config.RecipeCollection.InsertOne(ctx, recipe)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save recipe")
		return
	}

	recipe.ID = result.InsertedID.(primitive.ObjectID)
	respondJSON(w, http.StatusCreated, recipe)
}

// 2. Get All Recipes
func GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use aggregation to dynamically compute average rating for all recipes too
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "avg_rating", Value: bson.D{{Key: "$avg", Value: "$ratings"}}}}}},
	}

	cursor, err := config.RecipeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}
	defer cursor.Close(ctx)

	var recipes []bson.M
	if err = cursor.All(ctx, &recipes); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode recipes")
		return
	}

	if recipes == nil {
		recipes = []bson.M{}
	}

	respondJSON(w, http.StatusOK, recipes)
}

// 3. Get Recipe by ID with in-memory caching
func GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if val, ok := recipeCache.Load(id); ok {
		respondJSON(w, http.StatusOK, val)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: objID}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "avg_rating", Value: bson.D{{Key: "$avg", Value: "$ratings"}}}}}},
	}

	cursor, err := config.RecipeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		respondError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	recipe := results[0]
	recipeCache.Store(id, recipe)
	respondJSON(w, http.StatusOK, recipe)
}

// 4. Multi-Ingredient Search
func SearchRecipes(w http.ResponseWriter, r *http.Request) {
	ingredientsStr := r.URL.Query().Get("ingredients")
	if ingredientsStr == "" {
		respondError(w, http.StatusBadRequest, "Missing ingredients parameter")
		return
	}

	var ingredients []string
	for _, ing := range strings.Split(ingredientsStr, ",") {
		trimmed := strings.TrimSpace(ing)
		if trimmed != "" {
			// support exact match logic or partial match depending on indexing,
			// $all applies to arrays smoothly if exact items exist
			ingredients = append(ingredients, trimmed)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Dynamically compute average rating during search as well
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "ingredients", Value: bson.D{{Key: "$all", Value: ingredients}}}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "avg_rating", Value: bson.D{{Key: "$avg", Value: "$ratings"}}}}}},
	}

	cursor, err := config.RecipeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Search failed")
		return
	}
	defer cursor.Close(ctx)

	var recipes []bson.M
	if err = cursor.All(ctx, &recipes); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode recipes")
		return
	}

	if recipes == nil {
		recipes = []bson.M{}
	}

	respondJSON(w, http.StatusOK, recipes)
}

// 5. Rate Recipe
func RateRecipe(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var input models.RatingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.Rating < 1 || input.Rating > 5 {
		respondError(w, http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$push": bson.M{"ratings": input.Rating}}
	result, err := config.RecipeCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add rating")
		return
	}
	if result.MatchedCount == 0 {
		respondError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Invalidate cache immediately when rated
	recipeCache.Delete(id)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Rating added successfully"})
}

// 6 & 7. Top Rated Recipes with Dynamic Average Rating
func GetTopRecipes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "avg_rating", Value: bson.D{{Key: "$avg", Value: "$ratings"}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "avg_rating", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 5}},
	}

	cursor, err := config.RecipeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Aggregation failed")
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode top recipes")
		return
	}

	if results == nil {
		results = []bson.M{}
	}

	respondJSON(w, http.StatusOK, results)
}
