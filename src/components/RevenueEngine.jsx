import { useState } from 'react';
import { apiUrl, ADMIN_KEY } from '../api';

const ACTIONS = [
  { id: 'run-full-sprint', label: '🚀 Run Full Sprint', endpoint: '/admin/run-full-sprint', color: 'cyan' },
  { id: 'run-scheduled-sprint', label: '⏰ Scheduled Sprint', endpoint: '/admin/run-scheduled-sprint', color: 'cyan' },
  { id: 'trigger_affiliate', label: '💰 Trigger Affiliate', endpoint: '/admin/trigger_affiliate', color: 'purple' },
  { id: 'trigger_bounty', label: '🏆 Trigger Bounty', endpoint: '/admin/trigger_bounty', color: 'acid' },
  { id: 'trigger_content', label: '📝 Trigger Content', endpoint: '/admin/trigger_content', color: 'cyan' },
];

export default function RevenueEngine({ onAction }) {
  const [loading, setLoading] = useState({});
  const [results, setResults] = useState({});

  const trigger = async (action) => {
    if (!ADMIN_KEY) return;
    setLoading((prev) => ({ ...prev, [action.id]: true }));
    setResults((prev) => ({ ...prev, [action.id]: null }));

    try {
      const res = await fetch(apiUrl('koola10', action.endpoint), {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${ADMIN_KEY}`, 'Content-Type': 'application/json' },
      });
      const data = await res.json().catch(() => ({ status: res.status }));
      setResults((prev) => ({
        ...prev,
        [action.id]: { ok: res.ok, message: data.message || data.status || (res.ok ? 'OK' : 'Failed') },
      }));
      if (res.ok && onAction) onAction();
    } catch (e) {
      setResults((prev) => ({ ...prev, [action.id]: { ok: false, message: e.message } }));
    } finally {
      setLoading((prev) => ({ ...prev, [action.id]: false }));
      setTimeout(() => {
        setResults((prev) => {
          const next = { ...prev };
          delete next[action.id];
          return next;
        });
      }, 4000);
    }
  };

  return (
    <div className="glass-card p-6">
      <h2 className="text-lg font-bold font-mono text-purple uppercase tracking-wider mb-4">
        ⚡ REVENUE ENGINE
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        {ACTIONS.map((action) => {
          const isRunning = loading[action.id];
          const result = results[action.id];
          const btnClass = action.color === 'cyan' ? 'btn-cyan' : action.color === 'purple' ? 'btn-purple' : 'btn-acid';

          return (
            <div key={action.id} className="flex flex-col gap-2">
              <button
                onClick={() => trigger(action)}
                disabled={isRunning || !ADMIN_KEY}
                className={`${btnClass} px-4 py-3 rounded-lg font-mono text-sm uppercase tracking-wider cursor-pointer disabled:cursor-not-allowed transition-all`}
              >
                {isRunning ? (
                  <span className="inline-flex items-center gap-2">
                    <span className="animate-spin inline-block w-3 h-3 border border-current border-t-transparent rounded-full" />
                    RUNNING...
                  </span>
                ) : (
                  action.label
                )}
              </button>
              {result && (
                <p
                  className={`text-xs font-mono px-2 py-1 rounded animate-slide-down ${
                    result.ok ? 'text-acid bg-acid/10' : 'text-red-400 bg-red-400/10'
                  }`}
                >
                  {result.ok ? '✓' : '✗'} {result.message}
                </p>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
