import { useState, useEffect } from 'react';
import { apiUrl, ADMIN_KEY } from '../api';

export default function SubscriptionsPanel() {
  const [subs, setSubs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchSubs = () => {
    setLoading(true);
    setError(null);
    fetch(apiUrl('koola10', '/admin/subscriptions'), {
      headers: ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}` } : {},
    })
      .then(async (r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        const data = await r.json();
        setSubs(Array.isArray(data) ? data : data.subscriptions || []);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  };

  useEffect(() => { fetchSubs(); }, []);

  return (
    <div className="glass-card p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-bold font-mono text-purple uppercase tracking-wider">
          💳 SUBSCRIPTIONS
        </h2>
        <button onClick={fetchSubs} className="text-xs text-cyan/50 hover:text-cyan font-mono transition-colors">
          [ REFRESH ]
        </button>
      </div>

      {loading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="skeleton h-12 rounded" />
          ))}
        </div>
      ) : error ? (
        <div className="text-center py-4">
          <p className="text-red-400 text-sm font-mono mb-2">[ ERR: {error} ]</p>
          <button onClick={fetchSubs} className="btn-cyan px-4 py-1.5 text-xs font-mono rounded">
            RETRY
          </button>
        </div>
      ) : subs.length === 0 ? (
        <p className="text-cyan/40 text-sm font-mono text-center py-6">[ NO ACTIVE SUBSCRIPTIONS ]</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-cyan/10">
                <th className="py-2 px-3 text-xs font-mono text-cyan/60 uppercase tracking-wider">Service</th>
                <th className="py-2 px-3 text-xs font-mono text-cyan/60 uppercase tracking-wider">Amount</th>
                <th className="py-2 px-3 text-xs font-mono text-cyan/60 uppercase tracking-wider">Interval</th>
                <th className="py-2 px-3 text-xs font-mono text-cyan/60 uppercase tracking-wider">Next Payment</th>
                <th className="py-2 px-3 text-xs font-mono text-cyan/60 uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody>
              {subs.map((sub, i) => (
                <tr
                  key={sub.id || i}
                  className="border-b border-cyan/5 hover:bg-cyan/5 transition-colors"
                >
                  <td className="py-2.5 px-3 text-sm font-mono">{sub.service || sub.name || '---'}</td>
                  <td className="py-2.5 px-3 text-sm font-mono text-acid">
                    ${typeof sub.amount === 'number' ? sub.amount.toFixed(2) : sub.amount || '---'}
                  </td>
                  <td className="py-2.5 px-3 text-sm font-mono text-cyan/70">{sub.interval || '---'}</td>
                  <td className="py-2.5 px-3 text-sm font-mono text-cyan/70">{sub.next_payment || sub.nextPayment || '---'}</td>
                  <td className="py-2.5 px-3 text-sm font-mono">
                    <span
                      className={`px-2 py-0.5 rounded text-xs ${
                        (sub.status || '').toLowerCase() === 'active'
                          ? 'text-acid bg-acid/10 border border-acid/20'
                          : (sub.status || '').toLowerCase() === 'cancelled'
                          ? 'text-red-400 bg-red-400/10 border border-red-400/20'
                          : 'text-cyan/60 bg-cyan/5 border border-cyan/10'
                      }`}
                    >
                      {sub.status || 'unknown'}
                    </span>
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
