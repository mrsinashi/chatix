import { create } from 'zustand';
import { getSystemSettings, SystemSettings } from '../api/settingsApi';

interface SettingsState {
  system: SystemSettings;
  isLoading: boolean;
  loaded: boolean;
  loadSystem: () => Promise<void>;
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  system: {},
  isLoading: false,
  loaded: false,

  loadSystem: async () => {
    if (get().loaded || get().isLoading) return;
    set({ isLoading: true });
    try {
      const settings = await getSystemSettings();
      set({ system: settings, loaded: true, isLoading: false });

      // Обновляем заголовок окна
      const orgName = typeof settings.org_name === 'string' ? settings.org_name : 'Название организации';
      document.title = `Chatix – ${orgName}`;
    } catch {
      set({ isLoading: false, loaded: true });
    }
  },
}));