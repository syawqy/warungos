import { useState } from 'react';
import { useAuth } from '../../hooks/useAuth';
import { LayoutDashboard, ShoppingCart, ClipboardList, Package, LogOut, Menu, X, Store } from 'lucide-react';

type Page = 'pos' | 'orders' | 'dashboard' | 'inventory';

interface SidebarProps {
  currentPage: Page;
  onNavigate: (page: Page) => void;
}

export default function Sidebar({ currentPage, onNavigate }: SidebarProps) {
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  const navItems: { id: Page; label: string; icon: typeof LayoutDashboard }[] = [
    { id: 'pos', label: 'Kasir', icon: ShoppingCart },
    { id: 'orders', label: 'Pesanan', icon: ClipboardList },
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'inventory', label: 'Inventaris', icon: Package },
  ];

  return (
    <aside className={`${collapsed ? 'w-16' : 'w-64'} bg-warung-dark text-white flex flex-col transition-all duration-300`}>
      {/* Header */}
      <div className="p-4 flex items-center justify-between border-b border-gray-700">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <Store className="w-6 h-6 text-warung-orange" />
            <span className="font-bold text-lg">WarungOS</span>
          </div>
        )}
        <button onClick={() => setCollapsed(!collapsed)} className="text-gray-400 hover:text-white">
          {collapsed ? <Menu className="w-5 h-5" /> : <X className="w-5 h-5" />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4">
        {navItems.map(item => {
          const Icon = item.icon;
          const isActive = currentPage === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={`w-full flex items-center gap-3 px-4 py-3 text-sm transition-colors ${
                isActive
                  ? 'bg-warung-orange text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-white'
              }`}
            >
              <Icon className="w-5 h-5 flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </button>
          );
        })}
      </nav>

      {/* User info */}
      <div className="p-4 border-t border-gray-700">
        {!collapsed && user && (
          <div className="mb-3">
            <p className="text-sm font-medium text-white truncate">{user.name}</p>
            <p className="text-xs text-gray-400 capitalize">{user.role}</p>
          </div>
        )}
        <button
          onClick={logout}
          className="w-full flex items-center gap-2 text-gray-400 hover:text-red-400 text-sm transition-colors"
        >
          <LogOut className="w-4 h-4" />
          {!collapsed && <span>Keluar</span>}
        </button>
      </div>
    </aside>
  );
}
