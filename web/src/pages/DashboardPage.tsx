import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { TrendingUp, ShoppingCart, DollarSign, Package } from 'lucide-react';
import { formatCurrency } from '../lib/utils';
import { orderAPI, menuAPI, inventoryAPI } from '../api/client';

const COLORS = ['#f97316', '#22c55e', '#3b82f6', '#8b5cf6', '#ec4899', '#14b8a6'];

export default function DashboardPage() {
  // Fetch data
  const { data: ordersData } = useQuery({
    queryKey: ['dashboard-orders'],
    queryFn: async () => {
      const res = await orderAPI.list({ limit: 100 });
      return res.data;
    },
  });

  const { data: statsData } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const res = await menuAPI.stats();
      return res.data;
    },
  });

  const { data: alertsData } = useQuery({
    queryKey: ['dashboard-alerts'],
    queryFn: async () => {
      const res = await inventoryAPI.alerts();
      return res.data;
    },
  });

  const orders = ordersData?.data || [];
  const categories = statsData?.data || [];
  const alerts = alertsData?.data || [];

  // Compute stats
  const todayOrders = orders.filter((o: any) => {
    const d = new Date(o.created_at);
    const today = new Date();
    return d.toDateString() === today.toDateString();
  });

  const todayRevenue = todayOrders.reduce((sum: number, o: any) => sum + o.total, 0);
  const completedOrders = todayOrders.filter((o: any) => o.status === 'completed').length;

  // Revenue by day (mock data for last 7 days)
  const dailyData = Array.from({ length: 7 }, (_, i) => {
    const d = new Date();
    d.setDate(d.getDate() - (6 - i));
    return {
      date: d.toLocaleDateString('id-ID', { weekday: 'short' }),
      revenue: Math.floor(Math.random() * 500000) + 200000,
      orders: Math.floor(Math.random() * 30) + 10,
    };
  });

  const summaryCards = [
    { label: 'Pendapatan Hari Ini', value: formatCurrency(todayRevenue), icon: DollarSign, color: 'text-green-600 bg-green-100' },
    { label: 'Pesanan Hari Ini', value: todayOrders.length.toString(), icon: ShoppingCart, color: 'text-blue-600 bg-blue-100' },
    { label: 'Selesai', value: completedOrders.toString(), icon: TrendingUp, color: 'text-purple-600 bg-purple-100' },
    { label: 'Stok Menipis', value: alerts.length.toString(), icon: Package, color: 'text-red-600 bg-red-100' },
  ];

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {summaryCards.map((card, i) => {
          const Icon = card.icon;
          return (
            <div key={i} className="card">
              <div className="flex items-center gap-3">
                <div className={`p-3 rounded-xl ${card.color}`}>
                  <Icon className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">{card.label}</p>
                  <p className="text-xl font-bold">{card.value}</p>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Revenue Chart */}
        <div className="lg:col-span-2 card">
          <h3 className="font-bold mb-4">Pendapatan 7 Hari</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={dailyData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="date" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`} />
              <Tooltip formatter={(value: number) => formatCurrency(value)} />
              <Bar dataKey="revenue" fill="#f97316" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Category Distribution */}
        <div className="card">
          <h3 className="font-bold mb-4">Kategori Menu</h3>
          {categories.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={categories}
                  dataKey="count"
                  nameKey="category"
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  label={({ category, count }) => `${category}: ${count}`}
                >
                  {categories.map((_: any, i: number) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-64 text-gray-400 text-sm">
              Belum ada data menu
            </div>
          )}
        </div>
      </div>

      {/* Low Stock Alerts */}
      {alerts.length > 0 && (
        <div className="mt-6 card border-red-200">
          <h3 className="font-bold text-red-600 mb-3">⚠️ Stok Menipis</h3>
          <div className="space-y-2">
            {alerts.map((item: any) => (
              <div key={item.id} className="flex items-center justify-between p-2 bg-red-50 rounded-lg">
                <span className="text-sm">{item.item_name}</span>
                <span className="text-sm text-red-600 font-medium">
                  {item.quantity} {item.unit} (min: {item.min_stock})
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
