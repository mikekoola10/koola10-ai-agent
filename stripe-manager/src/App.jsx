import React, { useState } from 'react';
import Dashboard from './components/Dashboard';
import Products from './components/Products';
import Customers from './components/Customers';
import Subscriptions from './components/Subscriptions';
import Payments from './components/Payments';
import Checkout from './components/Checkout';
import Webhooks from './components/Webhooks';
import Settings from './components/Settings';

const PAGES = [
  { id: 'dashboard', label: 'Dashboard', icon: '📊' },
  { id: 'products', label: 'Products', icon: '📦' },
  { id: 'customers', label: 'Customers', icon: '👥' },
  { id: 'subscriptions', label: 'Subscriptions', icon: '🔄' },
  { id: 'payments', label: 'Payments', icon: '💰' },
  { id: 'checkout', label: 'Checkout', icon: '💳' },
  { id: 'webhooks', label: 'Webhooks', icon: '🔗' },
  { id: 'settings', label: 'Settings', icon: '⚙️' },
];

export default function App() {
  const [page, setPage] = useState('dashboard');

  const renderPage = () => {
    switch (page) {
      case 'dashboard': return <Dashboard />;
      case 'products': return <Products />;
      case 'customers': return <Customers />;
      case 'subscriptions': return <Subscriptions />;
      case 'payments': return <Payments />;
      case 'checkout': return <Checkout />;
      case 'webhooks': return <Webhooks />;
      case 'settings': return <Settings />;
      default: return <Dashboard />;
    }
  };

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="sidebar w-56 shrink-0 flex flex-col">
        <div className="p-4 border-b border-cyan/10">
          <h1 className="text-lg font-bold text-cyan uppercase tracking-[3px]">
            💳 STRIPE
          </h1>
          <p className="text-[10px] text-cyan/40 uppercase tracking-widest mt-1">
            Spiral Manager v1
          </p>
        </div>
        <nav className="flex-1 py-2">
          {PAGES.map((p) => (
            <button
              key={p.id}
              onClick={() => setPage(p.id)}
              className={`sidebar-link w-full text-left ${page === p.id ? 'active' : ''}`}
            >
              <span>{p.icon}</span>
              <span>{p.label}</span>
            </button>
          ))}
        </nav>
        <div className="p-4 border-t border-cyan/10">
          <p className="text-[9px] text-cyan/20 uppercase tracking-wider text-center">
            koola10 • spiral • apex
          </p>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto p-6">
        <div className="max-w-6xl mx-auto">
          {renderPage()}
        </div>
      </main>
    </div>
  );
}
