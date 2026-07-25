import { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';

const ToastContext = createContext(null);

let toastId = 0;

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const timersRef = useRef({});

  const removeToast = useCallback((id) => {
    clearTimeout(timersRef.current[id]);
    delete timersRef.current[id];
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback(
    (message, { type = 'info', duration = 4000, icon = null } = {}) => {
      const id = ++toastId;
      const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️',
        loading: '⏳',
      };
      setToasts((prev) => [
        ...prev,
        { id, message, type, icon: icon || icons[type] || icons.info, removing: false },
      ]);

      if (duration > 0) {
        timersRef.current[id] = setTimeout(() => {
          setToasts((prev) =>
            prev.map((t) => (t.id === id ? { ...t, removing: true } : t))
          );
          setTimeout(() => removeToast(id), 300);
        }, duration);
      }

      return id;
    },
    [removeToast]
  );

  const toast = useCallback(
    (message, opts) => addToast(message, opts),
    [addToast]
  );
  toast.success = (message, opts) => addToast(message, { ...opts, type: 'success' });
  toast.error = (message, opts) => addToast(message, { ...opts, type: 'error' });
  toast.warning = (message, opts) => addToast(message, { ...opts, type: 'warning' });
  toast.info = (message, opts) => addToast(message, { ...opts, type: 'info' });
  toast.loading = (message, opts) => addToast(message, { ...opts, type: 'loading', duration: 0 });

  // Allow updating an existing toast (useful for loading → success/error)
  const updateToast = useCallback(
    (id, message, { type = 'info', icon = null, duration = 4000 } = {}) => {
      const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️',
      };
      clearTimeout(timersRef.current[id]);
      delete timersRef.current[id];
      setToasts((prev) =>
        prev.map((t) =>
          t.id === id ? { ...t, message, type, icon: icon || icons[type] || icons.info, removing: false } : t
        )
      );
      if (duration > 0) {
        timersRef.current[id] = setTimeout(() => {
          setToasts((prev) =>
            prev.map((t) => (t.id === id ? { ...t, removing: true } : t))
          );
          setTimeout(() => removeToast(id), 300);
        }, duration);
      }
    },
    [removeToast]
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      Object.values(timersRef.current).forEach(clearTimeout);
    };
  }, []);

  return (
    <ToastContext.Provider value={{ toast, updateToast, removeToast }}>
      {children}
      {/* Toast container */}
      <div className="fixed bottom-6 right-6 z-[9999] flex flex-col gap-3 pointer-events-none max-w-sm">
        {toasts.map((t) => {
          const borderColor =
            t.type === 'success'
              ? '#39ff14'
              : t.type === 'error'
              ? '#ff3366'
              : t.type === 'warning'
              ? '#ffd93d'
              : t.type === 'loading'
              ? '#8b00ff'
              : '#00f0ff';

          return (
            <div
              key={t.id}
              className={`pointer-events-auto glass-card px-4 py-3 flex items-start gap-3 transition-all duration-300 ${
                t.removing ? 'opacity-0 translate-x-8 scale-95' : 'opacity-100 translate-x-0 scale-100'
              }`}
              style={{
                borderColor: `${borderColor}44`,
                boxShadow: `0 0 20px ${borderColor}22, inset 0 0 20px ${borderColor}08`,
                animation: t.type === 'loading' ? 'pulse-glow 2s ease-in-out infinite' : undefined,
              }}
            >
              <span className="text-base mt-0.5 shrink-0">{t.icon}</span>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-mono leading-relaxed" style={{ color: borderColor }}>
                  {t.message}
                </p>
              </div>
              <button
                onClick={() => {
                  clearTimeout(timersRef.current[t.id]);
                  delete timersRef.current[t.id];
                  setToasts((prev) =>
                    prev.map((tt) => (tt.id === t.id ? { ...tt, removing: true } : tt))
                  );
                  setTimeout(() => removeToast(t.id), 300);
                }}
                className="text-cyan/30 hover:text-cyan transition-colors text-xs shrink-0"
              >
                ✕
              </button>
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}
