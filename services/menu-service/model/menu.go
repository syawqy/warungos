package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Modifier represents an add-on or option group for a menu item.
type Modifier struct {
	Name   string  `bson:"name" json:"name"`
	Price  float64 `bson:"price" json:"price"`
	IsFree bool    `bson:"is_free" json:"is_free"`
}

// Variant represents a size or variant of a menu item.
type Variant struct {
	Name       string  `bson:"name" json:"name"`
	Price      float64 `bson:"price" json:"price"`
	SKU        string  `bson:"sku" json:"sku"`
	IsDefault  bool    `bson:"is_default" json:"is_default"`
	StockCount int     `bson:"stock_count" json:"stock_count"`
}

// MenuItem is the core domain object for menu entries.
type MenuItem struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name" validate:"required"`
	Description string        `bson:"description" json:"description"`
	Category    string        `bson:"category" json:"category" validate:"required"`
	Price       float64       `bson:"price" json:"price" validate:"required"`
	ImageURL    string        `bson:"image_url" json:"image_url"`
	IsAvailable bool          `bson:"is_available" json:"is_available"`
	Variants    []Variant     `bson:"variants,omitempty" json:"variants,omitempty"`
	Modifiers   []Modifier    `bson:"modifiers,omitempty" json:"modifiers,omitempty"`
	BranchIDs   []string      `bson:"branch_ids,omitempty" json:"branch_ids,omitempty"`
	Tags        []string      `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

// CategoryStats represents aggregated statistics for a menu category.
type CategoryStats struct {
	Category   string  `bson:"category" json:"category"`
	TotalItems int     `bson:"total_items" json:"total_items"`
	AvgPrice   float64 `bson:"avg_price" json:"avg_price"`
	MinPrice   float64 `bson:"min_price" json:"min_price"`
	MaxPrice   float64 `bson:"max_price" json:"max_price"`
}
