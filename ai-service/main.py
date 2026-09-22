"""
WarungOS AI Service — FastAPI sidecar for smart features.
Provides menu recommendations, sales forecasting, and inventory suggestions.
"""
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from datetime import datetime, timedelta
import random
import math

app = FastAPI(title="WarungOS AI Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


class Recommendation(BaseModel):
    menu_item_id: str
    name: str
    score: float
    reason: str


class ForecastDay(BaseModel):
    date: str
    predicted_revenue: float
    confidence_low: float
    confidence_high: float
    predicted_orders: int


class InventorySuggestion(BaseModel):
    item_name: str
    current_stock: float
    consumption_rate: float
    suggested_reorder: float
    days_until_stockout: int


@app.get("/health")
async def health():
    return {"status": "ok", "service": "ai-service"}


@app.get("/api/v1/ai/recommend/{branch_id}", response_model=list[Recommendation])
async def recommend_menu(branch_id: str, time_of_day: str = "afternoon"):
    """
    Menu recommendation engine using time-of-day + popularity patterns.

    Simulates a real recommendation system:
    - Morning (6-10): breakfast items, coffee
    - Afternoon (11-14): lunch items, rice dishes
    - Evening (17-20): dinner, heavier meals
    - Late night: snacks, drinks
    """
    # Time-based recommendation profiles
    time_profiles = {
        "morning": [
            ("Bubur Ayam", 0.95, "Sarapan populer di pagi hari"),
            ("Kopi Susu", 0.92, "Minuman favorit pagi"),
            ("Roti Bakar", 0.88, "Cocok untuk sarapan ringan"),
            ("Nasi Uduk", 0.85, "Sarapan berat"),
            ("Teh Hangat", 0.82, "Minuman hangat pagi hari"),
        ],
        "afternoon": [
            ("Nasi Goreng Spesial", 0.98, "Menu utama paling populer"),
            ("Mie Ayam Jamur", 0.94, "Pilihan populer makan siang"),
            ("Es Teh Manis", 0.91, "Minuman wajib makan siang"),
            ("Soto Ayam", 0.87, "Pilihan ringan tapi mengenyangkan"),
            ("Ayam Goreng", 0.85, "Favorit banyak pelanggan"),
        ],
        "evening": [
            ("Rendang Padang", 0.96, "Menu spesial malam hari"),
            ("Nasi Padang", 0.93, "Pilihan lengkap untuk makan malam"),
            ("Gado-Gado", 0.88, "Salad sehat malam hari"),
            ("Kopi Susu", 0.85, "Temani santai malam"),
            ("Pisang Goreng", 0.80, "Camilan penutup"),
        ],
        "latenight": [
            ("Martabak Mini", 0.90, "Camilan larut malam"),
            ("Es Krim Vanilla", 0.85, "Pencuci mulut"),
            ("Kopi Susu", 0.82, "Tetap terjaga"),
            ("Risoles", 0.78, "Camilan ringan"),
            ("Jus Alpukat", 0.75, "Minuman segar"),
        ],
    }

    # Determine time period
    hour = datetime.now().hour
    if 6 <= hour < 10:
        period = "morning"
    elif 11 <= hour < 15:
        period = "afternoon"
    elif 17 <= hour < 21:
        period = "evening"
    else:
        period = "latenight"

    # Override with parameter if provided
    if time_of_day in time_profiles:
        period = time_of_day

    items = time_profiles[period]
    # Add slight randomization for realism
    recommendations = []
    for name, score, reason in items:
        jitter = random.uniform(-0.03, 0.03)
        recommendations.append(Recommendation(
            menu_item_id=f"menu-{name.lower().replace(' ', '-')}",
            name=name,
            score=round(min(1.0, max(0.0, score + jitter)), 2),
            reason=reason,
        ))

    return sorted(recommendations, key=lambda x: x.score, reverse=True)


@app.get("/api/v1/ai/forecast/{branch_id}", response_model=list[ForecastDay])
async def forecast_sales(branch_id: str, days: int = 7):
    """
    Sales forecasting using simple moving average with seasonal adjustment.

    Real implementation would use ARIMA/Prophet. This demonstrates the pattern:
    - Pull historical data from analytics DB
    - Calculate base trend (7-day moving average)
    - Apply day-of-week seasonal factors
    - Add confidence intervals
    """
    # Simulated historical base (would come from MongoDB analytics)
    base_daily_revenue = 850_000  # Rp 850K average daily
    base_daily_orders = 45

    # Day-of-week seasonal factors (weekend = higher)
    dow_factors = {
        0: 0.85,  # Monday
        1: 0.90,  # Tuesday
        2: 0.95,  # Wednesday
        3: 1.00,  # Thursday
        4: 1.15,  # Friday
        5: 1.25,  # Saturday
        6: 1.10,  # Sunday
    }

    # Slight upward trend (+2% per week)
    trend_factor = 1.02

    forecasts = []
    for i in range(days):
        future_date = datetime.now() + timedelta(days=i + 1)
        dow = future_date.weekday()
        seasonal = dow_factors.get(dow, 1.0)
        trend = trend_factor ** (i / 7)

        predicted_revenue = base_daily_revenue * seasonal * trend
        predicted_orders = int(base_daily_orders * seasonal * trend)

        # Confidence interval widens with distance
        uncertainty = 0.10 + (i * 0.02)  # 10% base + 2% per day
        confidence_low = predicted_revenue * (1 - uncertainty)
        confidence_high = predicted_revenue * (1 + uncertainty)

        forecasts.append(ForecastDay(
            date=future_date.strftime("%Y-%m-%d"),
            predicted_revenue=round(predicted_revenue),
            confidence_low=round(confidence_low),
            confidence_high=round(confidence_high),
            predicted_orders=predicted_orders,
        ))

    return forecasts


@app.get("/api/v1/ai/inventory-suggest/{branch_id}", response_model=list[InventorySuggestion])
async def suggest_inventory(branch_id: str):
    """
    Smart inventory reorder suggestions based on consumption patterns.

    Analyzes:
    - Historical consumption rate from order data
    - Current stock level
    - Lead time for restocking (simulated 2 days)
    """
    # Simulated current inventory (would come from PostgreSQL)
    inventory_data = [
        {"name": "Tepung Terigu", "stock": 15.0, "unit": "kg", "daily_rate": 8.0},
        {"name": "Beras Premium", "stock": 50.0, "unit": "kg", "daily_rate": 12.0},
        {"name": "Minyak Goreng", "stock": 20.0, "unit": "liter", "daily_rate": 5.0},
        {"name": "Ayam Segar", "stock": 10.0, "unit": "kg", "daily_rate": 7.0},
        {"name": "Es Teh", "stock": 100.0, "unit": "pcs", "daily_rate": 35.0},
        {"name": "Kopi Arabika", "stock": 2.0, "unit": "kg", "daily_rate": 1.5},
        {"name": "Gula Pasir", "stock": 8.0, "unit": "kg", "daily_rate": 3.0},
        {"name": "Telur", "stock": 60.0, "unit": "pcs", "daily_rate": 25.0},
    ]

    suggestions = []
    for item in inventory_data:
        days_until_stockout = max(0, int(item["stock"] / item["daily_rate"]))
        # Suggest reorder if less than 3 days of stock
        if days_until_stockout <= 3:
            suggested = item["daily_rate"] * 7  # Order 7 days worth
            suggestions.append(InventorySuggestion(
                item_name=item["name"],
                current_stock=item["stock"],
                consumption_rate=item["daily_rate"],
                suggested_reorder=round(suggested, 1),
                days_until_stockout=days_until_stockout,
            ))

    return sorted(suggestions, key=lambda x: x.days_until_stockout)
