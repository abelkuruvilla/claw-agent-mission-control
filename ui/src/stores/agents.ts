import { create } from 'zustand';
import type { Agent } from '@/types';
import { agentsApi } from '@/services/api';

interface AgentsState {
  agents: Agent[];
  loading: boolean;
  error: string | null;
  selectedAgent: Agent | null;
  /** Tracks in-flight fetchAgents to prevent duplicate concurrent requests */
  _fetchPromise: Promise<void> | null;

  fetchAgents: () => Promise<void>;
  fetchAgent: (id: string) => Promise<void>;
  createAgent: (agent: Partial<Agent>) => Promise<Agent>;
  updateAgent: (id: string, updates: Partial<Agent>) => Promise<void>;
  deleteAgent: (id: string) => Promise<void>;
  setSelectedAgent: (agent: Agent | null) => void;
}

export const useAgentsStore = create<AgentsState>((set, get) => ({
  agents: [],
  loading: false,
  error: null,
  selectedAgent: null,
  _fetchPromise: null,

  fetchAgents: async () => {
    // Dedup: reuse in-flight request
    const existing = get()._fetchPromise;
    if (existing) {
      return existing;
    }

    set({ loading: true, error: null });

    const promise = (async () => {
      try {
        const agents = await agentsApi.list();
        set({ agents, loading: false, _fetchPromise: null });
      } catch (error) {
        set({ error: (error as Error).message, loading: false, _fetchPromise: null });
      }
    })();

    set({ _fetchPromise: promise });
    return promise;
  },

  fetchAgent: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const agent = await agentsApi.get(id);
      set({ selectedAgent: agent, loading: false });
    } catch (error) {
      set({ error: (error as Error).message, loading: false });
    }
  },

  createAgent: async (agent: Partial<Agent>) => {
    const created = await agentsApi.create(agent);
    set((state) => ({ agents: [...state.agents, created] }));
    return created;
  },

  updateAgent: async (id: string, updates: Partial<Agent>) => {
    const updated = await agentsApi.update(id, updates);
    set((state) => ({
      agents: state.agents.map((a) => (a.id === id ? updated : a)),
      selectedAgent: state.selectedAgent?.id === id ? updated : state.selectedAgent,
    }));
  },

  deleteAgent: async (id: string) => {
    await agentsApi.delete(id);
    set((state) => ({
      agents: state.agents.filter((a) => a.id !== id),
      selectedAgent: state.selectedAgent?.id === id ? null : state.selectedAgent,
    }));
  },

  setSelectedAgent: (agent) => set({ selectedAgent: agent }),
}));
