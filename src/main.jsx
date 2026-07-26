import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

// Error display fallback so errors are visible on screen instead of black
window.addEventListener('error', (e) => {
  const root = document.getElementById('root');
  if (root && root.children.length === 0) {
    root.innerHTML = `<div style="position:fixed;inset:0;background:#0a0a0a;color:#ff3333;padding:40px;font-family:monospace;font-size:14px;z-index:9999;overflow:auto">
      <div style="color:#39ff14;font-size:18px;margin-bottom:16px">[ LOAD ERROR ]</div>
      <div style="color:#00f0ff;margin-bottom:8px">${e.message || 'Unknown error'}</div>
      <div style="color:#8b00ff">at ${e.filename || '?'}:${e.lineno || '?'}</div>
    </div>`;
  }
});
window.addEventListener('unhandledrejection', (e) => {
  console.error('[UnhandledPromiseRejection]', e.reason);
});

try {
  ReactDOM.createRoot(document.getElementById('root')).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
} catch (err) {
  document.getElementById('root').innerHTML = `<div style="position:fixed;inset:0;background:#0a0a0a;color:#ff3333;padding:40px;font-family:monospace;font-size:14px;z-index:9999">RENDER ERROR: ${err.message}</div>`;
}
