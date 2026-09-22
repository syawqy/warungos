package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/warungos/menu-service/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MenuRepository defines the interface for menu data access.
type MenuRepository interface {
	Create(ctx context.Context, item *model.MenuItem) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.MenuItem, error)
	GetAll(ctx context.Context, branchID string, category string, search string) ([]model.MenuItem, error)
	Update(ctx context.Context, id bson.ObjectID, item *model.MenuItem) error
	Delete(ctx context.Context, id bson.ObjectID) error
	GetCategoryStats(ctx context.Context, branchID string) ([]model.CategoryStats, error)
}

// MongoMenuRepository implements MenuRepository using MongoDB.
type MongoMenuRepository struct {
	collection *mongo.Collection
}

// NewMongoMenuRepository creates a new MongoDB-backed menu repository.
func NewMongoMenuRepository(db *mongo.Database) *MongoMenuRepository {
	return &MongoMenuRepository{
		collection: db.Collection("menu_items"),
	}
}

// Create inserts a new menu item into the database.
func (r *MongoMenuRepository) Create(ctx context.Context, item *model.MenuItem) error {
	item.ID = bson.NewObjectID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to create menu item: %w", err)
	}
	return nil
}

// GetByID retrieves a menu item by its ObjectID.
func (r *MongoMenuRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.MenuItem, error) {
	var item model.MenuItem
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("menu item not found: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get menu item: %w", err)
	}
	return &item, nil
}

// GetAll retrieves menu items with optional filtering by branch, category, and text search.
func (r *MongoMenuRepository) GetAll(ctx context.Context, branchID string, category string, search string) ([]model.MenuItem, error) {
	filter := bson.M{}

	if branchID != "" {
		filter["branch_ids"] = branchID
	}
	if category != "" {
		filter["category"] = category
	}
	if search != "" {
		filter["$text"] = bson.M{"$search": search}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list menu items: %w", err)
	}
	defer cursor.Close(ctx)

	var items []model.MenuItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to decode menu items: %w", err)
	}
	return items, nil
}

// Update modifies an existing menu item by its ObjectID.
func (r *MongoMenuRepository) Update(ctx context.Context, id bson.ObjectID, item *model.MenuItem) error {
	item.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":         item.Name,
			"description":  item.Description,
			"category":     item.Category,
			"price":        item.Price,
			"image_url":    item.ImageURL,
			"is_available": item.IsAvailable,
			"variants":     item.Variants,
			"modifiers":    item.Modifiers,
			"branch_ids":   item.BranchIDs,
			"tags":         item.Tags,
			"updated_at":   item.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update menu item: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("menu item not found: %s", id.Hex())
	}
	return nil
}

// Delete removes a menu item by its ObjectID.
func (r *MongoMenuRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete menu item: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("menu item not found: %s", id.Hex())
	}
	return nil
}

// GetCategoryStats runs an aggregation pipeline to compute per-category statistics.
func (r *MongoMenuRepository) GetCategoryStats(ctx context.Context, branchID string) ([]model.CategoryStats, error) {
	matchStage := bson.M{}
	if branchID != "" {
		matchStage["branch_ids"] = branchID
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$category",
			"total_items": bson.M{"$sum": 1},
			"avg_price":   bson.M{"$avg": "$price"},
			"min_price":   bson.M{"$min": "$price"},
			"max_price":   bson.M{"$max": "$price"},
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":         0,
			"category":    "$_id",
			"total_items": 1,
			"avg_price":   bson.M{"$round": []interface{}{"$avg_price", 2}},
			"min_price":   1,
			"max_price":   1,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "category", Value: 1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate category stats: %w", err)
	}
	defer cursor.Close(ctx)

	var stats []model.CategoryStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, fmt.Errorf("failed to decode category stats: %w", err)
	}
	return stats, nil
}
