'use client';

import React from 'react';
import { Bell, X, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

interface NotificationPanelProps {
  notification: {
    alert_id: string;
    title: string;
    body: string;
  } | null;
  onProceed: (id: string) => void;
  onDismiss: (id: string) => void;
}

export function NotificationPanel({ notification, onProceed, onDismiss }: NotificationPanelProps) {
  React.useEffect(() => {
    if (notification && typeof window !== 'undefined' && 'Notification' in window) {
      if (Notification.permission === 'granted') {
        new Notification(notification.title, { body: notification.body });
      } else if (Notification.permission !== 'denied') {
        Notification.requestPermission();
      }
    }
  }, [notification]);

  if (!notification) return null;

  return (
    <div className="fixed top-4 right-4 z-50 w-full max-w-sm animate-in slide-in-from-right">
      <Card className="bg-indigo-950/90 border-amber-400/50 backdrop-blur-xl p-4 shadow-2xl">
        <div className="flex items-start gap-4">
          <div className="p-2 bg-amber-400 rounded-lg">
            <Bell className="h-5 w-5 text-black" />
          </div>
          <div className="flex-1 space-y-1">
            <h3 className="text-white font-bold">{notification.title}</h3>
            <p className="text-white/70 text-sm">{notification.body}</p>
            <div className="flex gap-2 pt-3">
              <Button
                variant="default"
                size="sm"
                className="bg-amber-400 text-black hover:bg-amber-500"
                onClick={() => onProceed(notification.alert_id)}
              >
                <Check className="h-4 w-4 mr-2" />
                Proceed
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="border-white/20 text-white hover:bg-white/10"
                onClick={() => onDismiss(notification.alert_id)}
              >
                <X className="h-4 w-4 mr-2" />
                Dismiss
              </Button>
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}
