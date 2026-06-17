import React, { useState, useEffect } from 'react';
import { StyleSheet, Text, View, ScrollView } from 'react-native';
import { StatusBar } from 'expo-status-bar';

export default function App() {
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
        setLogs(prev => [msg.entry, ...prev].slice(0, 20));
      }
    };
    ws.onerror = (e) => setData(prev => ({ ...prev, status: 'error' }));
    ws.onclose = () => setData(prev => ({ ...prev, status: 'offline' }));

    return () => ws.close();
  }, []);

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>KOOLA10_MOBILE</Text>
        <Text style={[styles.status, { color: data.status === 'online' ? '#0f0' : '#f00' }]}>
          {data.status.toUpperCase()}
        </Text>
      </View>

      <View style={styles.card}>
        <Text style={styles.label}>TOTAL_REVENUE</Text>
        <Text style={styles.value}>${data.revenue.toFixed(2)}</Text>
      </View>

      <View style={styles.card}>
        <Text style={styles.label}>CURRENT_BALANCE</Text>
        <Text style={styles.value}>${data.balance.toFixed(2)}</Text>
      </View>

      <Text style={styles.sectionTitle}>SYSTEM_AUDIT_LOG</Text>
      <ScrollView style={styles.logs}>
        {logs.map((log, i) => (
          <Text key={i} style={styles.logItem}>
            [{log.timestamp.split('T')[1].split('.')[0]}] {log.action}
          </Text>
        ))}
      </ScrollView>

      <StatusBar style="light" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#000', padding: 20, paddingTop: 60 },
  header: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 30 },
  title: { color: '#fff', fontSize: 20, fontWeight: 'bold', letterSpacing: 2 },
  status: { fontWeight: 'bold' },
  card: { backgroundColor: '#111', padding: 20, borderRadius: 10, marginBottom: 15, borderWidth: 1, borderColor: '#333' },
  label: { color: '#aaa', fontSize: 12, marginBottom: 5 },
  value: { color: '#fff', fontSize: 28, fontWeight: 'bold' },
  sectionTitle: { color: '#666', marginTop: 20, marginBottom: 10, fontSize: 14 },
  logs: { flex: 1 },
  logItem: { color: '#0f0', fontFamily: 'monospace', fontSize: 11, marginBottom: 4 }
});
