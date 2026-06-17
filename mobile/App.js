import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import WebSocket from 'react-native-websocket';

export default function App() {
  const [data, setData] = useState({ revenue: 0, balance: 0, health: 'unknown' });

  useEffect(() => {
    const ws = new WebSocket('wss://koola10.fly.dev/ws');
    ws.onopen = () => console.log('Connected to Koola10');
    ws.onmessage = (e) => {
      try {
        const payload = JSON.parse(e.data);
        setData(payload);
      } catch (err) {
        console.error('Parse error', err);
      }
    };
    ws.onerror = (e) => console.error('WebSocket error', e);
    return () => ws.close();
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>🌀 Koola10 Dashboard</Text>
      <Text style={styles.value}>Revenue: ${data.revenue || 0}</Text>
      <Text style={styles.value}>Balance: ${data.balance || 0}</Text>
      <Text style={styles.value}>Health: {data.health || 'unknown'}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#0a0a0a' },
  title: { fontSize: 24, color: '#0f0', marginBottom: 20 },
  value: { fontSize: 18, color: '#fff', marginVertical: 5 },
});
