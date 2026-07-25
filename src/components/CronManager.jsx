import React, { useState, useEffect, useCallback } from 'react';
import { apiUrl, authHeaders } from '../api';

export default function CronManager() {
  const [jobs, setJobs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [status, setStatus] = useState(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [setupRunning, setSetupRunning] = useState(false);
  const [selectedJob, setSelectedJob] = useState(null);
  const [history, setHistory] = useState(null);
  const [historyLoading, setHistoryLoading] = useState(false);

  const [form, setForm] = useState({
    title: '',
    url: '',
    frequency: 'every-6-hours',
    method: 'POST',
    timezone: 'America/New_York',
  });

  const frequencies = [
    { value: 'every-minute', label: 'Every Minute' },
    { value: 'every-5-min', label: 'Every 5 Minutes' },
    { value: 'every-10-min', label: 'Every 10 Minutes' },
    { value: 'hourly', label: 'Hourly' },
    { value: 'every-6-hours', label: 'Every 6 Hours' },
    { value: 'daily', label: 'Daily (6 AM)' },
    { value: 'daily-morning', label: 'Daily (8 AM)' },
    { value: 'daily-evening', label: 'Daily (8 PM)' },
    { value: 'weekly', label: 'Weekly (Monday)' },
  ];

  const quickPresets = [
    { title: 'Revenue Sprint', url: '/admin/run-scheduled-sprint', frequency: 'every-6-hours', method: 'POST', description: 'Run full revenue sprint every 6 hours' },
    { title: 'Health Check', url: '/health', frequency: 'every-10-min', method: 'GET', description: 'Monitor system health every 10 minutes' },
    { title: 'Vault Report', url: '/vault/summary', frequency: 'daily', method: 'GET', description: 'Daily financial summary at 6 AM' },
    { title: 'Grant Monitor', url: '/grants/monitor', frequency: 'every-6-hours', method: 'POST', description: 'Check grant application statuses' },
  ];

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const base = apiUrl('koola10', '/admin/cron/jobs');
      const res = await fetch(base, { headers: authHeaders() });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || data.message || `HTTP ${res.status}`);
      }
      const data = await res.json();
      setJobs(data.jobs || []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchStatus = useCallback(async () => {
    try {
      const base = apiUrl('koola10', '/admin/cron/status');
      const res = await fetch(base, { headers: authHeaders() });
      const data = await res.json();
      setStatus(data);
    } catch (e) {
      setStatus({ configured: false, error: e.message });
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    fetchJobs();
  }, [fetchStatus, fetchJobs]);

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      const base = apiUrl('koola10', '/admin/cron/jobs');
      const res = await fetch(base, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          ...form,
          url: form.url.startsWith('http') ? form.url : `https://koola10-ai-agent.onrender.com${form.url}`,
        }),
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || `HTTP ${res.status}`);
      }
      setShowCreate(false);
      setForm({ title: '', url: '', frequency: 'every-6-hours', method: 'POST', timezone: 'America/New_York' });
      fetchJobs();
    } catch (e) {
      setError(e.message);
    } finally {
      setCreating(false);
    }
  };

  const handleToggle = async (jobId, enabled) => {
    try {
      const base = apiUrl('koola10', `/admin/cron/jobs?jobId=${jobId}`);
      const res = await fetch(base, {
        method: 'PATCH',
        headers: authHeaders(),
        body: JSON.stringify({ enabled: !enabled }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      fetchJobs();
    } catch (e) {
      setError(e.message);
    }
  };

  const handleDelete = async (jobId) => {
    if (!confirm(`Delete cron job ${jobId}?`)) return;
    try {
      const base = apiUrl('koola10', `/admin/cron/jobs?jobId=${jobId}`);
      const res = await fetch(base, { method: 'DELETE', headers: authHeaders() });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      fetchJobs();
    } catch (e) {
      setError(e.message);
    }
  };

  const handleHistory = async (jobId) => {
    setSelectedJob(jobId);
    setHistoryLoading(true);
    setHistory(null);
    try {
      const base = apiUrl('koola10', `/admin/cron/history?jobId=${jobId}`);
      const res = await fetch(base, { headers: authHeaders() });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setHistory(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleSetupDefaults = async () => {
    setSetupRunning(true);
    setError(null);
    try {
      const base = apiUrl('koola10', '/admin/cron/setup');
      const res = await fetch(base, { method: 'POST', headers: authHeaders() });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || data.message || `HTTP ${res.status}`);
      fetchJobs();
    } catch (e) {
      setError(e.message);
    } finally {
      setSetupRunning(false);
    }
  };

  const handlePreset = (preset) => {
    setForm({
      title: preset.title,
      url: preset.url,
      frequency: preset.frequency,
      method: preset.method,
      timezone: 'America/New_York',
    });
    setShowCreate(true);
  };

  const formatTimestamp = (ts) => {
    if (!ts) return '—';
    return new Date(ts * 1000).toLocaleString();
  };

  const statusColor = (status) => {
    if (status === 1) return 'text-acid';
    if (status === 2) return 'text-red-400';
    return 'text-cyan/50';
  };

  const statusLabel = (status) => {
    if (status === 0) return 'UNKNOWN';
    if (status === 1) return 'OK';
    if (status === 2) return 'FAILED';
    return 'N/A';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono text-cyan uppercase tracking-[3px]">⏰ CRON SCHEDULER</h1>
          <p className="text-xs text-cyan/40 mt-1 font-mono">
            {status?.configured ? `Connected to cron-job.org • ${jobs.length} jobs` : 'Not configured'}
          </p>
        </div>
        <div className="flex gap-3">
          <button onClick={fetchJobs} className="text-xs text-cyan/50 hover:text-cyan font-mono">[ REFRESH ]</button>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="btn-cyan px-4 py-2 rounded text-xs font-mono uppercase"
          >
            {showCreate ? '✕ CANCEL' : '+ NEW JOB'}
          </button>
        </div>
      </div>

      {/* Status Banner */}
      {status && !status.configured && (
        <div className="glass-card p-6 border border-yellow-400/30">
          <div className="flex items-start gap-4">
            <div className="text-2xl">🔑</div>
            <div className="flex-1">
              <h3 className="text-sm font-bold text-yellow-400 uppercase tracking-wider mb-2">Setup Required</h3>
              <p className="text-xs text-cyan/60 mb-3">
                Add <code className="text-acid">CRONJOB_API_KEY</code> to your Render environment variables to enable cron scheduling.
              </p>
              <ol className="text-xs text-cyan/50 space-y-1 mb-4">
                <li>1. Go to <a href="https://cron-job.org" target="_blank" rel="noopener" className="text-cyan hover:text-acid underline">cron-job.org</a> and create a free account</li>
                <li>2. Go to Settings → API Keys and generate a key</li>
                <li>3. Add <code className="text-acid">CRONJOB_API_KEY</code> in your Render dashboard environment variables</li>
                <li>4. Click "Setup Default Jobs" below to create the standard schedule</li>
              </ol>
              <button onClick={handleSetupDefaults} disabled={setupRunning} className="btn-acid px-4 py-2 rounded text-xs font-mono uppercase">
                {setupRunning ? 'CREATING...' : '🚀 SETUP DEFAULT JOBS'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="glass-card p-4 text-red-400 text-sm font-mono" style={{ borderColor: '#ff333366' }}>
          [ ERR: {error} ]
          <button onClick={() => setError(null)} className="ml-4 text-cyan/50 hover:text-cyan">dismiss</button>
        </div>
      )}

      {/* Quick Presets */}
      {status?.configured && (
        <div className="glass-card p-6">
          <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-4">⚡ Quick Setup</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
            {quickPresets.map((preset, i) => (
              <button
                key={i}
                onClick={() => handlePreset(preset)}
                className="p-4 bg-black/30 rounded border border-cyan/10 hover:border-cyan/30 text-left transition-all group"
              >
                <div className="text-xs font-bold text-cyan group-hover:text-acid transition-colors mb-1">{preset.title}</div>
                <div className="text-[10px] text-cyan/40">{preset.description}</div>
                <div className="text-[10px] text-cyan/30 mt-2 font-mono">{preset.frequency}</div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Create Form */}
      {showCreate && (
        <form onSubmit={handleCreate} className="glass-card p-6 space-y-4 animate-slide-down">
          <h3 className="text-sm font-bold text-cyan uppercase tracking-wider">Create Cron Job</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-[10px] text-cyan/50 uppercase mb-1 block font-mono">Job Title</label>
              <input
                type="text"
                value={form.title}
                onChange={e => setForm({ ...form, title: e.target.value })}
                placeholder="e.g., Revenue Sprint"
                required
                className="w-full"
              />
            </div>
            <div>
              <label className="text-[10px] text-cyan/50 uppercase mb-1 block font-mono">Endpoint URL</label>
              <input
                type="text"
                value={form.url}
                onChange={e => setForm({ ...form, url: e.target.value })}
                placeholder="/admin/run-scheduled-sprint"
                required
                className="w-full"
              />
            </div>
            <div>
              <label className="text-[10px] text-cyan/50 uppercase mb-1 block font-mono">Frequency</label>
              <select value={form.frequency} onChange={e => setForm({ ...form, frequency: e.target.value })} className="w-full">
                {frequencies.map(f => (
                  <option key={f.value} value={f.value}>{f.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-[10px] text-cyan/50 uppercase mb-1 block font-mono">HTTP Method</label>
              <select value={form.method} onChange={e => setForm({ ...form, method: e.target.value })} className="w-full">
                <option value="GET">GET</option>
                <option value="POST">POST</option>
              </select>
            </div>
          </div>
          <button type="submit" disabled={creating} className="btn-acid px-6 py-2 rounded text-xs font-mono uppercase">
            {creating ? 'CREATING...' : 'CREATE JOB'}
          </button>
        </form>
      )}

      {/* Jobs List */}
      {loading ? (
        <div className="space-y-2">{[...Array(3)].map((_, i) => <div key={i} className="skeleton h-16 rounded" />)}</div>
      ) : jobs.length === 0 ? (
        <div className="glass-card p-8 text-center">
          <p className="text-cyan/40 font-mono">[ NO CRON JOBS ]</p>
          <p className="text-xs text-cyan/30 mt-2">Create a job above or use Quick Setup to get started.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {jobs.map((job) => (
            <div
              key={job.jobId}
              className={`glass-card p-4 transition-all ${selectedJob === job.jobId ? 'ring-1 ring-acid/40' : ''}`}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <span className={`w-2.5 h-2.5 rounded-full ${job.enabled ? 'bg-acid animate-pulse' : 'bg-red-400'}`} />
                    <span className="text-sm font-bold text-cyan font-mono">{job.title || `Job #${job.jobId}`}</span>
                    <span className={`text-[10px] font-mono px-2 py-0.5 rounded ${job.enabled ? 'bg-acid/10 text-acid border border-acid/20' : 'bg-red-400/10 text-red-400 border border-red-400/20'}`}>
                      {job.enabled ? 'ENABLED' : 'DISABLED'}
                    </span>
                  </div>
                  <div className="text-[10px] text-cyan/50 font-mono truncate ml-5">{job.url}</div>
                  <div className="flex gap-4 mt-1 ml-5 text-[10px] text-cyan/30 font-mono">
                    <span>ID: {job.jobId}</span>
                    {job.nextExecution && <span>Next: {formatTimestamp(job.nextExecution)}</span>}
                    {job.lastExecution > 0 && (
                      <span className={statusColor(job.lastStatus)}>
                        Last: {statusLabel(job.lastStatus)} ({job.lastDuration}ms)
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <button
                    onClick={() => handleHistory(job.jobId)}
                    className="px-3 py-1.5 text-[10px] border border-cyan/20 text-cyan/60 rounded hover:bg-cyan/5 transition-all font-mono"
                  >
                    HISTORY
                  </button>
                  <button
                    onClick={() => handleToggle(job.jobId, job.enabled)}
                    className={`px-3 py-1.5 text-[10px] rounded transition-all font-mono ${
                      job.enabled
                        ? 'border border-yellow-400/30 text-yellow-400 hover:bg-yellow-400/10'
                        : 'border border-acid/30 text-acid hover:bg-acid/10'
                    }`}
                  >
                    {job.enabled ? 'PAUSE' : 'RESUME'}
                  </button>
                  <button
                    onClick={() => handleDelete(job.jobId)}
                    className="px-3 py-1.5 text-[10px] border border-red-400/30 text-red-400 rounded hover:bg-red-400/10 transition-all font-mono"
                  >
                    DELETE
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* History Panel */}
      {selectedJob && (
        <div className="glass-card p-6 animate-slide-down">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-bold text-cyan uppercase tracking-wider">
              Execution History — Job #{selectedJob}
            </h3>
            <button onClick={() => { setSelectedJob(null); setHistory(null); }} className="text-xs text-cyan/40 hover:text-cyan">✕ Close</button>
          </div>
          {historyLoading ? (
            <div className="skeleton h-24 rounded" />
          ) : history && history.history && history.history.length > 0 ? (
            <div className="overflow-x-auto">
              <table>
                <thead>
                  <tr>
                    <th>Status</th>
                    <th>HTTP</th>
                    <th>Planned</th>
                    <th>Actual</th>
                    <th>Duration</th>
                    <th>URL</th>
                  </tr>
                </thead>
                <tbody>
                  {history.history.slice(0, 10).map((item) => (
                    <tr key={item.jobLogId}>
                      <td>
                        <span className={`text-[10px] font-mono ${item.status === 1 ? 'text-acid' : 'text-red-400'}`}>
                          {statusLabel(item.status)}
                        </span>
                      </td>
                      <td className="text-[10px] font-mono text-cyan/60">{item.httpStatus || '—'}</td>
                      <td className="text-[10px] font-mono text-cyan/50">{formatTimestamp(item.datePlanned)}</td>
                      <td className="text-[10px] font-mono text-cyan/50">{formatTimestamp(item.date)}</td>
                      <td className="text-[10px] font-mono text-cyan/50">{item.duration}ms</td>
                      <td className="text-[10px] font-mono text-cyan/40 truncate max-w-xs">{item.url}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {history.predictions && history.predictions.length > 0 && (
                <div className="mt-3 text-[10px] text-cyan/30 font-mono">
                  Next predicted runs: {history.predictions.map(formatTimestamp).join(', ')}
                </div>
              )}
            </div>
          ) : (
            <p className="text-xs text-cyan/40 text-center py-4 font-mono">No execution history yet.</p>
          )}
        </div>
      )}
    </div>
  );
}
