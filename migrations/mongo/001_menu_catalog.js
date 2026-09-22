// WARUNGOS MongoDB: Menu Catalog Collection
// Migration 001: Create menu_items collection with JSON schema validation

db.createCollection("menu_items", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["name", "description", "category", "price", "is_available"],
            properties: {
                name: {
                    bsonType: "string",
                    description: "Menu item name"
                },
                description: {
                    bsonType: "string",
                    description: "Menu item description"
                },
                category: {
                    bsonType: "string",
                    enum: ["main_course", "snack", "drink", "dessert", "rice", "noodle", "soup", "side_dish"],
                    description: "Menu category"
                },
                price: {
                    bsonType: "long",
                    minimum: 0,
                    description: "Price in IDR"
                },
                image_url: {
                    bsonType: "string"
                },
                is_available: {
                    bsonType: "bool",
                    description: "Whether the item is currently available"
                },
                is_active: {
                    bsonType: "bool"
                },
                tags: {
                    bsonType: "array",
                    items: { bsonType: "string" }
                },
                spice_level: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 5
                },
                preparation_time_minutes: {
                    bsonType: "int",
                    minimum: 1
                },
                branch_ids: {
                    bsonType: "array",
                    items: { bsonType: "string" },
                    description: "Branch IDs where item is available (empty = all branches)"
                },
                created_at: {
                    bsonType: "date"
                },
                updated_at: {
                    bsonType: "date"
                }
            }
        }
    }
});

// Text index for search
db.menu_items.createIndex(
    { name: "text", description: "text" },
    { weights: { name: 10, description: 1 }, name: "menu_text_search" }
);

// Additional indexes
db.menu_items.createIndex({ category: 1 });
db.menu_items.createIndex({ price: 1 });
db.menu_items.createIndex({ is_available: 1, is_active: 1 });
db.menu_items.createIndex({ tags: 1 });
