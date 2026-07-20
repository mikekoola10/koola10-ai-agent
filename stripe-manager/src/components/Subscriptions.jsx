import React, { useState, useEffect } from 'react';
import { listSubscriptions, cancelSubscription } from '../api';

export default function Subscriptions() {
  const [subs, setSubs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchSubs = async () => {
    setLoading(true);
    try {
      const data = await listSubscriptions();
      setSubs(Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchSubs(); }, []);

  const handleCancel = async (id) => {
    if (!confirm('Cancel this subscription?')) return;
    try {
      await cancelSubscription(id);
      fetchSubs();
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">🔄 SUBSCRIPTIONS</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">{subs.length} total</p>
        </div>
        <button onClick={fetchSubs} className="text-xs text-cyan/50 hover:text-cyan font-mono">[ REFRESH ]</button>
      </div>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
          <button onClick={() => setError(null)} className="ml-4 text-cyan/50 hover:text-cyan">dismiss</button>
        </div>
      )}

      {loading ? (
        <div className="space-y-2">{[...Array(3)].map((_, i) => <div key={i} className="skeleton h-12 rounded" />)}</div>
      ) : subs.length === 0 ? (
        <div className="glass-card p-8 text-center"><p className="text-cyan/40 font-mono">[ NO SUBSCRIPTIONS ]</p></div>
      ) : (
        <div className="glass-card overflow-hidden">
          <table>
            <thead>
              <tr className="border-b border-cyan/10">
                <th>ID</th>
                <th>Status</th>
                <th>Period End</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {subs.map((s) => (
                <tr key={s.id} className="hover:bg-cyan/5">
                  <td className="font-mono text-xs text-cyan/70">{s.id}</td>
                  <td>
                    <span className={`px-2 py-0.5 rounded text-xs ${s.status === 'active' ? 'text-acid bg-acid/10 border border-acid/20' : s.status === 'canceled' ? 'text-red-400 bg-red-400/10 border border-red-400/20' : 'text-yellow-400 bg-yellow-400/10 border border-yellow-400/20'}`}>
                      {(s.status || 'unknown').toUpperCase()}
                    </span>
                  </td>
                  <td className="font-mono text-xs text-cyan/60">
                    {s.current_period_end ? new Date(s.current_period_end * 1000).toLocaleDateString() : '---'}
                  </td>
                  <td className="font-mono text-xs text-cyan/60">
                    {s.created ? new Date(s.created * 1000).toLocaleDateString() : '---'}
                  </td>
                  <td>
                    {s.status === 'active' && (
                      <button onClick={() => handleCancel(s.id)} className="btn-red px-3 py-1 rounded text-xs font-mono">
                        CANCEL
                      </button>
                    )}
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
