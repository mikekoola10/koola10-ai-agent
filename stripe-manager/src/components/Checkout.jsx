import React, { useState, useEffect } from 'react';
import { listProducts, createCheckout } from '../api';

export default function Checkout() {
  const [products, setProducts] = useState([]);
  const [form, setForm] = useState({ price_id: '', customer_email: '', success_url: 'https://koola10.ai/thanks', cancel_url: 'https://koola10.ai/pricing', mode: 'payment' });
  const [creating, setCreating] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    listProducts().then(d => setProducts(Array.isArray(d) ? d : [])).catch(() => {});
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setCreating(true);
    setError(null);
    setResult(null);
    try {
      const data = await createCheckout(form);
      setResult(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">💳 CHECKOUT</h1>
        <p className="text-xs text-cyan/40 mt-1 font-mono">Create payment or subscription sessions</p>
      </div>

      <form onSubmit={handleSubmit} className="glass-card p-6 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Price ID</label>
            <input placeholder="price_..." value={form.price_id} onChange={e => setForm({...form, price_id: e.target.value})} required />
          </div>
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Customer Email</label>
            <input placeholder="customer@example.com" type="email" value={form.customer_email} onChange={e => setForm({...form, customer_email: e.target.value})} />
          </div>
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Mode</label>
            <select value={form.mode} onChange={e => setForm({...form, mode: e.target.value})}>
              <option value="payment">One-time payment</option>
              <option value="subscription">Subscription</option>
            </select>
          </div>
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Success URL</label>
            <input value={form.success_url} onChange={e => setForm({...form, success_url: e.target.value})} />
          </div>
          <div>
            <label className="text-xs font-mono text-cyan/50 uppercase mb-1 block">Cancel URL</label>
            <input value={form.cancel_url} onChange={e => setForm({...form, cancel_url: e.target.value})} />
          </div>
        </div>

        <button type="submit" disabled={creating || !form.price_id} className="btn-acid px-6 py-3 rounded text-xs font-mono uppercase tracking-wider">
          {creating ? 'CREATING SESSION...' : '🚀 CREATE CHECKOUT SESSION'}
        </button>
      </form>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
        </div>
      )}

      {result && (
        <div className="glass-card p-6 animate-slide-down">
          <h3 className="text-sm font-mono text-acid uppercase tracking-wider mb-3">✓ SESSION CREATED</h3>
          <div className="space-y-2 text-sm font-mono">
            <div><span className="text-cyan/50">Session ID:</span> {result.id}</div>
            <div><span className="text-cyan/50">Status:</span> {result.status}</div>
            <div><span className="text-cyan/50">Mode:</span> {result.mode}</div>
          </div>
          {result.url && (
            <a href={result.url} target="_blank" rel="noopener noreferrer"
              className="inline-block mt-4 btn-cyan px-6 py-2 rounded text-xs font-mono uppercase">
              OPEN CHECKOUT URL →
            </a>
          )}
        </div>
      )}
    </div>
  );
}
