'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { useAgentsStore } from '@/stores/agents';
import { useTasksStore } from '@/stores/tasks';
import { useEventsStore } from '@/stores/events';

interface WebSocketMessage {
  type: string;
  payload: Record<string, unknown>;
  timestamp: string;
}

/** Minimum interval (ms) between re-fetches for the same resource type */
const DEBOUNCE_MS = 1000;

/**
 * WebSocket hook with:
 * - Debounced re-fetches to prevent rapid-fire API calls from burst messages
 * - Proper cleanup of timers on unmount
 * - Exponential backoff for reconnection (caps at 30s)
 * - Connection state tracking
 */
export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isMountedRef = useRef(true);
  const [isConnected, setIsConnected] = useState(false);

  // Use refs for store actions to avoid stale closures
  const fetchAgentsRef = useRef(useAgentsStore.getState().fetchAgents);
  const fetchTasksRef = useRef(useTasksStore.getState().fetchTasks);
  const addEventRef = useRef(useEventsStore.getState().addEvent);

  // Keep refs up to date
  useEffect(() => {
    const unsubAgents = useAgentsStore.subscribe(
      (state) => { fetchAgentsRef.current = state.fetchAgents; }
    );
    const unsubTasks = useTasksStore.subscribe(
      (state) => { fetchTasksRef.current = state.fetchTasks; }
    );
    const unsubEvents = useEventsStore.subscribe(
      (state) => { addEventRef.current = state.addEvent; }
    );
    return () => {
      unsubAgents();
      unsubTasks();
      unsubEvents();
    };
  }, []);

  // Debounce timers for each resource type
  const debounceTimersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  /**
   * Debounced fetch: coalesces rapid sequential messages of the same type
   * into a single API call. Prevents N rapid WebSocket events from
   * triggering N separate fetch calls.
   */
  const debouncedFetch = useCallback((key: string, fetchFn: () => void) => {
    const existing = debounceTimersRef.current.get(key);
    if (existing) {
      clearTimeout(existing);
    }
    const timer = setTimeout(() => {
      debounceTimersRef.current.delete(key);
      if (isMountedRef.current) {
        console.log(`[WebSocket] Debounced fetch: ${key}`);
        fetchFn();
      }
    }, DEBOUNCE_MS);
    debounceTimersRef.current.set(key, timer);
  }, []);

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      console.log(`[WebSocket] Received: ${message.type}`);

      switch (message.type) {
        case 'agent.status':
          debouncedFetch('agents', () => fetchAgentsRef.current());
          break;
        case 'task.status':
        case 'phase.updated':
        case 'story.updated':
          debouncedFetch('tasks', () => fetchTasksRef.current());
          break;
        case 'event.new': {
          // Events are lightweight - add directly without debounce
          const payload = message.payload as { type?: string; [k: string]: unknown };
          addEventRef.current(payload as any);
          // Status changes from API (logEvent) may not send task.status; trigger task refetch so UI updates
          if (payload?.type === 'status_changed') {
            debouncedFetch('tasks', () => fetchTasksRef.current());
          }
          break;
        }
        default:
          console.log(`[WebSocket] Unknown message type: ${message.type}`);
      }
    } catch (e) {
      console.error('[WebSocket] Failed to parse message:', e);
    }
  }, [debouncedFetch]);

  const connect = useCallback(() => {
    if (!isMountedRef.current) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    try {
      console.log('[WebSocket] Connecting...');
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!isMountedRef.current) {
          ws.close();
          return;
        }
        console.log('[WebSocket] Connected');
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = handleMessage;

      ws.onclose = () => {
        console.log('[WebSocket] Disconnected');
        setIsConnected(false);
        wsRef.current = null;

        if (!isMountedRef.current) return;

        // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s max
        const delay = Math.min(
          1000 * Math.pow(2, reconnectAttemptsRef.current),
          30000
        );
        reconnectAttemptsRef.current++;
        console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`);
        reconnectTimeoutRef.current = setTimeout(connect, delay);
      };

      ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
      };
    } catch (error) {
      console.error('[WebSocket] Failed to connect:', error);
      if (isMountedRef.current) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
        reconnectAttemptsRef.current++;
        reconnectTimeoutRef.current = setTimeout(connect, delay);
      }
    }
  }, [handleMessage]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;

      // Clear reconnect timer
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      // Clear all debounce timers
      debounceTimersRef.current.forEach(timer => clearTimeout(timer));
      debounceTimersRef.current.clear();

      // Close WebSocket
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  return { isConnected };
}
