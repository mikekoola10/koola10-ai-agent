import { useState, useEffect, useRef } from 'react';
import { apiUrl, ADMIN_KEY } from '../api';

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
        {prefix}
        {typeof value === 'number' ? value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : value ?? '---'}
        {suffix}
      </p>
    </div>
  );
}

export default function LedgerDashboard({ onRefresh }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchLedger = () => {
    setLoading(true);
    setError(null);
    fetch(apiUrl('koola10', '/vault/summary'), {
      headers: ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}` } : {},
    })
      .then(async (r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        const d = await r.json();
        setData(d);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  };

  const fetchRef = useRef(fetchLedger);
  fetchRef.current = fetchLedger;

  useEffect(() => { fetchLedger(); }, []);

  useEffect(() => {
    if (onRefresh) onRefresh.current = () => fetchRef.current();
  }, [onRefresh]);

  if (loading) {
    return (
      <div className="glass-card p-6">
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="skeleton h-24 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="glass-card p-6 text-center" style={{ borderColor: '#ff333366' }}>
        <p className="text-red-400 text-sm font-mono mb-2">[ ERR: {error} ]</p>
        <button onClick={fetchLedger} className="btn-cyan px-4 py-1.5 text-xs font-mono rounded">
          RETRY
        </button>
      </div>
    );
  }

  const stats = [
    { label: 'Total Revenue', value: data?.total_revenue, prefix: '$', color: '#00f0ff', delay: 0 },
    { label: 'Operations Fund (30%)', value: data?.operations_fund, prefix: '$', color: '#8b00ff', delay: 0.1 },
    { label: 'Spendable Fund (70%)', value: data?.spendable_fund, prefix: '$', color: '#39ff14', delay: 0.2 },
    { label: 'Total Costs', value: data?.total_costs, prefix: '$', color: '#ff6b6b', delay: 0.3 },
    { label: 'ROI Ratio', value: data?.roi_ratio, suffix: 'x', color: '#ffd93d', delay: 0.4 },
  ];

  return (
    <div className="glass-card p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-bold font-mono text-cyan uppercase tracking-wider">
          📒 LEDGER
        </h2>
        <button onClick={fetchLedger} className="text-xs text-cyan/50 hover:text-cyan font-mono transition-colors">
          [ REFRESH ]
        </button>
      </div>
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {stats.map((s, i) => (
          <StatCard key={i} {...s} />
        ))}
      </div>
    </div>
  );
}
