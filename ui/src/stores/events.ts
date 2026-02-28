import { create } from 'zustand';
import type { Event } from '@/types';
import { eventsApi } from '@/services/api';

/**
 * Maximum number of events kept in memory.
 * Prevents unbounded array growth from WebSocket event.new messages.
 */
const MAX_EVENTS = 500;

interface EventsState {
  events: Event[];
  loading: boolean;
  error: string | null;
  /** Tracks in-flight fetch to prevent duplicate concurrent requests */
  _fetchPromise: Promise<void> | null;

  fetchEvents: (params?: { task_id?: string; agent_id?: string; limit?: number }) => Promise<void>;
  addEvent: (event: Event) => void;
  clearEvents: () => void;
}

export const useEventsStore = create<EventsState>((set, get) => ({
  events: [],
  loading: false,
  error: null,
  _fetchPromise: null,

  fetchEvents: async (params) => {
    // Dedup: if a fetch is already in flight, return that promise
    const existing = get()._fetchPromise;
    if (existing) {
      return existing;
    }

    set({ loading: true, error: null });

    const promise = (async () => {
      try {
        const events = await eventsApi.list(params);
        // Only keep events up to the cap
        set({
          events: events.slice(0, MAX_EVENTS),
          loading: false,
          _fetchPromise: null,
        });
      } catch (error) {
        set({
          error: (error as Error).message,
          loading: false,
          _fetchPromise: null,
        });
      }
    })();

    set({ _fetchPromise: promise });
    return promise;
  },

  addEvent: (event) =>
    set((state) => {
      // Deduplicate: check if event already exists
      if (state.events.some(e => e.id === event.id)) {
        return state;
      }
      // Prepend new event and cap the array to prevent memory bloat
      const newEvents = [event, ...state.events];
      if (newEvents.length > MAX_EVENTS) {
        newEvents.length = MAX_EVENTS;
      }
      return { events: newEvents };
    }),

  clearEvents: () => set({ events: [], error: null }),
}));
