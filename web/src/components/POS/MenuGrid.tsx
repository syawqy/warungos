import { useState } from 'react';
import { Search, Plus } from 'lucide-react';
import { useMenu } from '../../hooks/useMenu';
import { formatCurrency } from '../../lib/utils';
import type { MenuItem, CartItem } from '../../types';

interface MenuGridProps {
  onAddToCart: (item: CartItem) => void;
}

const CATEGORIES = [
  { id: '', label: 'Semua' },
  { id: 'makanan', label: 'Makanan' },
  { id: 'minuman', label: 'Minuman' },
  { id: 'side_dish', label: 'Side Dish' },
  { id: 'desserts', label: 'Dessert' },
  { id: 'snacks', label: 'Cemilan' },
  { id: 'promo', label: 'Paket Promo' },
];

export default function MenuGrid({ onAddToCart }: MenuGridProps) {
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const { data: menuItems, isLoading } = useMenu(undefined, selectedCategory, search);

  const handleAdd = (item: MenuItem) => {
    onAddToCart({
      menu_item: item,
      quantity: 1,
      selected_modifiers: [],
      notes: '',
    });
  };

  return (
    <div className="flex flex-col h-full">
      {/* Search & Categories */}
      <div className="p-4 border-b border-gray-200">
        <div className="relative mb-3">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Cari menu..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input-field pl-10"
          />
        </div>
        <div className="flex gap-2 overflow-x-auto pb-1">
          {CATEGORIES.map(cat => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-3 py-1.5 text-sm rounded-full whitespace-nowrap transition-colors ${
                selectedCategory === cat.id
                  ? 'bg-warung-orange text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      {/* Menu Items Grid */}
      <div className="flex-1 overflow-y-auto p-4">
        {isLoading ? (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {[...Array(8)].map((_, i) => (
              <div key={i} className="card animate-pulse">
                <div className="h-24 bg-gray-200 rounded-lg mb-2" />
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-1" />
                <div className="h-3 bg-gray-200 rounded w-1/2" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {menuItems?.map(item => (
              <button
                key={item.id}
                onClick={() => handleAdd(item)}
                disabled={!item.is_available}
                className="card hover:shadow-md hover:border-warung-orange transition-all text-left group disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <div className="h-24 bg-gray-100 rounded-lg mb-2 flex items-center justify-center overflow-hidden">
                  {item.image_url ? (
                    <img src={item.image_url} alt={item.name} className="w-full h-full object-cover" />
                  ) : (
                    <span className="text-3xl">🍽️</span>
                  )}
                </div>
                <h3 className="font-medium text-sm text-gray-900 truncate group-hover:text-warung-orange">
                  {item.name}
                </h3>
                <p className="text-xs text-gray-500 truncate">{item.description}</p>
                <div className="flex items-center justify-between mt-2">
                  <span className="text-sm font-bold text-warung-orange">{formatCurrency(item.price)}</span>
                  <Plus className="w-4 h-4 text-gray-400 group-hover:text-warung-orange" />
                </div>
                {item.tags?.includes('popular') && (
                  <span className="inline-block mt-1 px-2 py-0.5 bg-red-100 text-red-600 text-xs rounded-full">
                    Populer
                  </span>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
