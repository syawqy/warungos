import { Trash2, Minus, Plus, ShoppingBag, Utensils, Bike } from 'lucide-react';
import { formatCurrency } from '../../lib/utils';
import type { CartItem, OrderType } from '../../types';

interface CartPanelProps {
  items: CartItem[];
  customerName: string;
  orderType: OrderType;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  itemCount: number;
  onRemoveItem: (index: number) => void;
  onUpdateQuantity: (index: number, quantity: number) => void;
  onSetCustomerName: (name: string) => void;
  onSetOrderType: (type: OrderType) => void;
  onSetDiscount: (discount: number) => void;
  onClearCart: () => void;
  onPay: () => void;
}

const ORDER_TYPES: { value: OrderType; label: string; icon: typeof Utensils }[] = [
  { value: 'dine_in', label: 'Dine In', icon: Utensils },
  { value: 'takeaway', label: 'Take Away', icon: ShoppingBag },
  { value: 'delivery', label: 'Delivery', icon: Bike },
];

export default function CartPanel({
  items, customerName, orderType, subtotal, tax, discount, total, itemCount,
  onRemoveItem, onUpdateQuantity, onSetCustomerName, onSetOrderType,
  onSetDiscount, onClearCart, onPay,
}: CartPanelProps) {
  return (
    <div className="flex flex-col h-full bg-white border-l border-gray-200">
      {/* Header */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-bold text-lg">Pesanan</h2>
          <span className="text-sm text-gray-500">{itemCount} item</span>
        </div>
        <input
          type="text"
          placeholder="Nama pelanggan..."
          value={customerName}
          onChange={(e) => onSetCustomerName(e.target.value)}
          className="input-field text-sm"
        />
      </div>

      {/* Order Type */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex gap-2">
          {ORDER_TYPES.map(({ value, label, icon: Icon }) => (
            <button
              key={value}
              onClick={() => onSetOrderType(value)}
              className={`flex-1 flex items-center justify-center gap-1 py-2 text-xs rounded-lg transition-colors ${
                orderType === value
                  ? 'bg-warung-orange text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              <Icon className="w-3.5 h-3.5" />
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Cart Items */}
      <div className="flex-1 overflow-y-auto p-4">
        {items.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-400">
            <ShoppingBag className="w-12 h-12 mb-2" />
            <p className="text-sm">Belum ada item</p>
            <p className="text-xs">Pilih menu dari panel kiri</p>
          </div>
        ) : (
          <div className="space-y-3">
            {items.map((item, index) => (
              <div key={index} className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm truncate">{item.menu_item.name}</p>
                  <p className="text-xs text-gray-500">{formatCurrency(item.menu_item.price)}</p>
                  {item.notes && <p className="text-xs text-gray-400 italic">"{item.notes}"</p>}
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => onUpdateQuantity(index, item.quantity - 1)}
                    className="w-6 h-6 rounded bg-gray-200 hover:bg-gray-300 flex items-center justify-center"
                  >
                    <Minus className="w-3 h-3" />
                  </button>
                  <span className="w-6 text-center text-sm font-medium">{item.quantity}</span>
                  <button
                    onClick={() => onUpdateQuantity(index, item.quantity + 1)}
                    className="w-6 h-6 rounded bg-warung-orange hover:bg-orange-600 text-white flex items-center justify-center"
                  >
                    <Plus className="w-3 h-3" />
                  </button>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium">{formatCurrency(item.menu_item.price * item.quantity)}</p>
                  <button
                    onClick={() => onRemoveItem(index)}
                    className="text-red-400 hover:text-red-600 mt-1"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Discount */}
      <div className="px-4 pb-2">
        <label className="text-xs text-gray-500 mb-1 block">Diskon (Rp)</label>
        <input
          type="number"
          value={discount || ''}
          onChange={(e) => onSetDiscount(Number(e.target.value))}
          placeholder="0"
          className="input-field text-sm"
        />
      </div>

      {/* Summary & Pay */}
      <div className="p-4 border-t border-gray-200 bg-gray-50">
        <div className="space-y-1 mb-3">
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">Subtotal</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">PPN (10%)</span>
            <span>{formatCurrency(tax)}</span>
          </div>
          {discount > 0 && (
            <div className="flex justify-between text-sm text-green-600">
              <span>Diskon</span>
              <span>-{formatCurrency(discount)}</span>
            </div>
          )}
          <div className="flex justify-between text-lg font-bold border-t border-gray-300 pt-1">
            <span>Total</span>
            <span className="text-warung-orange">{formatCurrency(total)}</span>
          </div>
        </div>

        <div className="flex gap-2">
          <button onClick={onClearCart} className="btn-secondary flex-1 text-sm" disabled={items.length === 0}>
            Batal
          </button>
          <button
            onClick={onPay}
            className="btn-primary flex-1 text-sm"
            disabled={items.length === 0}
          >
            Bayar
          </button>
        </div>
      </div>
    </div>
  );
}
