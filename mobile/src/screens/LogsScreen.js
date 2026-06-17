import React, { useContext } from 'react';
import { StyleSheet, View, Text, FlatList } from 'react-native';
import { TelemetryContext } from '../context/TelemetryContext';

export default function LogsScreen() {
  const { logs } = useContext(TelemetryContext);

  return (
    <View style={styles.container}>
      <FlatList
        data={logs}
        renderItem={({ item }) => (
          <View style={styles.logItem}>
            <Text style={styles.timestamp}>[{item.timestamp.split('T')[1].split('.')[0]}]</Text>
            <Text style={styles.action}>{item.action}</Text>
          </View>
        )}
        keyExtractor={(item, index) => i.toString() + index}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#000' },
  logItem: { padding: 12, borderBottomWidth: 1, borderBottomColor: '#111', flexDirection: 'row' },
  timestamp: { color: '#444', fontFamily: 'monospace', fontSize: 12, marginRight: 10 },
  action: { color: '#0f0', fontFamily: 'monospace', fontSize: 12 }
});
