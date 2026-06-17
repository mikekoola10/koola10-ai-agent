import React, { createContext, useState, useEffect } from 'react';

export const TelemetryContext = createContext();

export const TelemetryProvider = ({ children }) => {
  const [data, setData] = useState({ balance: 0, revenue: 0, status: 'connecting' });
  const [logs, setLogs] = useState([]);

  useEffect(() => {
    const ws = new WebSocket('wss://koola10.fly.dev/ws');

    ws.onopen = () => setData(prev => ({ ...prev, status: 'online' }));
    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      if (msg.type === 'init' || msg.type === 'transaction') {
        setData(prev => ({
          ...prev,
          balance: msg.balance || prev.balance,
          revenue: msg.revenue || msg.total_revenue || prev.revenue
        }));
      } else if (msg.type === 'audit') {
        setLogs(prev => [msg.entry, ...prev].slice(0, 50));
      }
    };
    ws.onerror = () => setData(prev => ({ ...prev, status: 'error' }));
    ws.onclose = () => setData(prev => ({ ...prev, status: 'offline' }));

    return () => ws.close();
  }, []);

  return (
    <TelemetryContext.Provider value={{ data, logs }}>
      {children}
    </TelemetryContext.Provider>
  );
};
