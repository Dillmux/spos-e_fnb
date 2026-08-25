export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  errors?: any;
  error?: any;
}

export interface SetupStatus {
  initialized: boolean;
  db_connected?: boolean;
  outlet_name?: string;
  outlet?: Outlet;
  feature_flags?: FeatureFlags;
}

export interface SetupPayload {
  db_driver: 'sqlite' | 'postgres' | 'mysql';
  db_connection_string?: string;
  outlet_name: string;
  outlet_tagline?: string;
  outlet_address: string;
  outlet_phone: string;
  business_type: 'cafe' | 'resto' | 'cloud_kitchen';
  service_modes: ('dine_in' | 'take_away' | 'delivery')[];
  active_modules: {
    enable_tables: boolean;
    enable_kds: boolean;
    enable_wld: boolean;
    enable_bom: boolean;
    enable_midtrans: boolean;
  };
  tax_percentage: number;
  service_percentage: number;
  is_tax_inclusive: boolean;
  midtrans_server_key?: string;
  midtrans_client_key?: string;
  admin_name: string;
  admin_username: string;
  admin_password: string;
  admin_pin: string;
}

export interface Outlet {
  id: string;
  name: string;
  tagline?: string;
  address: string;
  phone: string;
  business_type: 'cafe' | 'resto' | 'cloud_kitchen';
  service_modes: ('dine_in' | 'take_away' | 'delivery')[];
  tax_percentage: number;
  service_percentage: number;
  is_tax_inclusive: boolean;
  receipt_footer?: string;
}

export interface User {
  id: string;
  outlet_id?: string;
  outlet?: Outlet;
  Outlet?: Outlet;
  name: string;
  username: string;
  email?: string;
  role: 'superadmin' | 'manager' | 'cashier' | 'kitchen';
  is_active: boolean;
  has_pin?: boolean;
}

export interface Category {
  id: number | string;
  outlet_id?: string;
  name: string;
  sort_order?: number;
}

export interface ProductRecipe {
  id?: number | string;
  product_id?: string;
  ingredient_id: string;
  ingredient_name?: string;
  quantity_required?: number;
  quantity?: number;
  unit?: string;
  Ingredient?: Ingredient;
}

export interface Product {
  id: string;
  category_id: number | string;
  category_name?: string;
  sku?: string;
  name: string;
  price: number;
  cogs?: number;
  is_available: boolean;
  Category?: Category;
  Recipes?: ProductRecipe[];
  recipes?: ProductRecipe[];
  recipe?: ProductRecipe[];
}

export interface Ingredient {
  id: string;
  outlet_id?: string;
  name: string;
  unit: string;
  current_stock?: number;
  stock?: number;
  min_stock_alert?: number;
  min_stock?: number;
  cost_per_unit?: number;
  cogs_per_unit?: number;
  updated_at?: string;
}

export interface Table {
  id: number | string;
  outlet_id?: string;
  table_number: string;
  name?: string;
  capacity: number;
  status: 'available' | 'occupied' | 'reserved';
  active_order_id?: string;
}

export interface CartItem {
  id: string; // generated unique key for cart row
  product: Product;
  quantity: number;
  notes: string;
  discount_amount: number;
  split_group?: 'A' | 'B';
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  product_name?: string;
  quantity: number;
  unit_price?: number;
  price?: number;
  subtotal: number;
  item_notes?: string;
  notes?: string;
  kitchen_status: 'queued' | 'cooking' | 'ready' | 'served' | 'pending' | 'preparing';
  Product?: Product;
  product?: Product;
}

export interface Order {
  id: string;
  invoice_number?: string;
  order_number?: string;
  parent_order_id?: string;
  parent_id?: string;
  table_id?: number | string;
  table_name?: string;
  table_number?: string;
  service_type: 'dine_in' | 'take_away' | 'delivery';
  service_mode?: 'dine_in' | 'take_away' | 'delivery';
  customer_name?: string;
  subtotal: number;
  tax_amount: number;
  service_amount: number;
  discount_amount: number;
  total_amount: number;
  queue_number?: number;
  status: 'open' | 'pending_payment' | 'paid' | 'cancelled' | 'split' | 'pending' | 'preparing' | 'ready' | 'completed' | 'void';
  OrderItems?: OrderItem[];
  items?: OrderItem[];
  Table?: Table;
  User?: User;
  Payments?: Payment[];
  created_at: string;
  updated_at?: string;
}

export interface Payment {
  id: string;
  order_id: string;
  payment_method: 'cash' | 'qris_static' | 'midtrans_dynamic' | 'debit_edc' | 'transfer' | 'qris';
  amount_paid: number;
  cash_tendered?: number;
  gross_amount?: number;
  change_due?: number;
  change_amount?: number;
  midtrans_transaction_id?: string;
  midtrans_payment_type?: string;
  qris_url?: string;
  qr_string?: string;
  qris_payload?: string;
  status: 'pending' | 'settlement' | 'expired' | 'denied' | 'expire' | 'cancel';
  created_at?: string;
  settled_at?: string;
}

export interface Shift {
  id: string;
  user_id: string;
  user_name?: string;
  start_time: string;
  end_time?: string;
  opening_cash: number;
  expected_cash?: number;
  actual_cash?: number;
  cash_difference?: number;
  total_cash_sales?: number;
  total_qris_sales?: number;
  status: 'open' | 'closed';
  notes?: string;
}

export interface CashMovement {
  id: string;
  shift_id: string;
  user_id?: string;
  type: 'cash_in' | 'cash_out' | 'in' | 'out';
  amount: number;
  notes?: string;
  reason?: string;
  created_at: string;
}

export interface FeatureFlags {
  enable_tables: boolean;
  enable_kds: boolean;
  enable_wld: boolean;
  enable_bom: boolean;
  enable_midtrans: boolean;
}

export interface DailyReport {
  date?: string;
  total_orders?: number;
  total_transactions?: number;
  total_revenue?: number;
  total_sales?: number;
  total_tax?: number;
  total_service?: number;
  gross_profit?: number;
  total_cogs?: number;
  payment_breakdowns?: {
    payment_method: string;
    total_amount: number;
    count: number;
  }[];
  sales_by_payment?: {
    cash: number;
    qris: number;
  };
  top_products?: {
    name: string;
    qty: number;
    revenue: number;
  }[];
}

