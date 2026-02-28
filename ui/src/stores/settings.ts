import { create } from 'zustand';
import type { Settings } from '@/types';
import { settingsApi } from '@/services/api';

interface SettingsState {
  settings: Settings | null;
  loading: boolean;
  error: string | null;
  
  fetchSettings: () => Promise<void>;
  updateSettings: (updates: Partial<Settings>) => Promise<void>;
  testConnection: () => Promise<boolean>;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  settings: null,
  loading: false,
  error: null,
  
  fetchSettings: async () => {
    set({ loading: true, error: null });
    try {
      const settings = await settingsApi.get();
      set({ settings, loading: false });
    } catch (error) {
      set({ error: (error as Error).message, loading: false });
    }
  },
  
  updateSettings: async (updates) => {
    const updated = await settingsApi.update(updates);
    set({ settings: updated });
  },
  
  testConnection: async () => {
    const result = await settingsApi.testConnection();
    return result.connected;
  },
}));
