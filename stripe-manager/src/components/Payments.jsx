import React, { useState, useEffect } from 'react';
import { listPayments } from '../api';

export default function Payments() {
  const [payments, setPayments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState('all');

  const fetchPayments = async () => {
    setLoading(true);
    try {
      const data = await listPayments();
      setPayments(Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchPayments(); }, []);

  const filtered = filter === 'all' ? payments : payments.filter(p => p.status === filter);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">💰 PAYMENTS</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">{payments.length} total</p>
        </div>
        <div className="flex gap-3 items-center">
          {['all', 'succeeded', 'pending', 'failed'].map(f => (
            <button key={f} onClick={() => setFilter(f)}
              className={`px-3 py-1 rounded text-xs font-mono uppercase ${filter === f ? 'bg-cyan/20 text-cyan border border-cyan/40' : 'text-cyan/40 hover:text-cyan'}`}>
              {f}
            </button>
          ))}
          <button onClick={fetchPayments} className="text-xs text-cyan/50 hover:text-cyan font-mono ml-2">[ REFRESH ]</button>
        </div>
      </div>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
        </div>
      )}

      {loading ? (
        <div className="space-y-2">{[...Array(5)].map((_, i) => <div key={i} className="skeleton h-12 rounded" />)}</div>
      ) : filtered.length === 0 ? (
        <div className="glass-card p-8 text-center"><p className="text-cyan/40 font-mono">[ NO PAYMENTS ]</p></div>
      ) : (
        <div className="glass-card overflow-hidden">
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
              {filtered.map((p, i) => (
                <tr key={p.id || i} className="hover:bg-cyan/5">
                  <td className="font-mono text-xs text-cyan/70">{p.id}</td>
                  <td className="font-mono text-acid font-bold">${((p.amount || 0) / 100).toFixed(2)}</td>
                  <td className="font-mono text-cyan/60">{(p.currency || '').toUpperCase()}</td>
                  <td>
                    <span className={`px-2 py-0.5 rounded text-xs ${p.status === 'succeeded' ? 'text-acid bg-acid/10 border border-acid/20' : p.status === 'pending' ? 'text-yellow-400 bg-yellow-400/10 border border-yellow-400/20' : 'text-red-400 bg-red-400/10 border border-red-400/20'}`}>
                      {p.status}
                    </span>
                  </td>
                  <td className="font-mono text-xs text-cyan/60">
                    {p.created ? new Date(p.created * 1000).toLocaleString() : '---'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
