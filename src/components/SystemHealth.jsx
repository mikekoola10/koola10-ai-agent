import { useState, useEffect } from 'react';
import { apiUrl, ADMIN_KEY } from '../api';

const NODES = [
  { id: 'koola10', label: 'Koola10' },
  { id: 'spiral', label: 'Spiral' },
  { id: 'apex', label: 'Apex' },
];

function StatusIndicator({ status }) {
  const colors = {
    online: '#39ff14',
    offline: '#ff3333',
    checking: '#ffaa00',
  };
  const labels = {
    online: 'ONLINE',
    offline: 'OFFLINE',
    checking: 'SCANNING',
  };
  const c = colors[status] || colors.checking;

  return (
    <span className="inline-flex items-center gap-2 text-xs font-mono" style={{ color: c }}>
      <span
        className="inline-block w-2.5 h-2.5 rounded-full"
        style={{ backgroundColor: c, boxShadow: `0 0 8px ${c}` }}
      />
      {labels[status] || 'UNKNOWN'}
    </span>
  );
}

export default function SystemHealth() {
  const [statuses, setStatuses] = useState({ koola10: 'checking', spiral: 'checking', apex: 'checking' });
  const [offlineMode, setOfflineMode] = useState(null);
  const [toggleLoading, setToggleLoading] = useState(false);
  const [toggleError, setToggleError] = useState(null);

  const checkHealth = async (node) => {
    try {
      const res = await fetch(apiUrl(node.id, '/health'), {
        headers: ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}` } : {},
      });
      setStatuses((p) => ({ ...p, [node.id]: 'online' }));
    } catch {
      setStatuses((p) => ({ ...p, [node.id]: 'offline' }));
    }
  };

  useEffect(() => {
    NODES.forEach((n) => checkHealth(n));
    const interval = setInterval(() => NODES.forEach((n) => checkHealth(n)), 30000);
    return () => clearInterval(interval);
  }, []);

  const toggleOffline = async () => {
    if (!ADMIN_KEY) return;
    setToggleLoading(true);
    setToggleError(null);
    try {
      const res = await fetch(apiUrl('koola10', '/admin/toggle-offline'), {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${ADMIN_KEY}`, 'Content-Type': 'application/json' },
      });
      const data = await res.json().catch(() => ({}));
      setOfflineMode(data.mode || (res.ok ? 'toggled' : null));
      setTimeout(() => NODES.forEach((n) => checkHealth(n)), 2000);
    } catch (e) {
      setToggleError(e.message);
    } finally {
      setToggleLoading(false);
    }
  };

  const allOnline = Object.values(statuses).every((s) => s === 'online');

  return (
    <div className="glass-card p-6">
      <div className="flex flex-wrap items-center justify-between gap-4 mb-4">
        <h2 className="text-lg font-bold font-mono text-acid uppercase tracking-wider">
          🫀 SYSTEM HEALTH
        </h2>
        <div className="flex items-center gap-4">
          <span className="text-xs font-mono" style={{ color: allOnline ? '#39ff14' : '#ff3333' }}>
            [ SWARM: {allOnline ? 'HEALTHY' : 'DEGRADED'} ]
          </span>
          <button
            onClick={toggleOffline}
            disabled={toggleLoading}
            className={`px-3 py-1.5 rounded text-xs font-mono uppercase tracking-wider transition-all ${
              offlineMode === 'offline'
                ? 'bg-red-500/20 text-red-400 border border-red-500/50'
                : 'bg-cyan/10 text-cyan border border-cyan/30 hover:bg-cyan/20'
            } disabled:opacity-40`}
          >
            {toggleLoading ? 'TOGGLING...' : offlineMode === 'offline' ? '🔴 OFFLINE' : 'MODE: ONLINE'}
          </button>
        </div>
      </div>

      {toggleError && (
        <p className="text-red-400 text-xs font-mono mb-3 animate-slide-down">[ ERR: {toggleError} ]</p>
      )}

      <div className="flex flex-wrap gap-6">
        {NODES.map((node) => (
          <div key={node.id} className="flex items-center gap-2">
            <span className="text-sm font-mono text-cyan/70 uppercase">{node.label}</span>
            <StatusIndicator status={statuses[node.id]} />
          </div>
        ))}
      </div>
    </div>
  );
}
