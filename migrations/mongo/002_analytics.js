// WARUNGOS MongoDB: Analytics Collections
// Migration 002: order_analytics (timeseries) + ai_insights

// Timeseries collection for order analytics
db.createCollection("order_analytics", {
    timeseries: {
        timeField: "timestamp",
        metaField: "metadata",
        granularity: "hours"
    },
    expireAfterSeconds: 31536000 // 1 year retention
});

db.order_analytics.createIndex({ "metadata.branch_id": 1, timestamp: -1 });
db.order_analytics.createIndex({ "metadata.status": 1, timestamp: -1 });

// AI insights collection for recommendations, demand forecasts, etc.
db.createCollection("ai_insights", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["type", "branch_id", "data", "generated_at"],
            properties: {
                type: {
                    bsonType: "string",
                    enum: ["demand_forecast", "menu_recommendation", "pricing_suggestion", "inventory_alert", "sales_pattern"]
                },
                branch_id: {
                    bsonType: "string"
                },
                data: {
                    bsonType: "object"
                },
                confidence: {
                    bsonType: "double",
                    minimum: 0,
                    maximum: 1
                },
                generated_at: {
                    bsonType: "date"
                },
                expires_at: {
                    bsonType: "date"
                }
            }
        }
    }
});

db.ai_insights.createIndex({ type: 1, branch_id: 1 });
db.ai_insights.createIndex({ generated_at: -1 });
db.ai_insights.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });
