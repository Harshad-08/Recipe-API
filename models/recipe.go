package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Recipe struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Ingredients []string           `bson:"ingredients" json:"ingredients"`
	ImagePath   string             `bson:"image_path" json:"image_path"`
	Ratings     []int              `bson:"ratings" json:"ratings"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type RatingInput struct {
	Rating int `json:"rating"`
}
