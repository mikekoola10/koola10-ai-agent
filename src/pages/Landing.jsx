import { useState, useEffect, useCallback } from 'react';
import { apiUrl, authHeaders } from '../api';
import { useToast } from '../components/Toast';

const FEATURES = [
  {
    icon: '🤖',
    title: 'Autonomous Agent Swarm',
    description: 'Deploy AI agents that work 24/7 — affiliate marketing, bounty hunting, content generation, and grant applications.',
  },
  {
    icon: '⚡',
    title: 'Real-Time Revenue Engine',
    description: 'Watch your revenue grow in real-time with our cyberpunk command center dashboard.',
  },
  {
    icon: '🔗',
    title: 'Multi-Service Integration',
    description: 'Connect Koola10, Spiral, and Apex AI engines into one unified revenue machine.',
  },
  {
    icon: '🛡️',
    title: 'Enterprise Security',
    description: 'AES-256 encryption, role-based access, and webhook verification built in.',
  },
];

const TESTIMONIALS = [
  {
    quote: "Koola10's agent swarm generated $1,905 in our first 60 seconds of running a sprint.",
    author: 'SaaS Founder',
    role: 'Y Combinator W24',
  },
  {
    quote: "The autonomous bounty hunter found $3,200 in grants we never knew existed.",
    author: 'Nonprofit Director',
    role: 'TechSoup Network',
  },
  {
    quote: "We replaced 3 freelancers with the content engine. Same quality, 10x speed.",
    author: 'Agency Owner',
    role: 'Growth Marketing',
  },
];

const PRICING_TIERS = [
  {
    name: 'STARTER',
    price: 29,
    period: '/mo',
    color: '#00f0ff',
    priceIdKey: 'STARTER',
    features: [
      '5 AI Agent Runs / Day',
      'Basic Revenue Dashboard',
      'Email Support',
      '1 API Key',
      'Standard Templates',
    ],
    cta: 'START FREE TRIAL',
    popular: false,
  },
  {
    name: 'PRO',
    price: 79,
    period: '/mo',
    color: '#8b00ff',
    priceIdKey: 'PRO',
    features: [
      '50 AI Agent Runs / Day',
      'Full Command Center',
      'Priority Support',
      '10 API Keys',
      'All Templates + Custom',
      'Webhook Integration',
    ],
    cta: 'START PRO',
    popular: true,
  },
  {
    name: 'ENTERPRISE',
    price: 199,
    period: '/mo',
    color: '#39ff14',
    priceIdKey: 'ENTERPRISE',
    features: [
      'Unlimited Agent Runs',
      'White-Label Dashboard',
      'Dedicated Support',
      'Unlimited API Keys',
      'Custom Agent Training',
      'SLA Guarantee',
    ],
    cta: 'CONTACT SALES',
    popular: false,
  },
];

const PRODUCTS = [
  {
    title: 'AI Agent Starter Kit',
    price: 29,
    color: '#00f0ff',
    priceId: 'price_1TxEyNC826DfrBa2NUekEqQA',
    description: 'Your first AI agent in 10 minutes. Includes setup guide, 5 ready-to-run agent configs, and a quickstart video tutorial.',
    features: ['Quickstart Guide', '5 Agent Configs', 'Video Tutorial', '1-Hour Setup'],
  },
  {
    title: 'Grant Proposal Templates',
    price: 39,
    color: '#39ff14',
    priceId: 'price_1TxEyiC826DfrBa2vKwkOgz5',
    description: '50 professionally written grant proposal templates for federal, state, and private foundations.',
    features: ['50 Proposal Templates', 'Federal & State Grants', 'AI Fill-in Sections', 'Copy-Paste Ready'],
  },
  {
    title: 'Koola10 Template Vault',
    price: 49,
    color: '#00f0ff',
    priceId: 'price_1TxEFJC826DfrBa2tZWKlh5C',
    description: '50+ ready-to-use templates for AI agent prompts, grant applications, content calendars, email sequences, and revenue dashboards.',
    features: ['50+ Agent Prompt Templates', 'Grant Application Templates', 'Content Calendar Kit', 'Revenue Dashboard Configs'],
  },
  {
    title: 'Content Calendar System',
    price: 59,
    color: '#8b00ff',
    priceId: 'price_1TxEyiC826DfrBa2ZAjr2tY8',
    description: '12-month AI-powered content calendar with 365 post ideas, optimal posting times, and cross-platform scheduling.',
    features: ['365 Post Ideas', 'Optimal Posting Times', 'Hashtag Research', 'Cross-Platform Templates'],
  },
  {
    title: 'Revenue Dashboard Configs',
    price: 69,
    color: '#ffd93d',
    priceId: 'price_1TxEyjC826DfrBa2YcD8yauo',
    description: '5 ready-to-deploy cyberpunk revenue dashboards for Stripe, PayPal, and manual sales tracking.',
    features: ['5 Dashboard Templates', 'Real-Time Charts', 'Goal Tracking', 'Alert System'],
  },
  {
    title: 'AI Agent Blueprint Pack',
    price: 79,
    color: '#8b00ff',
    priceId: 'price_1TxEFJC826DfrBa2IsMDj7n2',
    description: 'Step-by-step blueprints for deploying autonomous AI agents that generate revenue 24/7.',
    features: ['12 Agent Configurations', 'Revenue Engine Setup', 'ROI Calculator', 'Deployment Checklist'],
  },
  {
    title: 'Agent Training Dataset',
    price: 99,
    color: '#39ff14',
    priceId: 'price_1TxEyjC826DfrBa2MAO50VGO',
    description: 'Curated dataset of 10,000+ prompt-response pairs for training AI agents on affiliate marketing, grant writing, and content.',
    features: ['10K+ Prompt Pairs', 'Affiliate Marketing', 'Grant Writing', 'Content Generation'],
  },
  {
    title: 'Revenue Automation Course',
    price: 149,
    color: '#39ff14',
    priceId: 'price_1TxEFJC826DfrBa2JVKgLGVM',
    description: 'Complete video course: Build a $1K/day revenue engine with AI agents. 8 modules covering every vertical.',
    features: ['8 Video Modules', 'Affiliate Marketing', 'Grant Writing Mastery', 'Content Generation'],
  },
  {
    title: 'Full Stack AI Blueprint',
    price: 199,
    color: '#ff6b6b',
    priceId: 'price_1TxEyjC826DfrBa2dSl1ccYH',
    description: 'Complete system architecture for building an AI revenue engine. 24 agent configs, webhook flows, and deployment scripts.',
    features: ['24 Agent Configs', 'Webhook Flows', 'Database Schemas', 'Deployment Scripts'],
  },
  {
    title: 'Enterprise Agent Training',
    price: 499,
    color: '#ffd93d',
    priceId: null,
    description: 'Custom AI agent training for your specific business vertical.',
    features: ['Custom Agent Development', 'Industry-Specific Training', 'Performance Optimization', '30-Day Support'],
  },
];

const HERO_LINES = [
  'AUTONOMOUS',
  'AI REVENUE',
  'ENGINE',
];

export default function Landing({ onNavigate }) {
  const { toast } = useToast();
  const [email, setEmail] = useState('');
  const [subscribed, setSubscribed] = useState(false);
  const [scrollY, setScrollY] = useState(0);
  const [checkoutLoading, setCheckoutLoading] = useState(null);
  const [heroLine, setHeroLine] = useState(0);
  const [heroChar, setHeroChar] = useState(0);
  const [heroDone, setHeroDone] = useState(false);
  const [heroBlink, setHeroBlink] = useState(true);

  // ── Typing animation for hero ──────────────────────
  useEffect(() => {
    if (heroDone) return;
    const currentLine = HERO_LINES[heroLine];
    if (!currentLine) { setHeroDone(true); return; }
    if (heroChar < currentLine.length) {
      const t = setTimeout(() => setHeroChar(heroChar + 1), 80);
      return () => clearTimeout(t);
    } else if (heroLine < HERO_LINES.length - 1) {
      const t = setTimeout(() => { setHeroLine(heroLine + 1); setHeroChar(0); }, 300);
      return () => clearTimeout(t);
    } else {
      setHeroDone(true);
    }
  }, [heroLine, heroChar, heroDone]);

  useEffect(() => {
    const t = setInterval(() => setHeroBlink((b) => !b), 530);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // ── Stripe subscription checkout ─────────────────────
  const handleSubscribeCheckout = useCallback(async (tier) => {
    if (tier.name === 'ENTERPRISE') {
      toast.info('Contact sales@koola10.ai for Enterprise pricing');
      return;
    }
    setCheckoutLoading(tier.name);
    const loadingId = toast.loading(`Creating ${tier.name} checkout...`);
    try {
      const res = await fetch(apiUrl('koola10', '/admin/trigger_grants'), {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          price_id: null,
          mode: 'subscription',
          plan: tier.name.toLowerCase(),
          success_url: window.location.origin,
          cancel_url: window.location.origin,
        }),
      });
      const data = await res.json();
      if (data.checkout_url) {
        toast.updateToast(loadingId, `${tier.name} checkout ready — redirecting...`, { type: 'success', duration: 3000 });
        setTimeout(() => window.open(data.checkout_url, '_blank'), 500);
      } else if (data.error) {
        toast.updateToast(loadingId, `Checkout failed: ${data.error}`, { type: 'error', duration: 5000 });
      } else {
        toast.updateToast(loadingId, 'Checkout unavailable — backend may not support subscriptions yet. Contact support.', { type: 'warning', duration: 6000 });
      }
    } catch (err) {
      toast.updateToast(loadingId, `Network error: ${err.message}`, { type: 'error', duration: 5000 });
    } finally {
      setCheckoutLoading(null);
    }
  }, [toast]);

  // ── Product one-time checkout ────────────────────────
  const handleProductCheckout = useCallback(async (product) => {
    setCheckoutLoading(product.title);
    const loadingId = toast.loading(`Creating checkout for ${product.title}...`);
    try {
      const res = await fetch(apiUrl('koola10', '/admin/trigger_grants'), {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({ price_id: product.priceId }),
      });
      const data = await res.json();
      if (data.checkout_url) {
        toast.updateToast(loadingId, `${product.title} checkout ready — redirecting...`, { type: 'success', duration: 3000 });
        setTimeout(() => window.open(data.checkout_url, '_blank'), 500);
      } else {
        toast.updateToast(loadingId, 'Checkout unavailable. Please try again or contact support.', { type: 'error', duration: 5000 });
      }
    } catch (err) {
      toast.updateToast(loadingId, `Network error: ${err.message}`, { type: 'error', duration: 5000 });
    } finally {
      setCheckoutLoading(null);
    }
  }, [toast]);

  // ── Email waitlist ───────────────────────────────────
  const handleSubscribe = useCallback((e) => {
    e.preventDefault();
    if (email.trim()) {
      setSubscribed(true);
      setEmail('');
      toast.success('Welcome to the waitlist! Check your email for next steps.');
    }
  }, [email, toast]);

  return (
    <div className="relative min-h-screen font-mono">
      {/* ── Navigation Bar ──────────────────────────── */}
      <nav
        className="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
        style={{
          background: scrollY > 50 ? 'rgba(10, 10, 10, 0.95)' : 'transparent',
          backdropFilter: scrollY > 50 ? 'blur(12px)' : 'none',
          borderBottom: scrollY > 50 ? '1px solid rgba(0, 240, 255, 0.1)' : 'none',
        }}
      >
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🤖</span>
            <span className="text-cyan font-bold tracking-[3px] uppercase text-sm">KOOLA10</span>
          </div>
          <div className="hidden md:flex items-center gap-8">
            {['Features', 'Pricing', 'Products', 'API', 'Services'].map((item) => (
              <button
                key={item}
                onClick={() => document.getElementById(item.toLowerCase())?.scrollIntoView({ behavior: 'smooth' })}
                className="text-xs uppercase tracking-wider text-cyan/60 hover:text-cyan transition-colors"
              >
                {item}
              </button>
            ))}
            <button
              onClick={() => onNavigate('blog')}
              className="text-xs uppercase tracking-wider text-cyan/60 hover:text-cyan transition-colors"
            >
              Blog
            </button>}
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => onNavigate('login')}
              className="px-4 py-1.5 text-xs uppercase tracking-wider border border-cyan/30 text-cyan rounded hover:bg-cyan/10 transition-all"
            >
              Login
            </button>
            <button
              onClick={() => onNavigate('signup')}
              className="px-4 py-1.5 text-xs uppercase tracking-wider bg-cyan text-black rounded hover:bg-cyan/90 transition-all font-bold"
            >
              Get Started
            </button>
          </div>
        </div>
      </nav>

      {/* ── Hero Section ────────────────────────────── */}
      <section className="relative min-h-screen flex items-center justify-center px-4 pt-20">
        <div
          className="absolute inset-0 opacity-20"
          style={{
            backgroundImage:
              'linear-gradient(rgba(0,240,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0,240,255,0.1) 1px, transparent 1px)',
            backgroundSize: '50px 50px',
            animation: 'gridMove 20s linear infinite',
          }}
        />

        <div className="relative z-10 text-center max-w-5xl mx-auto">
          <div className="inline-block mb-6 px-4 py-1.5 border border-acid/30 rounded-full">
            <span className="text-acid text-xs uppercase tracking-widest">● v2.0 — Now Live</span>
          </div>

          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-6 leading-tight">
            {HERO_LINES.map((line, i) => {
              const colorClass = i === 0 ? 'text-cyan' : i === 1 ? 'text-purple' : 'text-acid';
              if (i < heroLine || heroDone) {
                return <span key={i} className={`${colorClass}`}>{line}</span>;
              }
              if (i === heroLine) {
                return (
                  <span key={i} className={`${colorClass}`}>{line.substring(0, heroChar)}<span className={`${heroBlink ? 'opacity-100' : 'opacity-0'} transition-opacity`}>_</span></span>
                );
              }
              return <span key={i} className={`${colorClass} opacity-0`}>{line}</span>;
            }).reduce((prev, curr, i) => {
              if (i === 0) return [curr];
              return [...prev, <br key={`br-${i}`} />, curr];
            }, [])}
          </h1>

          <p className="text-lg md:text-xl text-cyan/60 max-w-3xl mx-auto mb-8 leading-relaxed">
            Deploy AI agents that work 24/7 — generating revenue through affiliate marketing,
            bounty hunting, content creation, and grant applications. <span className="text-acid">No code required.</span>
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-12">
            <button
              onClick={() => onNavigate('signup')}
              className="px-8 py-3 bg-cyan text-black font-bold uppercase tracking-wider rounded-lg hover:bg-cyan/90 transition-all hover:shadow-[0_0_30px_rgba(0,240,255,0.5)]"
            >
              🚀 Start Free Trial
            </button>
            <button
              onClick={() => document.getElementById('pricing')?.scrollIntoView({ behavior: 'smooth' })}
              className="px-8 py-3 border border-purple text-purple font-bold uppercase tracking-wider rounded-lg hover:bg-purple/10 transition-all"
            >
              View Pricing →
            </button>
          </div>

          <div className="grid grid-cols-3 gap-8 max-w-2xl mx-auto">
            <div>
              <div className="text-3xl font-bold text-cyan">$1,905</div>
              <div className="text-xs text-cyan/40 uppercase tracking-wider mt-1">Avg Revenue/Sprint</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-purple">60s</div>
              <div className="text-xs text-cyan/40 uppercase tracking-wider mt-1">Time to Revenue</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-acid">$0.20</div>
              <div className="text-xs text-cyan/40 uppercase tracking-wider mt-1">Operating Costs</div>
            </div>
          </div>
        </div>

        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-bounce">
          <span className="text-cyan/30 text-xs">▼ SCROLL</span>
        </div>
      </section>

      {/* ── Features Section ────────────────────────── */}
      <section id="features" className="py-20 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-cyan uppercase tracking-wider mb-4">
              Why Koola10?
            </h2>
            <p className="text-cyan/50 max-w-2xl mx-auto">
              Built for SaaS founders, freelancers, and developers who want autonomous revenue generation.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {FEATURES.map((feature, i) => (
              <div
                key={feature.title}
                className="glass-card p-6 hover:border-cyan/40 transition-all group"
                style={{ animationDelay: `${i * 0.1}s` }}
              >
                <div className="text-4xl mb-4 group-hover:scale-110 transition-transform">{feature.icon}</div>
                <h3 className="text-sm font-bold text-cyan uppercase tracking-wider mb-2">{feature.title}</h3>
                <p className="text-xs text-cyan/50 leading-relaxed">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── 4 Revenue Paths Section ─────────────────── */}
      <section id="paths" className="py-20 px-4 bg-black/30">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-purple uppercase tracking-wider mb-4">
              4 Ways to Earn $1K/Day
            </h2>
            <p className="text-cyan/50 max-w-2xl mx-auto">
              Choose your revenue path or combine all four for maximum income.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="glass-card p-8 border-l-4 border-l-cyan hover:shadow-[0_0_30px_rgba(0,240,255,0.1)] transition-all">
              <div className="flex items-center gap-3 mb-4">
                <span className="text-3xl">💎</span>
                <div>
                  <h3 className="text-lg font-bold text-cyan uppercase tracking-wider">SaaS Subscriptions</h3>
                  <p className="text-xs text-cyan/40">$29-199/mo • Recurring Revenue</p>
                </div>
              </div>
              <p className="text-sm text-cyan/60 mb-4">
                Offer AI agent access as a subscription service. Customers get dashboard access, agent runs, and API keys.
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {['34 customers × $29/mo = $1K/day', 'Monthly recurring', 'Low churn', 'Scalable'].map((tag) => (
                  <span key={tag} className="px-2 py-0.5 text-[10px] bg-cyan/10 text-cyan/70 rounded">
                    {tag}
                  </span>
                ))}
              </div>
              <button
                onClick={() => document.getElementById('pricing')?.scrollIntoView({ behavior: 'smooth' })}
                className="text-xs text-cyan uppercase tracking-wider hover:text-cyan/80 transition-colors"
              >
                View Plans →
              </button>
            </div>

            <div className="glass-card p-8 border-l-4 border-l-purple hover:shadow-[0_0_30px_rgba(139,0,255,0.1)] transition-all">
              <div className="flex items-center gap-3 mb-4">
                <span className="text-3xl">📦</span>
                <div>
                  <h3 className="text-lg font-bold text-purple uppercase tracking-wider">Digital Products</h3>
                  <p className="text-xs text-cyan/40">$49-499 • One-Time Sales</p>
                </div>
              </div>
              <p className="text-sm text-cyan/60 mb-4">
                Sell templates, blueprints, courses, and agent configurations. High margin, instant delivery.
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {['20 sales × $49/day = $1K/day', '95% margin', 'Instant delivery', 'No support burden'].map((tag) => (
                  <span key={tag} className="px-2 py-0.5 text-[10px] bg-purple/10 text-purple/70 rounded">
                    {tag}
                  </span>
                ))}
              </div>
              <button
                onClick={() => document.getElementById('products')?.scrollIntoView({ behavior: 'smooth' })}
                className="text-xs text-purple uppercase tracking-wider hover:text-purple/80 transition-colors"
              >
                Browse Products →
              </button>
            </div>

            <div className="glass-card p-8 border-l-4 border-l-acid hover:shadow-[0_0_30px_rgba(57,255,20,0.1)] transition-all">
              <div className="flex items-center gap-3 mb-4">
                <span className="text-3xl">🔌</span>
                <div>
                  <h3 className="text-lg font-bold text-acid uppercase tracking-wider">API Platform</h3>
                  <p className="text-xs text-cyan/40">$20-200/mo • Usage-Based</p>
                </div>
              </div>
              <p className="text-sm text-cyan/60 mb-4">
                Developers pay per API call or monthly quota. Perfect for SaaS builders integrating AI agents.
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {['50 customers × $20/mo = $1K/day', 'Usage-based pricing', 'Developer-friendly', 'Network effects'].map((tag) => (
                  <span key={tag} className="px-2 py-0.5 text-[10px] bg-acid/10 text-acid/70 rounded">
                    {tag}
                  </span>
                ))}
              </div>
              <button
                onClick={() => document.getElementById('api')?.scrollIntoView({ behavior: 'smooth' })}
                className="text-xs text-acid uppercase tracking-wider hover:text-acid/80 transition-colors"
              >
                Get API Key →
              </button>
            </div>

            <div className="glass-card p-8 border-l-4 border-l-[#ffd93d] hover:shadow-[0_0_30px_rgba(255,217,61,0.1)] transition-all">
              <div className="flex items-center gap-3 mb-4">
                <span className="text-3xl">🎯</span>
                <div>
                  <h3 className="text-lg font-bold text-[#ffd93d] uppercase tracking-wider">Service Automation</h3>
                  <p className="text-xs text-cyan/40">$500-2000 • Per Client</p>
                </div>
              </div>
              <p className="text-sm text-cyan/60 mb-4">
                Run agent swarms for clients. Grant applications, content marketing, lead generation as a service.
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {['2 clients × $500/week = $1K/day', 'High-touch', 'Premium pricing', 'Retainer model'].map((tag) => (
                  <span key={tag} className="px-2 py-0.5 text-[10px] bg-[#ffd93d]/10 text-[#ffd93d]/70 rounded">
                    {tag}
                  </span>
                ))}
              </div>
              <button
                onClick={() => document.getElementById('services')?.scrollIntoView({ behavior: 'smooth' })}
                className="text-xs text-[#ffd93d] uppercase tracking-wider hover:text-[#ffd93d]/80 transition-colors"
              >
                Book Consultation →
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* ── Pricing Section ─────────────────────────── */}
      <section id="pricing" className="py-20 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-cyan uppercase tracking-wider mb-4">
              Subscription Plans
            </h2>
            <p className="text-cyan/50 max-w-2xl mx-auto">
              Start free. Scale as you grow. Cancel anytime.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {PRICING_TIERS.map((tier) => (
              <div
                key={tier.name}
                className={`glass-card p-8 relative transition-all hover:scale-105 ${
                  tier.popular ? 'border-2' : 'hover:border-opacity-40'
                }`}
                style={{
                  borderColor: tier.color,
                  boxShadow: tier.popular ? `0 0 40px ${tier.color}22` : undefined,
                }}
              >
                {tier.popular && (
                  <div
                    className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-1 text-[10px] font-bold uppercase tracking-wider rounded"
                    style={{ background: tier.color, color: '#0a0a0a' }}
                  >
                    Most Popular
                  </div>
                )}

                <div className="text-center mb-6">
                  <h3 className="text-sm font-bold uppercase tracking-wider mb-2" style={{ color: tier.color }}>
                    {tier.name}
                  </h3>
                  <div className="flex items-baseline justify-center gap-1">
                    <span className="text-4xl font-bold" style={{ color: tier.color }}>
                      ${tier.price}
                    </span>
                    <span className="text-xs text-cyan/40">{tier.period}</span>
                  </div>
                </div>

                <ul className="space-y-3 mb-8">
                  {tier.features.map((feature) => (
                    <li key={feature} className="flex items-center gap-2 text-xs text-cyan/60">
                      <span style={{ color: tier.color }}>✓</span>
                      {feature}
                    </li>
                  ))}
                </ul>

                <button
                  onClick={() => handleSubscribeCheckout(tier)}
                  disabled={checkoutLoading === tier.name}
                  className="w-full py-3 text-xs font-bold uppercase tracking-wider rounded-lg border transition-all disabled:opacity-50"
                  style={{
                    borderColor: tier.color,
                    color: tier.color,
                    background: tier.popular ? `${tier.color}11` : 'transparent',
                  }}
                  onMouseEnter={(e) => {
                    if (checkoutLoading !== tier.name) {
                      e.target.style.background = tier.color;
                      e.target.style.color = '#0a0a0a';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (checkoutLoading !== tier.name) {
                      e.target.style.background = tier.popular ? `${tier.color}11` : 'transparent';
                      e.target.style.color = tier.color;
                    }
                  }}
                >
                  {checkoutLoading === tier.name ? (
                    <span className="inline-flex items-center gap-2">
                      <span className="animate-spin inline-block w-3 h-3 border border-current border-t-transparent rounded-full" />
                      Redirecting...
                    </span>
                  ) : (
                    tier.cta
                  )}
                </button>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Products Section ────────────────────────── */}
      <section id="products" className="py-20 px-4 bg-black/30">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-purple uppercase tracking-wider mb-4">
              Digital Products
            </h2>
            <p className="text-cyan/50 max-w-2xl mx-auto">
              Instant download. Lifetime access. Start generating revenue today.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {PRODUCTS.map((product) => (
              <div
                key={product.title}
                className="glass-card p-6 hover:scale-105 transition-all group"
                style={{ borderColor: `${product.color}33` }}
              >
                <div
                  className="w-full h-2 rounded mb-4"
                  style={{ background: `linear-gradient(90deg, ${product.color}, transparent)` }}
                />
                <h3 className="text-sm font-bold uppercase tracking-wider mb-2" style={{ color: product.color }}>
                  {product.title}
                </h3>
                <div className="text-2xl font-bold mb-3" style={{ color: product.color }}>
                  ${product.price}
                </div>
                <p className="text-xs text-cyan/50 mb-4">{product.description}</p>
                <ul className="space-y-2 mb-6">
                  {product.features.map((f) => (
                    <li key={f} className="text-[10px] text-cyan/40 flex items-center gap-1">
                      <span style={{ color: product.color }}>•</span> {f}
                    </li>
                  ))}
                </ul>
                <button
                  onClick={() => handleProductCheckout(product)}
                  disabled={checkoutLoading === product.title}
                  className="w-full py-2 text-xs font-bold uppercase tracking-wider rounded border transition-all disabled:opacity-50"
                  style={{ borderColor: `${product.color}55`, color: product.color }}
                  onMouseEnter={(e) => {
                    if (checkoutLoading !== product.title) {
                      e.target.style.background = product.color;
                      e.target.style.color = '#0a0a0a';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (checkoutLoading !== product.title) {
                      e.target.style.background = 'transparent';
                      e.target.style.color = product.color;
                    }
                  }}
                >
                  {checkoutLoading === product.title ? '⏳ Redirecting...' : '💳 Buy Now'}
                </button>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── API Section ─────────────────────────────── */}
      <section id="api" className="py-20 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-2xl md:text-3xl font-bold text-acid uppercase tracking-wider mb-4">
                Developer API
              </h2>
              <p className="text-cyan/60 mb-6">
                Integrate AI agents into your own products. Usage-based pricing scales with your growth.
              </p>
              <div className="space-y-4 mb-8">
                {[
                  { label: 'REST API', desc: 'Simple HTTP endpoints' },
                  { label: 'WebSocket', desc: 'Real-time agent status' },
                  { label: 'SDKs', desc: 'Python, Node.js, Go' },
                  { label: 'Webhooks', desc: 'Event-driven integration' },
                ].map((item) => (
                  <div key={item.label} className="flex items-center gap-3">
                    <span className="w-2 h-2 rounded-full bg-acid" />
                    <span className="text-sm text-acid font-bold">{item.label}</span>
                    <span className="text-xs text-cyan/40">— {item.desc}</span>
                  </div>
                ))}
              </div>
              <button
                onClick={() => onNavigate('signup')}
                className="px-6 py-3 text-xs font-bold uppercase tracking-wider border border-acid text-acid rounded-lg hover:bg-acid hover:text-black transition-all"
              >
                Get API Key →
              </button>
            </div>

            <div className="glass-card p-6 border-acid/20">
              <div className="flex items-center gap-2 mb-4">
                <span className="w-2 h-2 rounded-full bg-red-500" />
                <span className="w-2 h-2 rounded-full bg-yellow-500" />
                <span className="w-2 h-2 rounded-full bg-green-500" />
                <span className="ml-4 text-[10px] text-cyan/40">example.py</span>
              </div>
              <pre className="text-xs text-acid/80 leading-relaxed overflow-x-auto">
                <code>{`import requests

# Initialize Koola10 API
API_KEY = "your-api-key"
BASE_URL = "https://koola10-ai-agent.onrender.com"

# Run a revenue sprint
response = requests.post(
    f"{BASE_URL}/api/v1/agents/sprint",
    headers={"Authorization": f"Bearer {API_KEY}"},
    json={"vertical": "affiliate", "agents": 10}
)

# Check revenue
revenue = requests.get(
    f"{BASE_URL}/api/v1/vault/summary",
    headers={"Authorization": f"Bearer {API_KEY}"}
)

print(f"Revenue: ${revenue.json()['total_revenue']}")
# Revenue: $1,905.54`}</code>
              </pre>
            </div>
          </div>
        </div>
      </section>

      {/* ── Services Section ────────────────────────── */}
      <section id="services" className="py-20 px-4 bg-black/30">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-[#ffd93d] uppercase tracking-wider mb-4">
              Managed Services
            </h2>
            <p className="text-cyan/50 max-w-2xl mx-auto">
              We run the agent swarms for you. Perfect for businesses that want results without the setup.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[
              {
                title: 'Grant Writing Service',
                price: '$499',
                color: '#00f0ff',
                features: ['10 Grant Applications/mo', 'AI-Powered Drafts', 'Deadline Management', 'Submission Tracking'],
              },
              {
                title: 'Content Marketing Engine',
                price: '$999',
                color: '#8b00ff',
                features: ['50 Articles/mo', 'SEO Optimization', 'Multi-Platform Distribution', 'Analytics Dashboard'],
              },
              {
                title: 'Full Revenue Automation',
                price: '$1,999',
                color: '#39ff14',
                features: ['All Agent Verticals', 'Dedicated Dashboard', 'Weekly Reports', 'Strategy Calls'],
              },
            ].map((service) => (
              <div
                key={service.title}
                className="glass-card p-8 text-center hover:scale-105 transition-all"
                style={{ borderColor: `${service.color}33` }}
              >
                <h3 className="text-sm font-bold uppercase tracking-wider mb-4" style={{ color: service.color }}>
                  {service.title}
                </h3>
                <div className="text-3xl font-bold mb-6" style={{ color: service.color }}>
                  {service.price}<span className="text-xs text-cyan/40">/mo</span>
                </div>
                <ul className="space-y-3 mb-8">
                  {service.features.map((f) => (
                    <li key={f} className="text-xs text-cyan/60 flex items-center justify-center gap-2">
                      <span style={{ color: service.color }}>✓</span> {f}
                    </li>
                  ))}
                </ul>
                <button
                  onClick={() => {
                    toast.info(`${service.title} — contact sales@koola10.ai to get started`);
                  }}
                  className="w-full py-3 text-xs font-bold uppercase tracking-wider rounded-lg border transition-all"
                  style={{ borderColor: service.color, color: service.color }}
                  onMouseEnter={(e) => {
                    e.target.style.background = service.color;
                    e.target.style.color = '#0a0a0a';
                  }}
                  onMouseLeave={(e) => {
                    e.target.style.background = 'transparent';
                    e.target.style.color = service.color;
                  }}
                >
                  Get Started
                </button>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Testimonials ────────────────────────────── */}
      <section className="py-20 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-cyan uppercase tracking-wider mb-4">
              Results, Not Promises
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {TESTIMONIALS.map((t, i) => (
              <div
                key={i}
                className="glass-card p-6"
                style={{ animationDelay: `${i * 0.1}s` }}
              >
                <p className="text-sm text-cyan/70 italic mb-4">"{t.quote}"</p>
                <div>
                  <p className="text-xs font-bold text-cyan">{t.author}</p>
                  <p className="text-[10px] text-cyan/40">{t.role}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Email Capture ───────────────────────────── */}
      <section className="py-20 px-4 bg-black/30">
        <div className="max-w-2xl mx-auto text-center">
          <h2 className="text-2xl font-bold text-cyan uppercase tracking-wider mb-4">
            Join the Waitlist
          </h2>
          <p className="text-cyan/50 mb-8">
            Get early access to new features and exclusive revenue strategies.
          </p>

          {subscribed ? (
            <div className="glass-card p-6 border-acid/30">
              <span className="text-acid text-sm font-bold uppercase tracking-wider">
                ✓ You're on the list! Check your email.
              </span>
            </div>
          ) : (
            <form onSubmit={handleSubscribe} className="flex flex-col sm:flex-row gap-3">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="your@email.com"
                required
                className="flex-1 px-4 py-3 bg-black/50 border border-cyan/20 text-cyan rounded-lg text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
              />
              <button
                type="submit"
                className="px-6 py-3 bg-cyan text-black font-bold uppercase tracking-wider rounded-lg hover:bg-cyan/90 transition-all"
              >
                Subscribe
              </button>
            </form>
          )}
        </div>
      </section>

      {/* ── Footer ──────────────────────────────────── */}
      <footer className="py-12 px-4 border-t border-cyan/10">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
            <div>
              <h4 className="text-xs font-bold text-cyan uppercase tracking-wider mb-4">Product</h4>
              <ul className="space-y-2">
                {['Features', 'Pricing', 'API Docs', 'Status'].map((item) => (
                  <li key={item}>
                    <a href={`#${item.toLowerCase()}`} className="text-xs text-cyan/40 hover:text-cyan transition-colors">
                      {item}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h4 className="text-xs font-bold text-cyan uppercase tracking-wider mb-4">Company</h4>
              <ul className="space-y-2">
                {['About', 'Blog', 'Careers', 'Contact'].map((item) => (
                  <li key={item}>
                    <a href="#" className="text-xs text-cyan/40 hover:text-cyan transition-colors">
                      {item}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h4 className="text-xs font-bold text-cyan uppercase tracking-wider mb-4">Legal</h4>
              <ul className="space-y-2">
                {['Privacy', 'Terms', 'Security', 'GDPR'].map((item) => (
                  <li key={item}>
                    <a href="#" className="text-xs text-cyan/40 hover:text-cyan transition-colors">
                      {item}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h4 className="text-xs font-bold text-cyan uppercase tracking-wider mb-4">Connect</h4>
              <ul className="space-y-2">
                {['Twitter', 'GitHub', 'Discord', 'LinkedIn'].map((item) => (
                  <li key={item}>
                    <a href="#" className="text-xs text-cyan/40 hover:text-cyan transition-colors">
                      {item}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <div className="text-center pt-8 border-t border-cyan/10">
            <p className="text-[10px] text-cyan/30 tracking-widest">
              © 2026 KOOLA10 AI AGENT ECOSYSTEM. ALL RIGHTS RESERVED.
            </p>
          </div>
        </div>
      </footer>

      <style>{`
        @keyframes gridMove {
          0% { transform: translate(0, 0); }
          100% { transform: translate(50px, 50px); }
        }
      `}</style>
    </div>
  );
}
