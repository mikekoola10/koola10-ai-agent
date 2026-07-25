import { useState, useEffect, useCallback } from 'react';
import { apiUrl, authHeaders } from '../api';
import { useToast } from '../components/Toast';

const SAMPLE_POSTS = [
  {
    slug: 'how-ai-agents-generate-revenue',
    title: 'How AI Agents Generate Revenue 24/7 Without Human Intervention',
    excerpt: 'Discover how autonomous AI agents work around the clock to find grants, create content, and drive affiliate sales — generating real revenue while you sleep.',
    category: 'Revenue Engine',
    date: 'July 25, 2026',
    readTime: '5 min',
    color: '#00f0ff',
  },
  {
    slug: 'grant-writing-automation',
    title: 'Grant Writing Automation: How AI Drafts Winning Proposals in Minutes',
    excerpt: 'Learn how AI-powered grant writing tools can analyze opportunities, draft proposals, and submit applications — replacing weeks of manual work.',
    category: 'Grant Writing',
    date: 'July 24, 2026',
    readTime: '4 min',
    color: '#39ff14',
  },
  {
    slug: 'building-revenue-engine',
    title: 'Building a $1K/Day Revenue Engine: The Complete Blueprint',
    excerpt: 'A step-by-step guide to building an autonomous revenue system with 4 income streams: SaaS, digital products, API access, and managed services.',
    category: 'Strategy',
    date: 'July 23, 2026',
    readTime: '7 min',
    color: '#8b00ff',
  },
  {
    slug: 'ai-content-marketing',
    title: 'AI Content Marketing: Generate 50 Articles Per Month Automatically',
    excerpt: 'How AI agents create, optimize, and distribute content across platforms — driving organic traffic without hiring a content team.',
    category: 'Content',
    date: 'July 22, 2026',
    readTime: '5 min',
    color: '#ffd93d',
  },
  {
    slug: 'affiliate-marketing-ai',
    title: 'Affiliate Marketing with AI: Earn Passive Income on Autopilot',
    excerpt: 'AI agents that find high-converting affiliate programs, create promotional content, and optimize campaigns for maximum commissions.',
    category: 'Affiliate',
    date: 'July 21, 2026',
    readTime: '4 min',
    color: '#ff6b6b',
  },
  {
    slug: 'saas-pricing-strategy',
    title: 'SaaS Pricing Strategy: How to Price AI Products for Maximum Revenue',
    excerpt: 'Data-driven pricing strategies for AI-powered SaaS products. Learn the psychology behind $29/$79/$199 tier structures.',
    category: 'Pricing',
    date: 'July 20, 2026',
    readTime: '6 min',
    color: '#00f0ff',
  },
];

export default function Blog() {
  const { toast } = useToast();
  const [posts, setPosts] = useState(SAMPLE_POSTS);
  const [selectedPost, setSelectedPost] = useState(null);
  const [generating, setGenerating] = useState(false);
  const [email, setEmail] = useState('');
  const [subscribed, setSubscribed] = useState(false);

  const handleGeneratePost = useCallback(async () => {
    setGenerating(true);
    const loadingId = toast.loading('Generating blog post with AI...');
    try {
      const res = await fetch(apiUrl('koola10', '/blog/generate'), {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          topic: 'AI revenue generation strategies',
          style: 'educational',
        }),
      });
      const data = await res.json();
      if (data.title) {
        setPosts((prev) => [data, ...prev]);
        toast.updateToast(loadingId, 'Blog post generated!', { type: 'success', duration: 3000 });
      } else {
        toast.updateToast(loadingId, 'Using sample posts — blog generation coming soon', { type: 'info', duration: 3000 });
      }
    } catch {
      toast.updateToast(loadingId, 'Blog generation endpoint not yet deployed', { type: 'info', duration: 3000 });
    } finally {
      setGenerating(false);
    }
  }, [toast]);

  const handleSubscribe = useCallback((e) => {
    e.preventDefault();
    if (email.trim()) {
      setSubscribed(true);
      setEmail('');
      toast.success('Subscribed to blog updates!');
    }
  }, [email, toast]);

  if (selectedPost) {
    return (
      <div className="min-h-screen bg-black text-white font-mono pt-24 px-4">
        <div className="max-w-3xl mx-auto">
          <button
            onClick={() => setSelectedPost(null)}
            className="text-cyan/60 text-sm mb-8 hover:text-cyan transition-colors"
          >
            ← Back to Blog
          </button>
          <div className="mb-6">
            <span
              className="px-3 py-1 text-[10px] uppercase tracking-wider rounded"
              style={{ background: `${selectedPost.color}22`, color: selectedPost.color }}
            >
              {selectedPost.category}
            </span>
            <span className="text-cyan/40 text-xs ml-3">{selectedPost.date} · {selectedPost.readTime}</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-bold mb-6" style={{ color: selectedPost.color }}>
            {selectedPost.title}
          </h1>
          <div className="glass-card p-8 mb-12">
            <p className="text-cyan/70 leading-relaxed text-sm mb-6">{selectedPost.excerpt}</p>
            <p className="text-cyan/50 leading-relaxed text-sm">
              This is a preview of the full article. The complete post will be generated by AI and published automatically
              on a regular schedule. Each article targets specific keywords that your potential customers are searching for,
              driving organic traffic directly to your product pages.
            </p>
            <div className="mt-8 p-4 border border-cyan/20 rounded-lg">
              <p className="text-cyan text-sm font-bold mb-2">Ready to automate your revenue?</p>
              <a href="/" className="text-xs text-acid hover:text-acid/80 transition-colors">
                Explore our products →
              </a>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white font-mono pt-24 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-3xl md:text-4xl font-bold text-cyan uppercase tracking-wider mb-4">
            AI Revenue Blog
          </h1>
          <p className="text-cyan/50 max-w-2xl mx-auto mb-6">
            Automated insights on AI agents, revenue generation, and autonomous business systems.
            New posts generated by AI every week.
          </p>
          <button
            onClick={handleGeneratePost}
            disabled={generating}
            className="px-6 py-2 text-xs font-bold uppercase tracking-wider border border-acid text-acid rounded-lg hover:bg-acid hover:text-black transition-all disabled:opacity-50"
          >
            {generating ? '⏳ Generating...' : '🤖 Generate New Post'}
          </button>
        </div>

        {/* Posts Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
          {posts.map((post, i) => (
            <div
              key={post.slug + i}
              className="glass-card p-6 hover:scale-105 transition-all cursor-pointer group"
              style={{ borderColor: `${post.color}33` }}
              onClick={() => setSelectedPost(post)}
            >
              <div
                className="w-full h-1 rounded mb-4"
                style={{ background: `linear-gradient(90deg, ${post.color}, transparent)` }}
              />
              <div className="flex items-center gap-2 mb-3">
                <span
                  className="px-2 py-0.5 text-[10px] uppercase tracking-wider rounded"
                  style={{ background: `${post.color}22`, color: post.color }}
                >
                  {post.category}
                </span>
                <span className="text-cyan/30 text-[10px]">{post.readTime}</span>
              </div>
              <h3 className="text-sm font-bold mb-3 group-hover:text-cyan transition-colors" style={{ color: post.color }}>
                {post.title}
              </h3>
              <p className="text-xs text-cyan/50 leading-relaxed mb-4">{post.excerpt}</p>
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-cyan/30">{post.date}</span>
                <span className="text-xs text-cyan/60 group-hover:text-cyan transition-colors">Read →</span>
              </div>
            </div>
          ))}
        </div>

        {/* Email Capture */}
        <div className="max-w-2xl mx-auto text-center mb-16">
          <h2 className="text-xl font-bold text-cyan uppercase tracking-wider mb-4">
            Get Weekly AI Revenue Insights
          </h2>
          <p className="text-cyan/50 mb-6">
            Join 500+ founders getting weekly tips on autonomous revenue generation.
          </p>
          {subscribed ? (
            <div className="glass-card p-4 border-acid/30">
              <span className="text-acid text-sm font-bold">✓ You're subscribed!</span>
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

        {/* SEO Content */}
        <div className="glass-card p-8 mb-16">
          <h2 className="text-lg font-bold text-purple uppercase tracking-wider mb-4">
            About AI Revenue Generation
          </h2>
          <div className="space-y-4 text-sm text-cyan/60 leading-relaxed">
            <p>
              AI revenue generation is the practice of using autonomous AI agents to create, manage, and optimize
              income streams with minimal human intervention. By deploying specialized agents for tasks like content
              creation, affiliate marketing, grant writing, and customer acquisition, businesses can scale their
              revenue operations 24/7.
            </p>
            <p>
              The Koola10 platform provides a complete AI revenue engine with four core verticals: SaaS subscriptions,
              digital product sales, API access, and managed services. Each vertical is powered by dedicated AI agents
              that work continuously to find opportunities, create deliverables, and optimize for maximum revenue.
            </p>
            <p>
              Whether you're a solo founder looking to automate your income or an enterprise team scaling your
              revenue operations, AI agents offer a proven path to autonomous growth. The key is choosing the right
              tools, setting up the right systems, and letting the AI do what it does best — work tirelessly,
              analyze data, and optimize for results.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
