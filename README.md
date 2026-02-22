🍽️ Recipe Sharing API (Go + MongoDB)

A production-ready REST API built in Go for a Recipe Sharing Platform.

Users can upload recipes with images, search by multiple ingredients, rate recipes, and retrieve top-rated results.

Beyond basic CRUD functionality, this implementation emphasizes backend engineering best practices, including:

🔎 Multi-ingredient search using MongoDB $all

📊 Aggregation pipelines for dynamic average rating computation

⚡ Lightweight in-memory caching for performance optimization

🗂 Indexed queries for faster ingredient lookup

🖼 Secure image validation, resizing, and compression

🧱 Clean modular architecture with separation of concerns

This project is backend-only and frontend-agnostic by design.

🚀 Tech Stack

Go 1.21+

net/http

Gorilla Mux

MongoDB (Official Go Driver)

Local image storage (/uploads)

In-memory caching using sync.Map

🏗️ Architecture Overview

The project follows a clean, modular structure:

recipe-api/
│
├── main.go
├── go.mod
│
├── config/
│     mongo.go
│
├── models/
│     recipe.go
│
├── handlers/
│     recipe_handler.go
│
├── uploads/
│
├── README.md
