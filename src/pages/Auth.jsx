import { useState } from 'react';
import { useToast } from '../components/Toast';

export default function Auth({ mode = 'login', onNavigate, onAuth }) {
  const { toast } = useToast();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const isLogin = mode === 'login';

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (!email || !password) {
        throw new Error('Email and password are required');
      }

      if (!isLogin && !name) {
        throw new Error('Name is required for signup');
      }

      // Simulate auth delay
      await new Promise((resolve) => setTimeout(resolve, 800));

      const user = {
        id: 'user_' + Date.now(),
        email,
        name: name || email.split('@')[0],
        plan: 'starter',
        apiKeys: 1,
        createdAt: new Date().toISOString(),
      };

      localStorage.setItem('koola10_user', JSON.stringify(user));
      localStorage.setItem('koola10_session', 'active');

      if (isLogin) {
        toast.success(`Welcome back, ${user.name}!`);
        onAuth(user);
      } else {
        setSuccess('Account created! Redirecting...');
        toast.success('Account created successfully! Welcome aboard.');
        setTimeout(() => onAuth(user), 1000);
      }
    } catch (err) {
      setError(err.message);
      toast.error(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative min-h-screen flex items-center justify-center px-4 font-mono">
      {/* Background grid */}
      <div
        className="absolute inset-0 opacity-10"
        style={{
          backgroundImage:
            'linear-gradient(rgba(0,240,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0,240,255,0.1) 1px, transparent 1px)',
          backgroundSize: '30px 30px',
        }}
      />

      <div className="relative z-10 w-full max-w-md">
        {/* Back to home */}
        <button
          onClick={() => onNavigate('landing')}
          className="mb-6 text-xs text-cyan/40 hover:text-cyan transition-colors uppercase tracking-wider"
        >
          ← Back to Home
        </button>

        {/* Auth Card */}
        <div className="glass-card p-8">
          {/* Header */}
          <div className="text-center mb-8">
            <span className="text-4xl mb-4 block">🤖</span>
            <h1 className="text-xl font-bold text-cyan uppercase tracking-[3px]">
              {isLogin ? 'Welcome Back' : 'Join Koola10'}
            </h1>
            <p className="text-xs text-cyan/40 mt-2">
              {isLogin ? 'Sign in to your command center' : 'Start your autonomous revenue engine'}
            </p>
          </div>

          {/* Error/Success */}
          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded text-xs text-red-400 animate-slide-down">
              {error}
            </div>
          )}
          {success && (
            <div className="mb-4 p-3 bg-acid/10 border border-acid/30 rounded text-xs text-acid animate-slide-down">
              {success}
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            {!isLogin && (
              <div>
                <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                  Full Name
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="John Doe"
                  className="w-full px-4 py-3 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                />
              </div>
            )}

            <div>
              <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
                className="w-full px-4 py-3 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
              />
            </div>

            <div>
              <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                minLength={6}
                className="w-full px-4 py-3 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-cyan text-black font-bold uppercase tracking-wider rounded hover:bg-cyan/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="inline-flex items-center gap-2">
                  <span className="animate-spin inline-block w-3 h-3 border border-current border-t-transparent rounded-full" />
                  {isLogin ? 'Signing In...' : 'Creating Account...'}
                </span>
              ) : isLogin ? (
                'Sign In'
              ) : (
                'Create Account'
              )}
            </button>
          </form>

          {/* Divider */}
          <div className="flex items-center gap-4 my-6">
            <div className="flex-1 h-px bg-cyan/10" />
            <span className="text-[10px] text-cyan/30 uppercase">or</span>
            <div className="flex-1 h-px bg-cyan/10" />
          </div>

          {/* Social logins (placeholder) */}
          <div className="space-y-3">
            <button
              onClick={() => toast.info('Google OAuth coming soon — use email signup in the meantime')}
              className="w-full py-2.5 border border-cyan/20 text-cyan/60 rounded text-xs uppercase tracking-wider hover:bg-cyan/5 transition-all flex items-center justify-center gap-2"
            >
              <span>🔗</span> Continue with Google
            </button>
            <button
              onClick={() => toast.info('GitHub OAuth coming soon — use email signup in the meantime')}
              className="w-full py-2.5 border border-cyan/20 text-cyan/60 rounded text-xs uppercase tracking-wider hover:bg-cyan/5 transition-all flex items-center justify-center gap-2"
            >
              <span>⚡</span> Continue with GitHub
            </button>
          </div>

          {/* Toggle mode */}
          <p className="text-center mt-6 text-xs text-cyan/40">
            {isLogin ? "Don't have an account?" : 'Already have an account?'}{' '}
            <button
              onClick={() => onNavigate(isLogin ? 'signup' : 'login')}
              className="text-cyan hover:text-cyan/80 transition-colors"
            >
              {isLogin ? 'Sign Up' : 'Sign In'}
            </button>
          </p>
        </div>

        {/* Footer note */}
        <p className="text-center mt-6 text-[10px] text-cyan/20 tracking-wider">
          BY SIGNING UP, YOU AGREE TO OUR TERMS OF SERVICE AND PRIVACY POLICY
        </p>
      </div>
    </div>
  );
}
