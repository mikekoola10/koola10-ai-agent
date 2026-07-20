import React, { useState, useEffect } from 'react';
import { listCustomers, getCustomer } from '../api';

export default function Customers() {
  const [customers, setCustomers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selected, setSelected] = useState(null);

  const fetchCustomers = async () => {
    setLoading(true);
    try {
      const data = await listCustomers();
      setCustomers(Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchCustomers(); }, []);

  const viewCustomer = async (id) => {
    try {
      const data = await getCustomer(id);
      setSelected(data);
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">👥 CUSTOMERS</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">{customers.length} customers</p>
        </div>
        <button onClick={fetchCustomers} className="text-xs text-cyan/50 hover:text-cyan font-mono">[ REFRESH ]</button>
      </div>

      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
          <button onClick={() => setError(null)} className="ml-4 text-cyan/50 hover:text-cyan">dismiss</button>
        </div>
      )}

      {selected && (
        <div className="glass-card p-6 animate-slide-down">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">Customer Detail</h3>
            <button onClick={() => setSelected(null)} className="text-cyan/50 hover:text-cyan text-xs font-mono">✕ CLOSE</button>
          </div>
          <div className="grid grid-cols-2 gap-4 text-sm font-mono">
            <div><span className="text-cyan/50">ID:</span> {selected.id}</div>
            <div><span className="text-cyan/50">Email:</span> {selected.email || '---'}</div>
            <div><span className="text-cyan/50">Name:</span> {selected.name || '---'}</div>
            <div><span className="text-cyan/50">Phone:</span> {selected.phone || '---'}</div>
            <div><span className="text-cyan/50">Subscriptions:</span> {selected.subscriptions?.length || 0}</div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => <div key={i} className="skeleton h-12 rounded" />)}
        </div>
      ) : customers.length === 0 ? (
        <div className="glass-card p-8 text-center">
          <p className="text-cyan/40 font-mono">[ NO CUSTOMERS YET ]</p>
        </div>
      ) : (
        <div className="glass-card overflow-hidden">
          <table>
            <thead>
              <tr className="border-b border-cyan/10">
                <th>Email</th>
                <th>Name</th>
                <th>ID</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {customers.map((c) => (
                <tr key={c.id} className="hover:bg-cyan/5">
                  <td className="font-mono text-cyan">{c.email || '---'}</td>
                  <td className="font-mono text-cyan/70">{c.name || '---'}</td>
                  <td className="font-mono text-xs text-cyan/50">{c.id}</td>
                  <td className="font-mono text-xs text-cyan/60">
                    {c.created ? new Date(c.created * 1000).toLocaleDateString() : '---'}
                  </td>
                  <td>
                    <button onClick={() => viewCustomer(c.id)} className="btn-cyan px-3 py-1 rounded text-xs font-mono">
                      VIEW
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
