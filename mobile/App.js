import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { LayoutDashboard, Activity, ScrollText, Settings } from 'lucide-react-native';
import { TelemetryProvider } from './src/context/TelemetryContext';

import HomeScreen from './src/screens/HomeScreen';
import HealthScreen from './src/screens/HealthScreen';
import LogsScreen from './src/screens/LogsScreen';
import SettingsScreen from './src/screens/SettingsScreen';

const Tab = createBottomTabNavigator();

export default function App() {
  return (
    <TelemetryProvider>
      <NavigationContainer>
        <Tab.Navigator
          screenOptions={{
            headerStyle: { backgroundColor: '#0a0a0a', borderBottomWidth: 1, borderBottomColor: '#333' },
            headerTitleStyle: { fontWeight: 'bold', letterSpacing: 2, fontSize: 16 },
            headerTintColor: '#0f0',
            tabBarStyle: { backgroundColor: '#0a0a0a', borderTopColor: '#333', height: 60, paddingBottom: 10 },
            tabBarActiveTintColor: '#0f0',
            tabBarInactiveTintColor: '#666',
          }}
        >
          <Tab.Screen
            name="Home"
            component={HomeScreen}
            options={{ tabBarIcon: ({ color }) => <LayoutDashboard color={color} size={22} /> }}
          />
          <Tab.Screen
            name="Health"
            component={HealthScreen}
            options={{ tabBarIcon: ({ color }) => <Activity color={color} size={22} /> }}
          />
          <Tab.Screen
            name="Logs"
            component={LogsScreen}
            options={{ tabBarIcon: ({ color }) => <ScrollText color={color} size={22} /> }}
          />
          <Tab.Screen
            name="Settings"
            component={SettingsScreen}
            options={{ tabBarIcon: ({ color }) => <Settings color={color} size={22} /> }}
          />
        </Tab.Navigator>
      </NavigationContainer>
    </TelemetryProvider>
  );
}
