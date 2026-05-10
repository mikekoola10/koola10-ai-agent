'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LayoutDashboard, Video, Cpu } from 'lucide-react';
import { cn } from '@/lib/utils';

const navItems = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Live Display', href: '/live', icon: Video },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-64 bg-black/20 backdrop-blur-xl border-r border-white/10 flex flex-col h-screen sticky top-0">
      <div className="p-6">
        <div className="flex items-center gap-3 mb-8">
          <div className="p-2 bg-amber-400 rounded-lg shadow-[0_0_15px_rgba(251,191,36,0.4)]">
            <Cpu className="h-6 w-6 text-black" />
          </div>
          <h1 className="text-xl font-black text-white tracking-tight">
            KOOLA10
          </h1>
        </div>

        <nav className="space-y-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.href;

            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group",
                  isActive
                    ? "bg-amber-400 text-black font-bold shadow-[0_0_20px_rgba(251,191,36,0.2)]"
                    : "text-white/60 hover:text-white hover:bg-white/5"
                )}
              >
                <Icon className={cn("h-5 w-5", isActive ? "text-black" : "text-amber-400/70 group-hover:text-amber-400")} />
                {item.name}
              </Link>
            );
          })}
        </nav>
      </div>

      <div className="mt-auto p-6 border-t border-white/5">
        <p className="text-[10px] text-white/20 font-bold uppercase tracking-[0.2em]">
          System Status: Nominal
        </p>
      </div>
    </aside>
  );
}
