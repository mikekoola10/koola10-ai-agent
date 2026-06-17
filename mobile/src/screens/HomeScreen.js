import React, { useContext } from 'react';
import { StyleSheet, Text, View, Dimensions } from 'react-native';
import { TelemetryContext } from '../context/TelemetryContext';

export default function HomeScreen() {
  const { data } = useContext(TelemetryContext);

  return (
    <View style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.label}>TOTAL_REVENUE</Text>
        <Text style={styles.value}>${data.revenue.toFixed(2)}</Text>
      </View>

      <View style={styles.card}>
        <Text style={styles.label}>CURRENT_BALANCE</Text>
        <Text style={styles.value}>${data.balance.toFixed(2)}</Text>
      </View>

      <View style={styles.card}>
        <Text style={styles.label}>SYSTEM_STATUS</Text>
        <Text style={[styles.value, { color: data.status === 'online' ? '#0f0' : '#f00' }]}>
          {data.status.toUpperCase()}
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#000', padding: 20 },
  card: { backgroundColor: '#0a0a0a', padding: 25, borderRadius: 12, marginBottom: 20, borderWidth: 1, borderColor: '#1f1f1f' },
  label: { color: '#666', fontSize: 12, fontWeight: 'bold', marginBottom: 10, letterSpacing: 1 },
  value: { color: '#fff', fontSize: 32, fontWeight: 'bold' }
});
