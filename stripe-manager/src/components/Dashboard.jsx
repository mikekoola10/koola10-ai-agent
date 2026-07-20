import React, { useState, useEffect } from 'react';
import { getRevenue, listPayments } from '../api';

function StatCard({ label, value, prefix = '', suffix = '', color = '#00f0ff', delay = 0 }) {
  return (
    <div
      className="glass-card p-5 text-center animate-fade-in-up"
      style={{ animationDelay: `${delay}s`, borderColor: `${color}33` }}
    >
      <p className="text-xs uppercase tracking-widest mb-2" style={{ color: `${color}99` }}>
        {label}
      </p>
      <p className="text-2xl md:text-3xl font-bold font-mono" style={{ color }}>
        {prefix}{typeof value === 'number' ? value.toLocaleString() : (value ?? '---')}{suffix}
      </p>
    </div>
  );
}

export default function Dashboard() {
  const [revenue, setRevenue] = useState(null);
  const [payments, setPayments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [rev, pay] = await Promise.all([getRevenue(), listPayments()]);
      setRevenue(rev);
      setPayments(Array.isArray(pay) ? pay.slice(0, 10) : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => <div key={i} className="skeleton h-24 rounded-lg" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="glass-card p-6 text-center" style={{ borderColor: '#ff333366' }}>
        <p className="text-red-400 font-mono mb-3">[ ERR: {error} ]</p>
        <button onClick={fetchData} className="btn-cyan px-4 py-1.5 text-xs font-mono rounded">
          RETRY
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">
            💳 STRIPE MANAGER
          </h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono uppercase tracking-wider">
            Revenue overview
          </p>
        </div>
        <button onClick={fetchData} className="text-xs text-cyan/50 hover:text-cyan font-mono transition-colors">
          [ REFRESH ]
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          label="Total Revenue"
          value={revenue?.total_revenue ? `$${(revenue.total_revenue / 100).toFixed(2)}` : '$0.00'}
          color="#00f0ff"
          delay={0}
        />
        <StatCard
          label="MRR"
          value={revenue?.total_mrr ? `$${(revenue.total_mrr / 100).toFixed(2)}` : '$0.00'}
          color="#8b00ff"
          delay={0.1}
        />
        <StatCard
          label="Active Subscriptions"
          value={revenue?.active_subscriptions ?? 0}
          color="#39ff14"
          delay={0.2}
        />
        <StatCard
          label="Total Payments"
          value={revenue?.total_payments ?? 0}
          suffix=""
          color="#ffd93d"
          delay={0.3}
        />
      </div>

      {/* Secondary Stats */}
      <div className="grid grid-cols-2 gap-4">
        <StatCard
          label="Products"
          value={revenue?.products ?? 0}
          color="#00f0ff"
          delay={0.4}
        />
        <StatCard
          label="Customers"
          value={revenue?.customers ?? 0}
          color="#8b00ff"
          delay={0.5}
        />
      </div>

      {/* Recent Payments */}
      <div className="glass-card p-6">
        <h2 className="text-lg font-bold font-mono text-cyan uppercase tracking-wider mb-4">
          💰 RECENT PAYMENTS
        </h2>
        {payments.length === 0 ? (
          <p className="text-cyan/40 text-sm font-mono text-center py-4">[ NO PAYMENTS YET ]</p>
        ) : (
          <div className="overflow-x-auto">
            <table>
              <thead>
                <tr className="border-b border-cyan/10">
                  <th>ID</th>
                  <th>Amount</th>
                  <th>Currency</th>
                  <th>Status</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {payments.map((p, i) => (
                  <tr key={p.id || i} className="hover:bg-cyan/5">
                    <td className="font-mono text-xs">{p.id?.slice(0, 20)}...</td>
                    <td className="font-mono text-acid">${((p.amount || 0) / 100).toFixed(2)}</td>
                    <td className="font-mono text-cyan/70">{(p.currency || '').toUpperCase()}</td>
                    <td>
                      <span className={`px-2 py-0.5 rounded text-xs ${
                        p.status === 'succeeded'
                          ? 'text-acid bg-acid/10 border border-acid/20'
                          : p.status === 'pending'
                          ? 'text-yellow-400 bg-yellow-400/10 border border-yellow-400/20'
                          : 'text-red-400 bg-red-400/10 border border-red-400/20'
                      }`}>
                        {p.status || 'unknown'}
                      </span>
                    </td>
                    <td className="font-mono text-cyan/60 text-xs">
                      {p.created ? new Date(p.created * 1000).toLocaleDateString() : '---'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
