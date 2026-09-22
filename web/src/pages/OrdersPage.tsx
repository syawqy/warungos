import { useState } from 'react';
import { useOrders, useUpdateOrderStatus } from '../hooks/useOrders';
import { formatCurrency, formatDate, getStatusColor } from '../lib/utils';
import { RefreshCw, ChefHat, CheckCircle, XCircle, Clock } from 'lucide-react';

const STATUS_OPTIONS = [
  { value: '', label: 'Semua' },
  { value: 'pending', label: 'Menunggu' },
  { value: 'confirmed', label: 'Dikonfirmasi' },
  { value: 'preparing', label: 'Disiapkan' },
  { value: 'ready', label: 'Siap' },
  { value: 'completed', label: 'Selesai' },
  { value: 'cancelled', label: 'Dibatalkan' },
];

const NEXT_STATUS: Record<string, string> = {
  pending: 'confirmed',
  confirmed: 'preparing',
  preparing: 'ready',
  ready: 'completed',
};

export default function OrdersPage() {
  const [statusFilter, setStatusFilter] = useState('');
  const { data, isLoading, refetch } = useOrders(undefined, statusFilter);
  const updateStatus = useUpdateOrderStatus();

  const orders = data?.data || [];

  const handleStatusUpdate = async (orderId: string, newStatus: string) => {
    try {
      await updateStatus.mutateAsync({ id: orderId, status: newStatus });
    } catch (error) {
      console.error('Failed to update status:', error);
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Pesanan</h1>
          <p className="text-gray-500 text-sm">Kelola semua pesanan</p>
        </div>
        <button onClick={() => refetch()} className="btn-secondary flex items-center gap-2">
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Status Filter */}
      <div className="flex gap-2 mb-6 overflow-x-auto">
        {STATUS_OPTIONS.map(opt => (
          <button
            key={opt.value}
            onClick={() => setStatusFilter(opt.value)}
            className={`px-4 py-2 text-sm rounded-lg whitespace-nowrap transition-colors ${
              statusFilter === opt.value
                ? 'bg-warung-orange text-white'
                : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'
            }`}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* Orders List */}
      {isLoading ? (
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="card animate-pulse h-24" />
          ))}
        </div>
      ) : orders.length === 0 ? (
        <div className="text-center py-12 text-gray-400">
          <Clock className="w-12 h-12 mx-auto mb-2" />
          <p>Belum ada pesanan</p>
        </div>
      ) : (
        <div className="space-y-3">
          {orders.map((order: any) => (
            <div key={order.id} className="card">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-1">
                    <span className="font-mono text-sm text-gray-500">#{order.id.slice(0, 8)}</span>
                    <span className={`px-2 py-0.5 text-xs rounded-full font-medium ${getStatusColor(order.status)}`}>
                      {order.status}
                    </span>
                    <span className="text-xs text-gray-400 capitalize">{order.order_type.replace('_', ' ')}</span>
                  </div>
                  <p className="text-sm text-gray-600">
                    {order.customer_name || 'Anonim'} — {order.items?.length || 0} item
                  </p>
                  <p className="text-xs text-gray-400 mt-1">{formatDate(order.created_at)}</p>
                </div>

                <div className="flex items-center gap-3">
                  <span className="text-lg font-bold text-warung-orange">
                    {formatCurrency(order.total)}
                  </span>

                  {NEXT_STATUS[order.status] && (
                    <button
                      onClick={() => handleStatusUpdate(order.id, NEXT_STATUS[order.status])}
                      className="btn-primary text-sm flex items-center gap-1"
                    >
                      {order.status === 'pending' && <CheckCircle className="w-3.5 h-3.5" />}
                      {order.status === 'confirmed' && <ChefHat className="w-3.5 h-3.5" />}
                      {order.status === 'preparing' && <CheckCircle className="w-3.5 h-3.5" />}
                      {order.status === 'ready' && <CheckCircle className="w-3.5 h-3.5" />}
                      {NEXT_STATUS[order.status] === 'confirmed' && 'Konfirmasi'}
                      {NEXT_STATUS[order.status] === 'preparing' && 'Mulai Masak'}
                      {NEXT_STATUS[order.status] === 'ready' && 'Siap Saji'}
                      {NEXT_STATUS[order.status] === 'completed' && 'Selesai'}
                    </button>
                  )}

                  {['pending', 'confirmed'].includes(order.status) && (
                    <button
                      onClick={() => handleStatusUpdate(order.id, 'cancelled')}
                      className="btn-danger text-sm flex items-center gap-1"
                    >
                      <XCircle className="w-3.5 h-3.5" />
                      Batal
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
