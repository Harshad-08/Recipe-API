# 🍽️ Recipe Sharing API (Go + MongoDB)

A production-ready REST API built in Go for a Recipe Sharing Platform.

Users can upload recipes with images, search by multiple ingredients, rate recipes, and retrieve top-rated results.

Beyond basic CRUD functionality, this implementation emphasizes backend engineering best practices, including:

- Lightweight in-memory caching for performance optimization
- Multi-ingredient search using MongoDB `$all`
- Aggregation pipelines for dynamic average rating computation
- Indexed queries for faster ingredient lookup
- Secure image validation, resizing, and compression
- Clean modular architecture with separation of concerns

This project is backend-only and frontend-agnostic by design.

---

# 🚀 Tech Stack

- Go 1.21+
- net/http
- Gorilla Mux
- MongoDB (Official Go Driver)
- Local image storage (`uploads/`)
- In-memory caching using `sync.Map`

---

# 🏗️ Project Structure

```
recipe-api/
│
├── main.go
├── go.mod
│
├── config/
│     └── mongo.go
│
├── models/
│     └── recipe.go
│
├── handlers/
│     └── recipe_handler.go
│
├── uploads/
│
└── README.md
```

---

# 🗄️ MongoDB Schema

Each recipe document:

```json
{
  "_id": "ObjectId",
  "title": "String",
  "description": "String",
  "ingredients": ["String"],
  "image_path": "String",
  "ratings": [4, 5, 3],
  "created_at": "ISODate"
}
```

### Design Decisions

- Ingredients stored as array → Enables efficient `$all` search
- Ratings stored as integer array → Enables aggregation
- Index created on `ingredients` field → Optimizes search queries

---

# 🖼️ Image Processing & Security

Image uploads are handled securely and efficiently.

### Validation Rules

- Only JPEG or PNG files
- Maximum file size: 2MB

### Processing Pipeline

1. Decode uploaded image
2. Resize to maximum width 800px
3. Maintain aspect ratio
4. Re-encode as JPEG
5. Compress with quality = 75
6. Save to `/uploads` folder

This ensures:

- Reduced storage footprint
- Optimized response size
- Protection against malicious uploads

---

# ⚡ Performance Optimizations

## 1️⃣ In-Memory Caching

`GET /recipes/{id}` uses a thread-safe `sync.Map`.

Flow:

- Check cache
- On miss → query MongoDB
- Store result in cache
- On rating update → invalidate cache entry

This reduces repeated database lookups.

---

## 2️⃣ MongoDB Aggregation Pipeline

Average ratings are calculated using:

- `$avg`
- `$addFields`
- `$sort`
- `$limit`

This ensures:

- Computation happens at database level
- Efficient ranking
- Cleaner API responses

---

## 3️⃣ Indexed Search

At application startup, an index is created on:

`ingredients (ascending)`

This significantly improves multi-ingredient search performance.

---

# 🔎 Multi-Ingredient Search

Endpoint:

GET /recipes/search?ingredients=tomato,onion

Query logic:

```go
bson.M{
  "ingredients": bson.M{
      "$all": []string{"tomato", "onion"},
  },
}
```

Returns recipes containing all specified ingredients.

---

# 📦 Setup Instructions

## Prerequisites

- Go 1.21+
- MongoDB running locally

Default MongoDB URI:

mongodb://localhost:27017

---

## Environment Variables

| Variable   | Description                     | Default                        |
|------------|---------------------------------|--------------------------------|
| MONGO_URI  | MongoDB connection string       | mongodb://localhost:27017      |

---

## Run the Application

```bash
cd recipe-api
go mod tidy
go run main.go
```

Server runs at:

http://localhost:8081

---

# 📡 API Endpoints

## 1️⃣ Create Recipe

POST /recipes  
Content-Type: multipart/form-data  

Fields:

- title
- description
- ingredients (comma-separated)
- image (file)

Example:

```bash
curl -X POST http://localhost:8081/recipes \
  -F "title=Tomato Soup" \
  -F "description=Hot soup" \
  -F "ingredients=tomato,onion,garlic" \
  -F "image=@soup.jpg"
```

Response: 201 Created

---

## 2️⃣ Get All Recipes

GET /recipes  

Returns all recipes with dynamically computed average rating.

---

## 3️⃣ Get Recipe by ID

GET /recipes/{id}  

Uses in-memory cache for performance.

---

## 4️⃣ Search Recipes

GET /recipes/search?ingredients=tomato,onion  

Returns recipes containing all specified ingredients.

---

## 5️⃣ Rate Recipe

POST /recipes/{id}/rate  

Body:

```json
{
  "rating": 4
}
```

Validation:

- Rating must be between 1 and 5

Response: 200 OK

---

## 6️⃣ Top Rated Recipes

GET /recipes/top  

Returns top 5 recipes sorted by average rating.

---

# ❗ Error Handling

The API follows standard REST conventions:

- 201 → Created
- 400 → Bad Request
- 404 → Not Found
- 500 → Internal Server Error

Example:

```json
{
  "error": "Invalid rating value"
}
```

---

# 🔮 Future Improvements

- JWT-based authentication
- Cloud image storage (AWS S3 / GCP)
- Redis distributed caching
- Pagination support
- Rate limiting
- Full-text search with MongoDB Atlas
- Docker containerization

---

# 🏆 Engineering Highlights

This project demonstrates:

- RESTful API design
- Secure file handling
- Image processing in Go
- MongoDB aggregation pipelines
- Performance-aware indexing
- Lightweight caching strategy
- Clean modular architecture

Designed to be scalable, maintainable, and frontend-agnostic.
