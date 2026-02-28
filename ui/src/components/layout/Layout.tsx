'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  LayoutDashboard, 
  Bot, 
  ListTodo, 
  Activity, 
  Settings,
  Cpu,
  FolderKanban,
  Menu,
  X
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useIsMobile } from '@/hooks/useIsMobile';
import { useWebSocket } from '@/hooks/useWebSocket';

interface LayoutProps {
  children: React.ReactNode;
}

/** Navigation items - static constant, never recreated */
const NAV_ITEMS = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/agents', icon: Bot, label: 'Agents' },
  { path: '/projects', icon: FolderKanban, label: 'Projects' },
  { path: '/tasks', icon: ListTodo, label: 'Tasks' },
  { path: '/events', icon: Activity, label: 'Events' },
  { path: '/settings', icon: Settings, label: 'Settings' },
] as const;

export function Layout({ children }: LayoutProps) {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const isMobile = useIsMobile();
  const { isConnected } = useWebSocket();

  // Close sidebar when navigating on mobile
  useEffect(() => {
    if (isMobile) {
      setSidebarOpen(false);
    }
  }, [pathname, isMobile]);

  // Close sidebar when switching from mobile to desktop
  useEffect(() => {
    if (!isMobile) {
      setSidebarOpen(false);
    }
  }, [isMobile]);

  const isActive = useCallback((path: string) => {
    if (path === '/') return pathname === '/';
    return pathname.startsWith(path);
  }, [pathname]);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen(prev => !prev);
  }, []);

  const closeSidebar = useCallback(() => {
    setSidebarOpen(false);
  }, []);

  return (
    <div className="h-screen min-h-screen flex flex-col bg-[#0d1117] overflow-hidden">
      {/* Mobile Header */}
      <header className="fixed top-0 left-0 right-0 z-50 flex h-14 items-center justify-between border-b border-[#30363d] bg-[#161b22] px-4 md:hidden">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
            <Cpu className="h-4 w-4 text-white" />
          </div>
          <span className="font-semibold text-white text-sm">Mission Control</span>
        </div>
        <button
          onClick={toggleSidebar}
          className="flex h-10 w-10 items-center justify-center rounded-lg text-[#8b949e] hover:bg-[#21262d] hover:text-white"
        >
          {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </header>

      {/* Mobile Sidebar Overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={closeSidebar}
        />
      )}

      {/* Sidebar */}
      <aside className={cn(
        "fixed top-0 left-0 z-50 h-screen w-64 border-r border-[#30363d] bg-[#161b22] transition-transform duration-300 ease-in-out",
        "md:translate-x-0",
        isMobile && !sidebarOpen ? "-translate-x-full" : "translate-x-0",
        isMobile && "pt-14"
      )}>
        {/* Logo - hidden on mobile (shown in header) */}
        <div className="hidden md:flex h-16 items-center gap-3 border-b border-[#30363d] px-6">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
            <Cpu className="h-6 w-6 text-white" />
          </div>
          <div>
            <h1 className="font-semibold text-white">Claw Agent</h1>
            <p className="text-xs text-[#8b949e]">Mission Control</p>
          </div>
        </div>

        {/* Navigation */}
        <nav className="space-y-1 p-4">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            const active = isActive(item.path);
            
            return (
              <Link
                key={item.path}
                href={item.path}
                onClick={() => isMobile && closeSidebar()}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-colors",
                  active
                    ? "bg-blue-600 text-white"
                    : "text-[#8b949e] hover:bg-[#21262d] hover:text-white"
                )}
              >
                <Icon className="h-5 w-5" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Footer */}
        <div className="absolute bottom-0 left-0 right-0 border-t border-[#30363d] p-4">
          <div className="text-xs text-[#8b949e]">
            <div>OpenClaw v1.0</div>
            <div className="mt-1 flex items-center gap-1.5">
              <div className={`h-1.5 w-1.5 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
              {isConnected ? 'Connected' : 'Disconnected'}
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content - fills viewport so pages can use full height */}
      <main className={cn(
        "flex-1 flex flex-col min-h-0 transition-all duration-300",
        "md:ml-64",
        "pt-14 md:pt-0"
      )}>
        <div className="flex-1 flex flex-col min-h-0 overflow-auto p-4 md:p-8">
          {children}
        </div>
      </main>
    </div>
  );
}
