import { useState, useEffect, useRef } from 'react';
import MatrixRain from './components/MatrixRain';
import LedgerDashboard from './components/LedgerDashboard';
import RevenueEngine from './components/RevenueEngine';
import SystemHealth from './components/SystemHealth';
import MemoryPanel from './components/MemoryPanel';
import SubscriptionsPanel from './components/SubscriptionsPanel';
import NavGrid from './components/NavGrid';
import Landing from './pages/Landing';
import Auth from './pages/Auth';
import DeveloperPortal from './pages/DeveloperPortal';
import ServicesPortal from './pages/ServicesPortal';

// Simple client-side router
function useRouter() {
  const [page, setPage] = useState(() => {
    // Check localStorage for last page, default to landing
    return localStorage.getItem('koola10_page') || 'landing';
  });
  const [params, setParams] = useState({});

  const navigate = (newPage, newParams = {}) => {
    setPage(newPage);
    setParams(newParams);
    localStorage.setItem('koola10_page', newPage);
  };

  return { page, params, navigate };
}

export default function App() {
  const { page, params, navigate } = useRouter();
  const [user, setUser] = useState(null);
  const [time, setTime] = useState(new Date().toLocaleTimeString('en-US', { hour12: false }));
  const ledgerRefreshRef = useRef(null);

  // ── Check auth on mount ──────────────────────────────
  useEffect(() => {
    const stored = localStorage.getItem('koola10_user');
    const session = localStorage.getItem('koola10_session');
    if (stored && session === 'active') {
      setUser(JSON.parse(stored));
    }
  }, []);

  // ── Clock ticker ────────────────────────────────────────
  useEffect(() => {
    const id = setInterval(() => {
      setTime(new Date().toLocaleTimeString('en-US', { hour12: false }));
    }, 1000);
    return () => clearInterval(id);
  }, []);

  // ── Auth handler ─────────────────────────────────────────
  const handleAuth = (userData) => {
    setUser(userData);
    navigate('dashboard');
  };

  const handleLogout = () => {
    localStorage.removeItem('koola10_user');
    localStorage.removeItem('koola10_session');
    setUser(null);
    navigate('landing');
  };

  // ── Refresh ledger after revenue engine actions ─────────
  const handleRevenueAction = () => {
    setTimeout(() => {
      if (ledgerRefreshRef.current) ledgerRefreshRef.current();
    }, 2000);
  };

  // ── Route rendering ──────────────────────────────────────
  const renderPage = () => {
    switch (page) {
      case 'landing':
        return <Landing onNavigate={navigate} />;

      case 'login':
        return <Auth mode="login" onNavigate={navigate} onAuth={handleAuth} />;

      case 'signup':
        return <Auth mode="signup" onNavigate={navigate} onAuth={handleAuth} />;

      case 'developers':
        return <DeveloperPortal onNavigate={navigate} />;

      case 'services':
        return <ServicesPortal onNavigate={navigate} />;

      case 'dashboard':
        if (!user) {
          return <Auth mode="login" onNavigate={navigate} onAuth={handleAuth} />;
        }
        return renderDashboard();

      default:
        return <Landing onNavigate={navigate} />;
    }
  };

  // ── Admin Dashboard ──────────────────────────────────────
  const renderDashboard = () => (
    <>
      <MatrixRain />
      <div className="relative z-10 min-h-screen font-mono">
        {/* ── Fixed Header ────────────────────────────── */}
        <header className="sticky top-0 z-20 glass-card border-b border-cyan/10 rounded-none mx-0">
          <div className="max-w-6xl mx-auto px-4 py-3 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h1 className="text-xl md:text-2xl font-bold uppercase tracking-[4px] glitch-text text-cyan">
                KOOLA10_SYSTEM
              </h1>
              <p className="text-[10px] md:text-xs uppercase tracking-widest text-cyan/50">
                Autonomous AI Agent Ecosystem — v2.0.0
              </p>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-xs text-cyan/40 tracking-wider">
                [ SYS_TIME: {time} ] [ USER: {user?.email || 'GUEST'} ]
              </span>
              <span className="hidden sm:inline text-xs text-acid tracking-wider">
                ● LIVE
              </span>
              <button
                onClick={handleLogout}
                className="px-3 py-1.5 text-[10px] border border-red-500/30 text-red-400 rounded hover:bg-red-500/10 transition-all uppercase tracking-wider"
              >
                Logout
              </button>
            </div>
          </div>
        </header>

        {/* ── Dashboard Body ──────────────────────────── */}
        <main className="max-w-6xl mx-auto px-4 py-6 space-y-5">
          {/* Quick nav to other portals */}
          <div className="flex gap-3 flex-wrap">
            <button
              onClick={() => navigate('landing')}
              className="px-3 py-1.5 text-[10px] border border-cyan/20 text-cyan/60 rounded hover:bg-cyan/5 transition-all uppercase tracking-wider"
            >
              ← Public Site
            </button>
            <button
              onClick={() => navigate('developers')}
              className="px-3 py-1.5 text-[10px] border border-acid/20 text-acid/60 rounded hover:bg-acid/5 transition-all uppercase tracking-wider"
            >
              🔌 API Portal
            </button>
            <button
              onClick={() => navigate('services')}
              className="px-3 py-1.5 text-[10px] border border-[#ffd93d]/20 text-[#ffd93d]/60 rounded hover:bg-[#ffd93d]/5 transition-all uppercase tracking-wider"
            >
              🎯 Services
            </button>
          </div>

          {/* Section: Health + Ledger */}
          <div className="animate-fade-in">
            <SystemHealth />
          </div>

          <div className="animate-fade-in" style={{ animationDelay: '0.1s' }}>
            <LedgerDashboard onRefresh={ledgerRefreshRef} />
          </div>

          {/* Section: Revenue Engine */}
          <div className="animate-fade-in" style={{ animationDelay: '0.2s' }}>
            <RevenueEngine onAction={handleRevenueAction} />
          </div>

          {/* Section: Memory + Subscriptions side by side on large screens */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <div className="animate-fade-in" style={{ animationDelay: '0.3s' }}>
              <MemoryPanel />
            </div>
            <div className="animate-fade-in" style={{ animationDelay: '0.35s' }}>
              <SubscriptionsPanel />
            </div>
          </div>

          {/* Section: Navigation */}
          <div className="animate-fade-in" style={{ animationDelay: '0.45s' }}>
            <NavGrid />
          </div>

          {/* ── Footer ─────────────────────────────────── */}
          <footer className="text-center py-8 text-[10px] md:text-xs text-cyan/30 tracking-widest">
            [ SWARM: APEX | SPIRAL | KOOLA10 ]&nbsp;&nbsp;
            [ ENCRYPTION: AES-256 ]&nbsp;&nbsp;
            [ ACCESS_LEVEL: ROOT ]&nbsp;&nbsp;
            [ STATUS: OPERATIONAL ]
          </footer>
        </main>
      </div>
    </>
  );

  return renderPage();
}
