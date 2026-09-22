import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { orderAPI } from '../api/client';
import type { CartItem, OrderType } from '../types';

export function useOrders(branchId?: string, status?: string) {
  return useQuery({
    queryKey: ['orders', branchId, status],
    queryFn: async () => {
      const res = await orderAPI.list({ branch_id: branchId, status, limit: 50 });
      return res.data;
    },
  });
}

export function useCreateOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof orderAPI.create>[0]) => orderAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });
}

export function useUpdateOrderStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => orderAPI.updateStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });
}

// Cart state management
interface CartState {
  items: CartItem[];
  customerName: string;
  orderType: OrderType;
  discount: number;
}

export function useCart() {
  const [cart, setCart] = useState<CartState>({
    items: [],
    customerName: '',
    orderType: 'dine_in',
    discount: 0,
  });

  const addItem = useCallback((item: CartItem) => {
    setCart(prev => {
      const existing = prev.items.find(
        i => i.menu_item.id === item.menu_item.id && i.selected_variant === item.selected_variant
      );
      if (existing) {
        return {
          ...prev,
          items: prev.items.map(i =>
            i === existing ? { ...i, quantity: i.quantity + item.quantity } : i
          ),
        };
      }
      return { ...prev, items: [...prev.items, item] };
    });
  }, []);

  const removeItem = useCallback((index: number) => {
    setCart(prev => ({
      ...prev,
      items: prev.items.filter((_, i) => i !== index),
    }));
  }, []);

  const updateQuantity = useCallback((index: number, quantity: number) => {
    if (quantity <= 0) {
      setCart(prev => ({ ...prev, items: prev.items.filter((_, i) => i !== index) }));
      return;
    }
    setCart(prev => ({
      ...prev,
      items: prev.items.map((item, i) => (i === index ? { ...item, quantity } : item)),
    }));
  }, []);

  const clearCart = useCallback(() => {
    setCart({ items: [], customerName: '', orderType: 'dine_in', discount: 0 });
  }, []);

  const subtotal = cart.items.reduce((sum, item) => {
    const price = item.menu_item.price;
    const variantMod = item.menu_item.variants?.find(v => v.name === item.selected_variant)?.price_modifier || 0;
    const modTotal = item.menu_item.modifiers
      ?.filter(m => item.selected_modifiers.includes(m.name))
      .reduce((s, m) => s + m.price, 0) || 0;
    return sum + (price + variantMod + modTotal) * item.quantity;
  }, 0);

  const tax = Math.round(subtotal * 0.1);
  const total = subtotal + tax - cart.discount;

  return {
    ...cart,
    addItem,
    removeItem,
    updateQuantity,
    clearCart,
    setCustomerName: (name: string) => setCart(prev => ({ ...prev, customerName: name })),
    setOrderType: (type: OrderType) => setCart(prev => ({ ...prev, orderType: type })),
    setDiscount: (discount: number) => setCart(prev => ({ ...prev, discount })),
    subtotal,
    tax,
    total,
    itemCount: cart.items.reduce((sum, item) => sum + item.quantity, 0),
  };
}
