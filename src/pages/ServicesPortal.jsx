import { useState } from 'react';

const SERVICES = [
  {
    id: 'grants',
    title: 'Grant Writing Service',
    price: 499,
    color: '#00f0ff',
    icon: '📝',
    description: 'AI-powered grant applications with human oversight.',
    features: [
      '10 Grant Applications per month',
      'Automated grant matching from grants.gov',
      'AI-drafted applications with custom editing',
      'Deadline tracking and reminders',
      'Submission management',
      'Monthly performance reports',
    ],
    turnaround: '3-5 business days',
  },
  {
    id: 'content',
    title: 'Content Marketing Engine',
    price: 999,
    color: '#8b00ff',
    icon: '📰',
    description: 'Automated content creation and distribution across platforms.',
    features: [
      '50 SEO-optimized articles per month',
      'Multi-platform distribution (Medium, LinkedIn, Dev.to)',
      'Social media scheduling',
      'Keyword research and optimization',
      'Analytics dashboard access',
      'Monthly strategy calls',
    ],
    turnaround: '24-48 hours',
  },
  {
    id: 'revenue',
    title: 'Full Revenue Automation',
    price: 1999,
    color: '#39ff14',
    icon: '🚀',
    description: 'Complete autonomous revenue generation across all verticals.',
    features: [
      'All agent verticals activated',
      'Dedicated dashboard with real-time metrics',
      'Weekly performance reports',
      'Bi-weekly strategy calls',
      'Custom agent training for your niche',
      'Priority support with 4-hour response',
    ],
    turnaround: 'Immediate activation',
  },
];

export default function ServicesPortal({ onNavigate }) {
  const [selectedService, setSelectedService] = useState(null);
  const [contactForm, setContactForm] = useState({
    name: '',
    email: '',
    company: '',
    message: '',
  });
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e) => {
    e.preventDefault();
    // In production, this would send to backend
    setSubmitted(true);
    setTimeout(() => setSubmitted(false), 5000);
    setContactForm({ name: '', email: '', company: '', message: '' });
  };

  return (
    <div className="relative min-h-screen font-mono">
      {/* Header */}
      <header className="sticky top-0 z-20 glass-card border-b border-cyan/10 rounded-none">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => onNavigate('landing')}
              className="text-xs text-cyan/40 hover:text-cyan transition-colors uppercase tracking-wider"
            >
              ← Home
            </button>
            <div>
              <h1 className="text-lg font-bold text-[#ffd93d] uppercase tracking-[3px]">
                🎯 MANAGED SERVICES
              </h1>
              <p className="text-[10px] text-cyan/40 uppercase tracking-widest">
                We Run The Agents For You
              </p>
            </div>
          </div>
          <button
            onClick={() => onNavigate('dashboard')}
            className="px-3 py-1.5 text-[10px] border border-[#ffd93d]/30 text-[#ffd93d] rounded hover:bg-[#ffd93d]/10 transition-all uppercase tracking-wider"
          >
            Dashboard
          </button>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* Intro */}
        <div className="text-center mb-12">
          <h2 className="text-2xl md:text-3xl font-bold text-[#ffd93d] uppercase tracking-wider mb-4">
            Done-For-You Revenue Generation
          </h2>
          <p className="text-cyan/60 max-w-2xl mx-auto">
            Don't want to set up the agents yourself? We'll run the entire revenue engine for you.
            Just sit back and watch the numbers grow.
          </p>
        </div>

        {/* Service Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          {SERVICES.map((service) => (
            <div
              key={service.id}
              className={`glass-card p-8 transition-all cursor-pointer hover:scale-105 ${
                selectedService?.id === service.id ? 'ring-2' : ''
              }`}
              style={{
                borderColor: `${service.color}33`,
                ringColor: selectedService?.id === service.id ? service.color : undefined,
              }}
              onClick={() => setSelectedService(service)}
            >
              <div className="text-4xl mb-4">{service.icon}</div>
              <h3 className="text-sm font-bold uppercase tracking-wider mb-2" style={{ color: service.color }}>
                {service.title}
              </h3>
              <div className="text-3xl font-bold mb-4" style={{ color: service.color }}>
                ${service.price}<span className="text-xs text-cyan/40">/mo</span>
              </div>
              <p className="text-xs text-cyan/60 mb-4">{service.description}</p>
              <ul className="space-y-2 mb-6">
                {service.features.map((f) => (
                  <li key={f} className="text-[10px] text-cyan/50 flex items-start gap-2">
                    <span style={{ color: service.color }}>✓</span>
                    <span>{f}</span>
                  </li>
                ))}
              </ul>
              <div className="text-[10px] text-cyan/40">
                Turnaround: <span style={{ color: service.color }}>{service.turnaround}</span>
              </div>
            </div>
          ))}
        </div>

        {/* How It Works */}
        <div className="glass-card p-8 mb-12">
          <h3 className="text-lg font-bold text-cyan uppercase tracking-wider mb-6 text-center">
            How It Works
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {[
              { step: '01', title: 'Choose Service', desc: 'Select the service that fits your needs' },
              { step: '02', title: 'Onboard', desc: 'Share your goals and we configure the agents' },
              { step: '03', title: 'Activation', desc: 'Agents start working within 24 hours' },
              { step: '04', title: 'Results', desc: 'Watch revenue grow in your dashboard' },
            ].map((item, i) => (
              <div key={item.step} className="text-center">
                <div className="text-3xl font-bold text-cyan/20 mb-2">{item.step}</div>
                <h4 className="text-sm font-bold text-cyan mb-1">{item.title}</h4>
                <p className="text-[10px] text-cyan/50">{item.desc}</p>
                {i < 3 && (
                  <div className="hidden md:block text-cyan/20 text-xl mt-4">→</div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Contact Form */}
        <div className="glass-card p-8 max-w-2xl mx-auto">
          <h3 className="text-lg font-bold text-[#ffd93d] uppercase tracking-wider mb-6 text-center">
            Get Started
          </h3>

          {submitted ? (
            <div className="text-center py-8">
              <div className="text-4xl mb-4">✓</div>
              <p className="text-acid font-bold uppercase tracking-wider">
                Request Submitted!
              </p>
              <p className="text-xs text-cyan/50 mt-2">
                We'll contact you within 24 hours to discuss your needs.
              </p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                    Name
                  </label>
                  <input
                    type="text"
                    value={contactForm.name}
                    onChange={(e) => setContactForm({ ...contactForm, name: e.target.value })}
                    required
                    className="w-full px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                    Email
                  </label>
                  <input
                    type="email"
                    value={contactForm.email}
                    onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
                    required
                    className="w-full px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                  />
                </div>
              </div>
              <div>
                <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                  Company
                </label>
                <input
                  type="text"
                  value={contactForm.company}
                  onChange={(e) => setContactForm({ ...contactForm, company: e.target.value })}
                  className="w-full px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                />
              </div>
              <div>
                <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                  Service Interest
                </label>
                <select
                  value={selectedService?.id || ''}
                  onChange={(e) => setSelectedService(SERVICES.find((s) => s.id === e.target.value))}
                  className="w-full px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors"
                >
                  <option value="">Select a service...</option>
                  {SERVICES.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.title} - ${s.price}/mo
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-[10px] text-cyan/50 uppercase tracking-wider mb-1">
                  Message
                </label>
                <textarea
                  value={contactForm.message}
                  onChange={(e) => setContactForm({ ...contactForm, message: e.target.value })}
                  rows={4}
                  placeholder="Tell us about your goals..."
                  className="w-full px-4 py-2 bg-black/50 border border-cyan/20 text-cyan rounded text-sm font-mono focus:outline-none focus:border-cyan transition-colors resize-none"
                />
              </div>
              <button
                type="submit"
                className="w-full py-3 bg-[#ffd93d] text-black font-bold uppercase tracking-wider rounded hover:bg-[#ffd93d]/90 transition-all"
              >
                Submit Request
              </button>
            </form>
          )}
        </div>
      </main>
    </div>
  );
}
