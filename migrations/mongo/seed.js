// WARUNGOS MongoDB: Seed Data
// Menu items, branches, and admin user

// Clear existing data
db.menu_items.deleteMany({});
db.branches_meta.deleteMany({});

const now = new Date();

const menuItems = [
    // Main Course - Nasi
    { name: "Nasi Goreng Spesial", description: "Nasi goreng dengan telur, ayam suwir, dan kerupuk", category: "main_course", price: 28000, spice_level: 2, tags: ["bestseller", "spicy"], preparation_time_minutes: 10 },
    { name: "Nasi Goreng Seafood", description: "Nasi goreng dengan udang, cumi, dan sayuran", category: "main_course", price: 35000, spice_level: 2, tags: ["premium"], preparation_time_minutes: 12 },
    { name: "Nasi Goreng Kampung", description: "Nasi goreng khas kampung dengan petai dan terasi", category: "main_course", price: 26000, spice_level: 3, tags: ["traditional"], preparation_time_minutes: 10 },
    { name: "Nasi Ayam Geprek", description: "Nasi dengan ayam geprek sambal matah", category: "main_course", price: 30000, spice_level: 4, tags: ["spicy", "bestseller"], preparation_time_minutes: 15 },
    { name: "Nasi Rendang", description: "Nasi dengan rendang sapi padang", category: "main_course", price: 38000, spice_level: 2, tags: ["premium", "traditional"], preparation_time_minutes: 8 },

    // Main Course - Mie
    { name: "Mie Ayam Bakso", description: "Mie ayam dengan bakso sapi dan pangsit goreng", category: "noodle", price: 22000, spice_level: 1, tags: ["bestseller"], preparation_time_minutes: 10 },
    { name: "Mie Goreng Jawa", description: "Mie goreng khas Jawa dengan kecap manis", category: "noodle", price: 22000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 8 },
    { name: "Mie Aceh Goreng", description: "Mie aceh goreng dengan bumbu kari rempah", category: "noodle", price: 28000, spice_level: 3, tags: ["spicy"], preparation_time_minutes: 12 },
    { name: "Mie Goreng Pedas Level 5", description: "Mie goreng ekstra pedas untuk pencinta pedas", category: "noodle", price: 25000, spice_level: 5, tags: ["extreme_spicy"], preparation_time_minutes: 10 },
    { name: "Bakmie Campur", description: "Bakmie ayam dan pangsit dengan pangsit rebus", category: "noodle", price: 24000, spice_level: 1, tags: [], preparation_time_minutes: 10 },

    // Main Course - Others
    { name: "Bakso Sapi Komplit", description: "Bakso sapi urat dengan mie dan bihun", category: "soup", price: 25000, spice_level: 1, tags: ["bestseller"], preparation_time_minutes: 8 },
    { name: "Soto Ayam Lamongan", description: "Soto ayam khas Lamongan dengan taburannya", category: "soup", price: 22000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 10 },
    { name: "Soto Betawi", description: "Soto Betawi kuah santan dengan daging dan jeroan", category: "soup", price: 28000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 8 },
    { name: "Rawon Daging", description: "Rawon daging sapi dengan kluwek dan kecambah", category: "soup", price: 30000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 8 },
    { name: "Ayam Bakar Madu", description: "Ayam bakar glasir madu dengan sambal", category: "main_course", price: 32000, spice_level: 2, tags: [], preparation_time_minutes: 12 },

    // Snack
    { name: "Sate Ayam 10 Tusuk", description: "Sate ayam dengan bumbu kacang", category: "snack", price: 20000, spice_level: 1, tags: ["bestseller"], preparation_time_minutes: 15 },
    { name: "Martabak Telur", description: "Martabak telur dengan daging cincang", category: "snack", price: 18000, spice_level: 1, tags: [], preparation_time_minutes: 12 },
    { name: "Tahu Tek", description: "Tahu goreng dengan petis dan bumbu kacang", category: "snack", price: 15000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 5 },
    { name: "Tempe Goreng", description: "Tempe goreng renyah dengan sambal kecap", category: "snack", price: 8000, spice_level: 0, tags: [], preparation_time_minutes: 5 },
    { name: "Pisang Goreng Crispy", description: "Pisang goreng tepung renyah", category: "snack", price: 12000, spice_level: 0, tags: ["sweet"], preparation_time_minutes: 8 },

    // Rice dishes
    { name: "Nasi Uduk Komplit", description: "Nasi uduk dengan ayam goreng, tempe, tahu, dan sayur", category: "rice", price: 25000, spice_level: 1, tags: ["traditional"], preparation_time_minutes: 5 },
    { name: "Nasi Kucing", description: "Nasi kecil dengan lauk pilihan (sambal teri, tempe orek, ayam)", category: "rice", price: 10000, spice_level: 1, tags: ["cheap"], preparation_time_minutes: 5 },
    { name: "Nasi Padang Sederhana", description: "Nasi putih dengan gulai ayam, daun singkong, sambal", category: "rice", price: 20000, spice_level: 2, tags: ["traditional"], preparation_time_minutes: 5 },

    // Side dishes
    { name: "Telur Dadar Padang", description: "Telur dadar tebal khas Padang", category: "side_dish", price: 10000, spice_level: 1, tags: [], preparation_time_minutes: 5 },
    { name: "Perkedel Kentang", description: "Perkedel kentang goreng", category: "side_dish", price: 8000, spice_level: 0, tags: [], preparation_time_minutes: 5 },

    // Drinks
    { name: "Es Teh Manis", description: "Es teh manis segar", category: "drink", price: 5000, spice_level: 0, tags: ["bestseller"], preparation_time_minutes: 2 },
    { name: "Es Jeruk Segar", description: "Es jeruk peras segar", category: "drink", price: 8000, spice_level: 0, tags: [], preparation_time_minutes: 3 },
    { name: "Es Kopi Susu", description: "Es kopi susu dengan gula aren", category: "drink", price: 15000, spice_level: 0, tags: ["trendy"], preparation_time_minutes: 3 },
    { name: "Es Cendol", description: "Es cendol dengan santan dan gula merah", category: "drink", price: 10000, spice_level: 0, tags: ["traditional"], preparation_time_minutes: 3 },
    { name: "Jus Alpukat", description: "Jus alpukat dengan susu coklat", category: "drink", price: 15000, spice_level: 0, tags: ["healthy"], preparation_time_minutes: 5 },
    { name: "Wedang Jahe", description: "Wedang jahe hangat dengan gula merah", category: "drink", price: 8000, spice_level: 0, tags: ["traditional", "warm"], preparation_time_minutes: 3 },

    // Dessert
    { name: "Es Krim Campina", description: "Es krim Campina 2 scoop pilihan rasa", category: "dessert", price: 12000, spice_level: 0, tags: [], preparation_time_minutes: 2 },
    { name: "Puding Coklat", description: "Puding coklat dengan vla vanila", category: "dessert", price: 10000, spice_level: 0, tags: [], preparation_time_minutes: 2 },
    { name: "Kolak Pisang", description: "Kolak pisang dengan ubi dan ketan hitam", category: "dessert", price: 10000, spice_level: 0, tags: ["traditional", "warm"], preparation_time_minutes: 3 },
    { name: "Bubur Sumsum", description: "Bubur sumsum dengan kinca gula merah", category: "dessert", price: 12000, spice_level: 0, tags: ["traditional"], preparation_time_minutes: 3 }
];

// Insert all menu items
const result = db.menu_items.insertMany(
    menuItems.map(item => ({
        ...item,
        image_url: "",
        is_available: true,
        is_active: true,
        branch_ids: [],
        created_at: now,
        updated_at: now
    }))
);

print(`Inserted ${result.insertedCount} menu items`);

// Create branches_meta collection for MongoDB
db.createCollection("branches_meta");

db.branches_meta.insertMany([
    {
        _id: "BRANCH-001",
        name: "WarungOS Pusat - Jakarta Selatan",
        address: "Jl. Kemang Raya No. 12, Jakarta Selatan",
        city: "Jakarta",
        province: "DKI Jakarta",
        phone: "+62 21 1234 5678",
        location: { type: "Point", coordinates: [-6.2607, 106.8106] },
        is_active: true,
        open_time: "07:00",
        close_time: "22:00",
        created_at: now,
        updated_at: now
    },
    {
        _id: "BRANCH-002",
        name: "WarungOS Bandung - Dago",
        address: "Jl. Dago No. 45, Bandung",
        city: "Bandung",
        province: "Jawa Barat",
        phone: "+62 22 8765 4321",
        location: { type: "Point", coordinates: [-6.8896, 107.6170] },
        is_active: true,
        open_time: "08:00",
        close_time: "21:00",
        created_at: now,
        updated_at: now
    },
    {
        _id: "BRANCH-003",
        name: "WarungOS Surabaya - Pemuda",
        address: "Jl. Pemuda No. 78, Surabaya",
        city: "Surabaya",
        province: "Jawa Timur",
        phone: "+62 31 1122 3344",
        location: { type: "Point", coordinates: [-7.2575, 112.7521] },
        is_active: true,
        open_time: "07:00",
        close_time: "22:00",
        created_at: now,
        updated_at: now
    }
]);

print("Inserted 3 branches");

// Note: Admin user is created in PostgreSQL via the auth-service
// This is just a reference document for MongoDB-side lookups
print("Seed complete. Note: Admin user (admin@warungos.id / admin123) must be created via auth-service API.");
