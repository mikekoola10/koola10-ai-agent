import { useState } from 'react';
import { apiUrl, ADMIN_KEY } from '../api';

export default function MemoryPanel() {
  const [rememberKey, setRememberKey] = useState('');
  const [rememberValue, setRememberValue] = useState('');
  const [rememberResult, setRememberResult] = useState(null);
  const [rememberLoading, setRememberLoading] = useState(false);

  const [recallKey, setRecallKey] = useState('');
  const [recallResult, setRecallResult] = useState(null);
  const [recallLoading, setRecallLoading] = useState(false);

  const handleRemember = async (e) => {
    e.preventDefault();
    if (!rememberKey.trim()) return;
    setRememberLoading(true);
    setRememberResult(null);
    try {
      const res = await fetch(apiUrl('koola10', '/ai/remember'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}` } : {}),
        },
        body: JSON.stringify({ key: rememberKey.trim(), value: rememberValue }),
      });
      const data = await res.json().catch(() => ({}));
      setRememberResult({ ok: res.ok, message: data.message || (res.ok ? 'Stored ✓' : 'Failed') });
      if (res.ok) {
        setRememberKey('');
        setRememberValue('');
      }
    } catch (e) {
      setRememberResult({ ok: false, message: e.message });
    } finally {
      setRememberLoading(false);
    }
  };

  const handleRecall = async (e) => {
    e.preventDefault();
    if (!recallKey.trim()) return;
    setRecallLoading(true);
    setRecallResult(null);
    try {
      const res = await fetch(apiUrl('koola10', `/ai/recall?key=${encodeURIComponent(recallKey.trim())}`), {
        headers: ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}` } : {},
      });
      const data = await res.json().catch(() => ({}));
      setRecallResult({
        ok: res.ok,
        value: data.value || data.message || 'Not found',
      });
    } catch (e) {
      setRecallResult({ ok: false, value: e.message });
    } finally {
      setRecallLoading(false);
    }
  };

  return (
    <div className="glass-card p-6">
      <h2 className="text-lg font-bold font-mono text-cyan uppercase tracking-wider mb-4">
        🧠 MEMORY (MIRROR)
      </h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Remember */}
        <div>
          <h3 className="text-sm font-mono text-cyan/70 uppercase mb-3 tracking-wider">STORE</h3>
          <form onSubmit={handleRemember} className="flex flex-col gap-2">
            <input
              type="text"
              placeholder="key"
              value={rememberKey}
              onChange={(e) => setRememberKey(e.target.value)}
              className="w-full font-mono text-sm"
            />
            <input
              type="text"
              placeholder="value"
              value={rememberValue}
              onChange={(e) => setRememberValue(e.target.value)}
              className="w-full font-mono text-sm"
            />
            <button
              type="submit"
              disabled={rememberLoading || !rememberKey.trim()}
              className="btn-cyan px-4 py-2 rounded text-xs font-mono uppercase tracking-wider cursor-pointer disabled:cursor-not-allowed"
            >
              {rememberLoading ? 'STORING...' : '💾 REMEMBER'}
            </button>
          </form>
          {rememberResult && (
            <p
              className={`text-xs font-mono mt-2 px-2 py-1 rounded animate-slide-down ${
                rememberResult.ok ? 'text-acid bg-acid/10' : 'text-red-400 bg-red-400/10'
              }`}
            >
              {rememberResult.ok ? '✓' : '✗'} {rememberResult.message}
            </p>
          )}
        </div>

        {/* Recall */}
        <div>
          <h3 className="text-sm font-mono text-purple/70 uppercase mb-3 tracking-wider">RECALL</h3>
          <form onSubmit={handleRecall} className="flex flex-col gap-2">
            <input
              type="text"
              placeholder="key"
              value={recallKey}
              onChange={(e) => setRecallKey(e.target.value)}
              className="w-full font-mono text-sm"
            />
            <button
              type="submit"
              disabled={recallLoading || !recallKey.trim()}
              className="btn-purple px-4 py-2 rounded text-xs font-mono uppercase tracking-wider cursor-pointer disabled:cursor-not-allowed"
            >
              {recallLoading ? 'SEARCHING...' : '🔍 RECALL'}
            </button>
          </form>
          {recallResult && (
            <div
              className={`text-xs font-mono mt-2 p-3 rounded animate-slide-down ${
                recallResult.ok ? 'text-cyan bg-cyan/10 border border-cyan/20' : 'text-red-400 bg-red-400/10'
              }`}
            >
              <p className="opacity-50 uppercase mb-1">VALUE:</p>
              <p className="break-all">{recallResult.value}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
