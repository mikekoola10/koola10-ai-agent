import React from 'react';
import { StyleSheet, Text, View, TouchableOpacity } from 'react-native';

export default function SettingsScreen() {
  return (
    <View style={styles.container}>
      <TouchableOpacity style={styles.button}>
        <Text style={styles.buttonText}>TRIGGER_FIRE_DRILL</Text>
      </TouchableOpacity>

      <TouchableOpacity style={styles.button}>
        <Text style={styles.buttonText}>RESTART_ORCHESTRATOR</Text>
      </TouchableOpacity>

      <TouchableOpacity style={[styles.button, { borderColor: '#f00' }]}>
        <Text style={[styles.buttonText, { color: '#f00' }]}>EMERGENCY_STOP</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#000', padding: 20 },
  button: { padding: 20, borderWidth: 1, borderColor: '#333', borderRadius: 8, marginBottom: 15 },
  buttonText: { color: '#fff', textAlign: 'center', fontWeight: 'bold', letterSpacing: 1 }
});
