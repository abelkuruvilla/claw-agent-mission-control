'use client';

import { useState, useEffect, useCallback } from 'react';

/** Shared breakpoint constant for mobile detection */
const MOBILE_BREAKPOINT = 768;

/**
 * Shared hook that detects whether the viewport is mobile-sized.
 * Uses a single resize listener with debounce to reduce layout thrashing.
 * Replaces duplicate mobile detection logic in Layout and TasksPage.
 * 
 * Returns undefined during SSR/initial render to prevent hydration mismatch.
 */
export function useIsMobile(): boolean | undefined {
  const [isMobile, setIsMobile] = useState<boolean | undefined>(undefined);

  const checkMobile = useCallback(() => {
    // Check if window is available (client-side)
    if (typeof window !== 'undefined') {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    }
  }, []);

  useEffect(() => {
    // Initial check on mount
    checkMobile();

    // Debounced resize handler to avoid excessive re-renders
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    const handleResize = () => {
      if (timeoutId) clearTimeout(timeoutId);
      timeoutId = setTimeout(checkMobile, 150);
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [checkMobile]);

  return isMobile;
}
