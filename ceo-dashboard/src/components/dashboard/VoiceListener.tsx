'use client';

import React, { useEffect, useState } from 'react';
import { Mic } from 'lucide-react';

interface VoiceListenerProps {
  isActive: boolean;
  onTranscription: (text: string) => void;
  onFinal: (text: string) => void;
}

export function VoiceListener({ isActive, onTranscription, onFinal }: VoiceListenerProps) {
  const [transcription, setTranscription] = useState('');


  useEffect(() => {
    if (!isActive) {
      return;
    }

    const timer = setTimeout(() => {
      onFinal('timeout');
    }, 10000);

    const SpeechRecognition = (window as { SpeechRecognition?: any, webkitSpeechRecognition?: any }).SpeechRecognition || (window as { SpeechRecognition?: any, webkitSpeechRecognition?: any }).webkitSpeechRecognition;
    if (!SpeechRecognition) {
      console.error('Speech recognition not supported');
      return;
    }

    const recognition = new SpeechRecognition();
    recognition.continuous = true;
    recognition.interimResults = true;
    recognition.lang = 'en-US';

    recognition.onresult = (event: { resultIndex: number, results: any[] }) => {
      let interimTranscript = '';
      let finalTranscript = '';

      for (let i = event.resultIndex; i < event.results.length; ++i) {
        if (event.results[i].isFinal) {
          finalTranscript += event.results[i][0].transcript;
        } else {
          interimTranscript += event.results[i][0].transcript;
        }
      }

      const current = finalTranscript || interimTranscript;
      setTranscription(current);
      onTranscription(current);

      if (finalTranscript) {
        onFinal(finalTranscript);
      }
    };

    recognition.start();

    return () => {
      recognition.stop();
      setTranscription('');
      clearTimeout(timer);
    };
  }, [isActive, onTranscription, onFinal]);

  if (!isActive) return null;

  return (
    <div className="fixed bottom-24 left-1/2 -translate-x-1/2 z-50 flex flex-col items-center gap-4">
      <div className="relative">
        <div className="absolute inset-0 bg-amber-400 rounded-full animate-ping opacity-20" />
        <div className="relative p-4 bg-amber-400 rounded-full shadow-lg">
          <Mic className="h-8 w-8 text-black" />
        </div>
      </div>
      <div className="bg-black/80 backdrop-blur-md px-6 py-3 rounded-full border border-amber-400/30">
        <p className="text-amber-400 font-medium whitespace-nowrap">
          {transcription || 'Listening...'}
        </p>
      </div>
    </div>
  );
}
