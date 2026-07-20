import React, { useState, useEffect } from 'react';
import { getRevenue } from '../api';

export default function Settings() {
  const [connected, setConnected] = useState(null);
  const [testing, setTesting] = useState(false);

  const testConnection = async () => {
    setTesting(true);
    try {
      await getRevenue();
      setConnected(true);
    } catch {
      setConnected(false);
    } finally {
      setTesting(false);
    }
  };

  useEffect(() => { testConnection(); }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">⚙️ SETTINGS</h1>
        <p className="text-xs text-cyan/40 mt-1 font-mono">Stripe API connection status</p>
      </div>

      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">API Connection</h3>
            <p className="text-xs text-cyan/40 mt-1">Tests the STRIPE_API_KEY on the backend</p>
          </div>
          <div className="flex items-center gap-3">
            <span className={`inline-flex items-center gap-2 text-xs font-mono ${connected === true ? 'text-acid' : connected === false ? 'text-red-400' : 'text-yellow-400'}`}>
              <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: connected === true ? '#39ff14' : connected === false ? '#ff3333' : '#ffaa00' }} />
              {connected === true ? 'CONNECTED' : connected === false ? 'FAILED' : 'CHECKING'}
            </span>
            <button onClick={testConnection} disabled={testing} className="btn-cyan px-3 py-1 rounded text-xs font-mono">
              {testing ? 'TESTING...' : 'TEST'}
            </button>
          </div>
        </div>
      </div>

      <div className="glass-card p-6 space-y-3">
        <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">Environment Variables</h3>
        <p className="text-xs text-cyan/50 font-mono">These are set in Render → Environment for koola10-ai-agent:</p>
        <div className="space-y-2 text-sm font-mono">
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_API_KEY</span>
            <span className="text-acid">✓ Required</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_WEBHOOK_SECRET</span>
            <span className="text-yellow-400">○ Optional</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_PRICE_GRANT</span>
            <span className="text-acid">✓ Required</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_PRICE_AFFILIATE</span>
            <span className="text-cyan/40">○ Optional</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_PRICE_BOUNTY</span>
            <span className="text-cyan/40">○ Optional</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan/70 w-48">STRIPE_PRICE_CONTENT</span>
            <span className="text-cyan/40">○ Optional</span>
          </div>
        </div>
      </div>

      <div className="glass-card p-6 space-y-3">
        <h3 className="text-sm font-mono text-cyan uppercase tracking-wider">Backend Endpoints</h3>
        <div className="space-y-1 text-xs font-mono text-cyan/60">
          <div>GET  /admin/stripe/products</div>
          <div>POST /admin/stripe/products</div>
          <div>GET  /admin/stripe/customers</div>
          <div>GET  /admin/stripe/subscriptions</div>
          <div>GET  /admin/stripe/payments</div>
          <div>POST /admin/stripe/checkout</div>
          <div>GET  /admin/stripe/webhooks</div>
          <div>POST /admin/stripe/webhooks</div>
          <div>GET  /admin/stripe/revenue</div>
        </div>
      </div>
    </div>
  );
}
