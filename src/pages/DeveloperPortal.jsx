import { useState, useEffect } from 'react';

const ENDPOINTS = [
  {
    method: 'POST',
    path: '/api/v1/agents/sprint',
    description: 'Run a revenue sprint with AI agents',
    params: ['vertical (string)', 'agents (number)', 'duration (string)'],
  },
  {
    method: 'GET',
    path: '/api/v1/vault/summary',
    description: 'Get current financial status',
    params: [],
  },
  {
    method: 'POST',
    path: '/api/v1/agents/trigger',
    description: 'Trigger a specific agent vertical',
    params: ['vertical (string: affiliate|bounty|content|grants)'],
  },
  {
    method: 'GET',
    path: '/api/v1/health',
    description: 'Check system health status',
    params: [],
  },
  {
    method: 'POST',
    path: '/api/v1/ai/remember',
    description: 'Store a key-value pair in memory',
    params: ['key (string)', 'value (string)'],
  },
  {
    method: 'GET',
    path: '/api/v1/ai/recall',
    description: 'Recall a value by key',
    params: ['key (string)'],
  },
];

const USAGE_PLANS = [
  { name: 'Free', calls: 100, period: 'month', color: '#00f0ff' },
  { name: 'Starter', calls: 1000, period: 'month', color: '#8b00ff' },
  { name: 'Pro', calls: 10000, period: 'month', color: '#39ff14' },
  { name: 'Enterprise', calls: -1, period: 'unlimited', color: '#ffd93d' },
];

export default function DeveloperPortal({ onNavigate }) {
  const [apiKeys, setApiKeys] = useState([]);
  const [newKeyName, setNewKeyName] = useState('');
  const [usage, setUsage] = useState({ calls: 0, limit: 100 });
  const [activeTab, setActiveTab] = useState('keys');
  const [copied, setCopied] = useState(null);

  useEffect(() => {
    // Load existing keys from localStorage
    const stored = localStorage.getItem('koola10_api_keys');
    if (stored) {
      setApiKeys(JSON.parse(stored));
    }
    // Load usage
    const storedUsage = localStorage.getItem('koola10_api_usage');
    if (storedUsage) {
      setUsage(JSON.parse(storedUsage));
    }
  }, []);

  const generateKey = () => {
    if (!newKeyName.trim()) return;
    const key = `koola10_${Array.from({ length: 32 }, () =>
      'abcdefghijklmnopqrstuvwxyz0123456789'.charAt(Math.floor(Math.random() * 36))
    ).join('')}`;
    const newKey = {
      id: Date.now(),
      name: newKeyName,
      key,
      created: new Date().toISOString(),
      lastUsed: null,
      requests: 0,
    };
    const updated = [...apiKeys, newKey];
    setApiKeys(updated);
    localStorage.setItem('koola10_api_keys', JSON.stringify(updated));
    setNewKeyName('');
  };

  const revokeKey = (id) => {
    const updated = apiKeys.filter((k) => k.id !== id);
    setApiKeys(updated);
    localStorage.setItem('koola10_api_keys', JSON.stringify(updated));
  };

  const copyKey = (key) => {
    navigator.clipboard.writeText(key);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  };

  return (
    <div className="relative min-h-screen font-mono">
      {/* Header */}
      <header className="sticky top-0 z-20 glass-card border-b border-cyan/10 rounded-none">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => onNavigate('landing')}
              className="text-xs text-cyan/40 hover:text-cyan transition-colors uppercase tracking-wider"
            >
              ← Home
            </button>
            <div>
              <h1 className="text-lg font-bold text-acid uppercase tracking-[3px]">
                🔌 DEVELOPER API
              </h1>
              <p className="text-[10px] text-cyan/40 uppercase tracking-widest">
                Build with Koola10 AI Agents
              </p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-xs text-cyan/40">
              {usage.calls}/{usage.limit === -1 ? '∞' : usage.limit} calls
            </span>
            <button
              onClick={() => onNavigate('dashboard')}
              className="px-3 py-1.5 text-[10px] border border-acid/30 text-acid rounded hover:bg-acid/10 transition-all uppercase tracking-wider"
            >
              Dashboard
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* Usage Bar */}
        <div className="glass-card p-6 mb-8">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-bold text-cyan uppercase tracking-wider">API Usage</h2>
            <span className="text-xs text-cyan/50">
              {usage.limit === -1 ? 'Unlimited' : `${Math.round((usage.calls / usage.limit) * 100)}% used`}
            </span>
          </div>
          <div className="h-3 bg-black/50 rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{
                width: usage.limit === -1 ? '5%' : `${Math.min((usage.calls / usage.limit) * 100, 100)}%`,
                background: 'linear-gradient(90deg, #00f0ff, #8b00ff)',
              }}
            />
          </div>
          <div className="flex justify-between mt-2">
            <span className="text-[10px] text-cyan/40">{usage.calls} requests</span>
            <span className="text-[10px] text-cyan/40">
              {usage.limit === -1 ? '∞' : `${usage.limit - usage.calls} remaining`}
            </span>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-4 mb-8 border-b border-cyan/10 pb-4">
          {[
            { id: 'keys', label: '🔑 API Keys' },
            { id: 'docs', label: '📖 Documentation' },
            { id: 'examples', label: '💻 Examples' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 text-xs uppercase tracking-wider rounded transition-all ${
                activeTab === tab.id
                  ? 'bg-acid/10 text-acid border border-acid/30'
                  : 'text-cyan/50 hover:text-cyan'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        {activeTab === 'keys' && (
          <div className="space-y-6">
            {/* Create new key */}
            <div className="glass-card p-6">
              <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">
                Create New API Key
              </h3>
              <div className="flex gap-3">
                <input
                  type="text"
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                  placeholder="Key name (e.g., production, staging)"
                  className="flex-1 px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                />
                <button
                  onClick={generateKey}
                  className="px-6 py-2 bg-acid text-black font-bold uppercase tracking-wider rounded hover:bg-acid/90 transition-all text-xs"
                >
                  Generate Key
                </button>
              </div>
            </div>

            {/* Existing keys */}
            <div className="glass-card p-6">
              <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">
                Your API Keys
              </h3>
              {apiKeys.length === 0 ? (
                <p className="text-xs text-cyan/40 text-center py-8">
                  No API keys yet. Create one above to get started.
                </p>
              ) : (
                <div className="space-y-3">
                  {apiKeys.map((k) => (
                    <div
                      key={k.id}
                      className="flex items-center justify-between p-4 bg-black/30 rounded border border-cyan/10 hover:border-cyan/30 transition-all"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-3 mb-1">
                          <span className="text-sm font-bold text-cyan">{k.name}</span>
                          <span className="text-[10px] text-cyan/40 px-2 py-0.5 bg-cyan/5 rounded">
                            {k.requests} requests
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <code className="text-[10px] text-cyan/60 font-mono truncate">
                            {k.key.substring(0, 20)}...{k.key.substring(k.key.length - 8)}
                          </code>
                          <button
                            onClick={() => copyKey(k.key)}
                            className="text-[10px] text-acid hover:text-acid/80 transition-colors"
                          >
                            {copied === k.key ? '✓ Copied' : 'Copy'}
                          </button>
                        </div>
                        <div className="text-[10px] text-cyan/30 mt-1">
                          Created: {new Date(k.created).toLocaleDateString()}
                        </div>
                      </div>
                      <button
                        onClick={() => revokeKey(k.id)}
                        className="ml-4 px-3 py-1.5 text-[10px] border border-red-500/30 text-red-400 rounded hover:bg-red-500/10 transition-all uppercase tracking-wider"
                      >
                        Revoke
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'docs' && (
          <div className="space-y-4">
            <div className="glass-card p-6">
              <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">
                Base URL
              </h3>
              <code className="text-xs text-acid bg-black/50 px-3 py-2 rounded block">
                https://koola10-ai-agent.onrender.com
              </code>
            </div>

            <div className="glass-card p-6">
              <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">
                Authentication
              </h3>
              <p className="text-xs text-cyan/60 mb-3">
                Include your API key in the Authorization header:
              </p>
              <code className="text-xs text-acid bg-black/50 px-3 py-2 rounded block">
                Authorization: Bearer YOUR_API_KEY
              </code>
            </div>

            <div className="glass-card p-6">
              <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">
                Endpoints
              </h3>
              <div className="space-y-3">
                {ENDPOINTS.map((ep) => (
                  <div
                    key={ep.path}
                    className="p-4 bg-black/30 rounded border border-cyan/10 hover:border-cyan/30 transition-all"
                  >
                    <div className="flex items-center gap-3 mb-2">
                      <span
                        className={`px-2 py-0.5 text-[10px] font-bold rounded ${
                          ep.method === 'GET'
                            ? 'bg-green-500/20 text-green-400'
                            : 'bg-blue-500/20 text-blue-400'
                        }`}
                      >
                        {ep.method}
                      </span>
                      <code className="text-xs text-cyan font-mono">{ep.path}</code>
                    </div>
                    <p className="text-[10px] text-cyan/50 mb-2">{ep.description}</p>
                    {ep.params.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {ep.params.map((p) => (
                          <span key={p} className="text-[10px] text-cyan/40 bg-cyan/5 px-2 py-0.5 rounded">
                            {p}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'examples' && (
          <div className="space-y-6">
            {/* Python */}
            <div className="glass-card p-6">
              <div className="flex items-center gap-2 mb-4">
                <span className="text-lg">🐍</span>
                <h3 className="text-sm font-bold text-cyan uppercase tracking-wider">Python</h3>
              </div>
              <pre className="text-xs text-acid/80 bg-black/50 p-4 rounded overflow-x-auto">
{`import requests

API_KEY = "your-api-key"
BASE_URL = "https://koola10-ai-agent.onrender.com"

# Run a revenue sprint
response = requests.post(
    f"{BASE_URL}/api/v1/agents/sprint",
    headers={"Authorization": f"Bearer {API_KEY}"},
    json={"vertical": "affiliate", "agents": 10}
)
print(response.json())`}
              </pre>
            </div>

            {/* JavaScript */}
            <div className="glass-card p-6">
              <div className="flex items-center gap-2 mb-4">
                <span className="text-lg">⚡</span>
                <h3 className="text-sm font-bold text-cyan uppercase tracking-wider">JavaScript</h3>
              </div>
              <pre className="text-xs text-acid/80 bg-black/50 p-4 rounded overflow-x-auto">
{`const API_KEY = "your-api-key";
const BASE_URL = "https://koola10-ai-agent.onrender.com";

// Run a revenue sprint
const response = await fetch(\`\${BASE_URL}/api/v1/agents/sprint\`, {
  method: "POST",
  headers: {
    "Authorization": \`Bearer \${API_KEY}\`,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({ vertical: "affiliate", agents: 10 })
});
const data = await response.json();
console.log(data);`}
              </pre>
            </div>

            {/* cURL */}
            <div className="glass-card p-6">
              <div className="flex items-center gap-2 mb-4">
                <span className="text-lg">🔧</span>
                <h3 className="text-sm font-bold text-cyan uppercase tracking-wider">cURL</h3>
              </div>
              <pre className="text-xs text-acid/80 bg-black/50 p-4 rounded overflow-x-auto">
{`curl -X POST https://koola10-ai-agent.onrender.com/api/v1/agents/sprint \\
  -H "Authorization: Bearer your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{"vertical": "affiliate", "agents": 10}'`}
              </pre>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
