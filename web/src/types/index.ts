// ============================================
// Type definitions for WarungOS
// ============================================

// Auth
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'owner' | 'manager' | 'cashier';
  branch_id: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  role: string;
  branch_id: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

// Menu
export interface Variant {
  name: string;
  price_modifier: number;
}

export interface Modifier {
  name: string;
  price: number;
}

export interface MenuItem {
  id: string;
  name: string;
  description: string;
  category: string;
  price: number;
  image_url: string;
  variants: Variant[];
  modifiers: Modifier[];
  branch_ids: string[];
  is_available: boolean;
  tags: string[];
  created_at: string;
}

// Orders
export type OrderStatus = 'pending' | 'confirmed' | 'preparing' | 'ready' | 'completed' | 'cancelled';
export type OrderType = 'dine_in' | 'takeaway' | 'delivery';

export interface OrderItem {
  id: string;
  order_id: string;
  menu_item_id: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  notes: string;
}

export interface Order {
  id: string;
  branch_id: string;
  cashier_id: string;
  customer_name: string;
  order_type: OrderType;
  status: OrderStatus;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  payment_method: string;
  payment_status: string;
  notes: string;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

// Inventory
export interface InventoryItem {
  id: string;
  branch_id: string;
  item_name: string;
  quantity: number;
  unit: string;
  min_stock: number;
  cost_per_unit: number;
  updated_at: string;
}

// Cart (POS specific)
export interface CartItem {
  menu_item: MenuItem;
  quantity: number;
  selected_variant?: string;
  selected_modifiers: string[];
  notes: string;
}

// Dashboard
export interface CategoryStats {
  category: string;
  count: number;
  avg_price: number;
}

export interface SalesData {
  date: string;
  revenue: number;
  orders: number;
}

// API Response
export interface APIResponse<T> {
  success: boolean;
  data: T;
  error?: string;
  meta?: {
    page: number;
    limit: number;
    total: number;
  };
}
