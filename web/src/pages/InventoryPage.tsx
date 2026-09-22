import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { inventoryAPI } from '../api/client';
import { formatCurrency } from '../lib/utils';
import { Plus, AlertTriangle, Package, RefreshCw } from 'lucide-react';
import type { InventoryItem } from '../types';

export default function InventoryPage() {
  const queryClient = useQueryClient();
  const [branchId] = useState('a0000000-0000-0000-0000-000000000001');
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    item_name: '', quantity: 0, unit: 'kg', min_quantity: 0, unit_cost: 0,
  });

  // Fetch inventory
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['inventory', branchId],
    queryFn: async () => {
      const res = await inventoryAPI.list({ branch_id: branchId, limit: 100 });
      return res.data;
    },
  });

  // Fetch low stock alerts
  const { data: alerts } = useQuery({
    queryKey: ['inventory-alerts', branchId],
    queryFn: async () => {
      const res = await inventoryAPI.alerts(branchId);
      return res.data;
    },
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => inventoryAPI.create({ ...data, branch_id: branchId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      setShowForm(false);
      setFormData({ item_name: '', quantity: 0, unit: 'kg', min_quantity: 0, unit_cost: 0 });
    },
  });

  const items = data?.data || [];
  const lowStockItems = alerts?.data || [];

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Inventaris</h1>
          <p className="text-gray-500 text-sm">Kelola stok bahan baku</p>
        </div>
        <div className="flex gap-2">
          <button onClick={() => refetch()} className="btn-secondary flex items-center gap-2">
            <RefreshCw className="w-4 h-4" />
          </button>
          <button onClick={() => setShowForm(!showForm)} className="btn-primary flex items-center gap-2">
            <Plus className="w-4 h-4" />
            Tambah Item
          </button>
        </div>
      </div>

      {/* Low Stock Alert */}
      {lowStockItems.length > 0 && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="w-5 h-5 text-red-500" />
            <h3 className="font-bold text-red-700">Stok Menipis ({lowStockItems.length} item)</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
            {lowStockItems.map((item: InventoryItem) => (
              <div key={item.id} className="flex justify-between text-sm p-2 bg-white rounded-lg">
                <span>{item.item_name}</span>
                <span className="text-red-600 font-medium">
                  {item.quantity} {item.unit}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Create Form */}
      {showForm && (
        <div className="mb-6 card border-warung-orange">
          <h3 className="font-bold mb-3">Tambah Item Baru</h3>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
            <input
              type="text"
              placeholder="Nama item"
              value={formData.item_name}
              onChange={e => setFormData({ ...formData, item_name: e.target.value })}
              className="input-field"
            />
            <input
              type="number"
              placeholder="Jumlah"
              value={formData.quantity || ''}
              onChange={e => setFormData({ ...formData, quantity: Number(e.target.value) })}
              className="input-field"
            />
            <select
              value={formData.unit}
              onChange={e => setFormData({ ...formData, unit: e.target.value })}
              className="input-field"
            >
              <option value="kg">kg</option>
              <option value="liter">liter</option>
              <option value="pcs">pcs</option>
              <option value="pack">pack</option>
            </select>
            <input
              type="number"
              placeholder="Min stok"
              value={formData.min_quantity || ''}
              onChange={e => setFormData({ ...formData, min_quantity: Number(e.target.value) })}
              className="input-field"
            />
            <button
              onClick={() => createMutation.mutate(formData)}
              className="btn-primary"
              disabled={!formData.item_name}
            >
              Simpan
            </button>
          </div>
        </div>
      )}

      {/* Inventory Table */}
      {isLoading ? (
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="card animate-pulse h-16" />
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="text-center py-12 text-gray-400">
          <Package className="w-12 h-12 mx-auto mb-2" />
          <p>Belum ada item inventaris</p>
        </div>
      ) : (
        <div className="card overflow-hidden p-0">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b">
                <th className="text-left p-3 text-sm font-medium text-gray-500">Item</th>
                <th className="text-right p-3 text-sm font-medium text-gray-500">Stok</th>
                <th className="text-right p-3 text-sm font-medium text-gray-500">Min Stok</th>
                <th className="text-right p-3 text-sm font-medium text-gray-500">Harga/Unit</th>
                <th className="text-center p-3 text-sm font-medium text-gray-500">Status</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item: InventoryItem) => {
                const isLow = item.quantity <= item.min_quantity;
                return (
                  <tr key={item.id} className={`border-b hover:bg-gray-50 ${isLow ? 'bg-red-50' : ''}`}>
                    <td className="p-3">
                      <p className="font-medium">{item.item_name}</p>
                      <p className="text-xs text-gray-400">{item.unit}</p>
                    </td>
                    <td className="p-3 text-right font-medium">{item.quantity}</td>
                    <td className="p-3 text-right text-gray-500">{item.min_quantity}</td>
                    <td className="p-3 text-right">{formatCurrency(item.unit_cost)}</td>
                    <td className="p-3 text-center">
                      <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                        isLow ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'
                      }`}>
                        {isLow ? 'Menipis' : 'Aman'}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
