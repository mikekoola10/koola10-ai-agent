import React, { useState, useEffect } from 'react';
import { listProducts, createProduct, deleteProduct } from '../api';

export default function Products() {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', description: '', price: '', mode: 'one_time' });
  const [creating, setCreating] = useState(false);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      const data = await listProducts();
      setProducts(Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchProducts(); }, []);

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreating(true);
    try {
      await createProduct({
        name: form.name,
        description: form.description,
        price: form.price ? parseInt(form.price) * 100 : 0,
        mode: form.mode,
        interval: form.mode === 'recurring' ? 'month' : undefined,
      });
      setForm({ name: '', description: '', price: '', mode: 'one_time' });
      setShowCreate(false);
      fetchProducts();
    } catch (e) {
      setError(e.message);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this product?')) return;
    try {
      await deleteProduct(id);
      fetchProducts();
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">📦 PRODUCTS</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">{products.length} products</p>
        </div>
        <div className="flex gap-3">
          <button onClick={fetchProducts} className="text-xs text-cyan/50 hover:text-cyan font-mono">[ REFRESH ]</button>
          <button onClick={() => setShowCreate(!showCreate)} className="btn-cyan px-4 py-2 rounded text-xs font-mono uppercase">
            {showCreate ? '✕ CANCEL' : '+ NEW PRODUCT'}
          </button>
        </div>
      </div>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
          <button onClick={() => setError(null)} className="ml-4 text-cyan/50 hover:text-cyan">dismiss</button>
        </div>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="glass-card p-6 space-y-4 animate-slide-down">
          <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">Create Product</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <input placeholder="Product name" value={form.name} onChange={e => setForm({...form, name: e.target.value})} required />
            <input placeholder="Description" value={form.description} onChange={e => setForm({...form, description: e.target.value})} />
            <input placeholder="Price (USD)" type="number" step="0.01" value={form.price} onChange={e => setForm({...form, price: e.target.value})} />
            <select value={form.mode} onChange={e => setForm({...form, mode: e.target.value})}>
              <option value="one_time">One-time payment</option>
              <option value="recurring">Recurring (subscription)</option>
            </select>
          </div>
          <button type="submit" disabled={creating || !form.name} className="btn-acid px-6 py-2 rounded text-xs font-mono uppercase">
            {creating ? 'CREATING...' : 'CREATE PRODUCT'}
          </button>
        </form>
      )}

      {loading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => <div key={i} className="skeleton h-16 rounded" />)}
        </div>
      ) : products.length === 0 ? (
        <div className="glass-card p-8 text-center">
          <p className="text-cyan/40 font-mono">[ NO PRODUCTS YET ]</p>
        </div>
      ) : (
        <div className="glass-card overflow-hidden">
          <table>
            <thead>
              <tr className="border-b border-cyan/10">
                <th>Name</th>
                <th>ID</th>
                <th>Active</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id} className="hover:bg-cyan/5">
                  <td className="font-mono font-bold text-cyan">{p.name}</td>
                  <td className="font-mono text-xs text-cyan/50">{p.id}</td>
                  <td>
                    <span className={`px-2 py-0.5 rounded text-xs ${p.active ? 'text-acid bg-acid/10 border border-acid/20' : 'text-red-400 bg-red-400/10 border border-red-400/20'}`}>
                      {p.active ? 'ACTIVE' : 'INACTIVE'}
                    </span>
                  </td>
                  <td className="font-mono text-xs text-cyan/60">
                    {p.created ? new Date(p.created * 1000).toLocaleDateString() : '---'}
                  </td>
                  <td>
                    <button onClick={() => handleDelete(p.id)} className="btn-red px-3 py-1 rounded text-xs font-mono">
                      DELETE
                    </button>
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
