import React, { useState, useEffect } from 'react';
import { listWebhooks, createWebhook, deleteWebhook } from '../api';

export default function Webhooks() {
  const [webhooks, setWebhooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ url: '', enabled_events: 'checkout.session.completed,invoice.payment_succeeded,customer.subscription.deleted' });
  const [creating, setCreating] = useState(false);

  const fetchWebhooks = async () => {
    setLoading(true);
    try {
      const data = await listWebhooks();
      setWebhooks(Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchWebhooks(); }, []);

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreating(true);
    try {
      await createWebhook({
        url: form.url,
        enabled_events: form.enabled_events.split(',').map(s => s.trim()),
      });
      setForm({ url: '', enabled_events: 'checkout.session.completed,invoice.payment_succeeded,customer.subscription.deleted' });
      setShowCreate(false);
      fetchWebhooks();
    } catch (e) {
      setError(e.message);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this webhook endpoint?')) return;
    try {
      await deleteWebhook(id);
      fetchWebhooks();
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">🔗 WEBHOOKS</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">{webhooks.length} endpoints</p>
        </div>
        <div className="flex gap-3">
          <button onClick={fetchWebhooks} className="text-xs text-cyan/50 hover:text-cyan font-mono">[ REFRESH ]</button>
          <button onClick={() => setShowCreate(!showCreate)} className="btn-cyan px-4 py-2 rounded text-xs font-mono uppercase">
            {showCreate ? '✕ CANCEL' : '+ NEW WEBHOOK'}
          </button>
        </div>
      </div>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
        </div>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="glass-card p-6 space-y-4 animate-slide-down">
          <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">Create Webhook Endpoint</h3>
          <input placeholder="Endpoint URL (https://...)" value={form.url} onChange={e => setForm({...form, url: e.target.value})} required className="w-full" />
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Events (comma-separated)</label>
            <textarea value={form.enabled_events} onChange={e => setForm({...form, enabled_events: e.target.value})} rows={3} className="w-full" />
          </div>
          <button type="submit" disabled={creating || !form.url} className="btn-acid px-6 py-2 rounded text-xs font-mono uppercase">
            {creating ? 'CREATING...' : 'CREATE ENDPOINT'}
          </button>
        </form>
      )}

      {loading ? (
        <div className="space-y-2">{[...Array(2)].map((_, i) => <div key={i} className="skeleton h-16 rounded" />)}</div>
      ) : webhooks.length === 0 ? (
        <div className="glass-card p-8 text-center"><p className="text-cyan/40 font-mono">[ NO WEBHOOK ENDPOINTS ]</p></div>
      ) : (
        <div className="space-y-3">
          {webhooks.map((w) => (
            <div key={w.id} className="glass-card p-4 flex items-center justify-between">
              <div className="space-y-1">
                <div className="font-mono text-sm text-cyan">{w.url}</div>
                <div className="font-mono text-xs text-cyan/50">{w.id}</div>
                <div className="flex gap-2 flex-wrap">
                  {(w.enabled_events || []).map((ev, i) => (
                    <span key={i} className="px-2 py-0.5 rounded text-[10px] font-mono bg-purple/10 text-purple border border-purple/20">{ev}</span>
                  ))}
                </div>
              </div>
              <button onClick={() => handleDelete(w.id)} className="btn-red px-3 py-1 rounded text-xs font-mono shrink-0">
                DELETE
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
