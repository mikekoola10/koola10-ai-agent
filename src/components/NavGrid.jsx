const CARDS = [
  {
    emoji: '🏦',
    title: 'THE_VAULT',
    description: 'Personal Finance Command Center. Track transactions, manage income, and secure your financial future via Supabase.',
    href: 'vault.html',
  },
  {
    emoji: '📊',
    title: 'ADMIN_PANEL',
    description: 'System monitoring, node management, and real-time telemetry for the Koola10 swarm network.',
    href: 'dashboard.html',
  },
  {
    emoji: '🎨',
    title: 'NOVA_STUDIO',
    description: 'Creative production engine for AI-driven media and content generation.',
    href: 'nova-studio.html',
  },
  {
    emoji: '👁️',
    title: 'PORTAL_WATCH',
    description: 'Real-time monitoring and event tracking for the Diner portal ecosystem.',
    href: 'diner-portal-watch.html',
  },
];

export default function NavGrid() {
  return (
    <div className="glass-card p-6">
      <h2 className="text-lg font-bold font-mono text-cyan uppercase tracking-wider mb-4">
        🧭 NAVIGATION
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {CARDS.map((card, i) => (
          <a
            key={i}
            href={card.href}
            className="card-hover block p-5 border rounded-lg no-underline"
            style={{
              borderColor: 'rgba(0, 240, 255, 0.15)',
              background: 'rgba(13, 13, 13, 0.6)',
              animation: `fadeInUp 0.5s ease-out ${i * 0.1}s both`,
            }}
          >
            <h3 className="text-base font-bold font-mono uppercase mb-1.5 tracking-wider text-cyan">
              {card.emoji} {card.title}
            </h3>
            <p className="text-xs opacity-60 font-mono" style={{ color: '#b0b0b0' }}>
              {card.description}
            </p>
          </a>
        ))}
      </div>
    </div>
  );
}
