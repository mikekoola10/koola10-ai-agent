import React, { useContext } from 'react';
import { StyleSheet, Text, View, FlatList } from 'react-native';
import { TelemetryContext } from '../context/TelemetryContext';

export default function HealthScreen() {
  const services = [
    { name: 'Orchestrator', status: 'Healthy', latency: '12ms' },
    { name: 'Browser Agent', status: 'Healthy', latency: '45ms' },
    { name: 'Semantic Agent', status: 'Healthy', latency: '8ms' },
    { name: 'Spiral Agent', status: 'Degraded', latency: '502' },
  ];

  return (
    <View style={styles.container}>
      <FlatList
        data={services}
        renderItem={({ item }) => (
          <View style={styles.item}>
            <Text style={styles.name}>{item.name}</Text>
            <Text style={[styles.status, { color: item.status === 'Healthy' ? '#0f0' : '#f00' }]}>
              {item.status} ({item.latency})
            </Text>
          </View>
        )}
        keyExtractor={item => item.name}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#000', padding: 20 },
  item: { flexDirection: 'row', justifyContent: 'space-between', padding: 15, borderBottomWidth: 1, borderBottomColor: '#222' },
  name: { color: '#fff', fontSize: 16 },
  status: { fontWeight: 'bold' }
});
