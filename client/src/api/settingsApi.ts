const API_BASE = '/api/v1';

export interface SystemSettings {
  'org.name'?: string;
  'client.theme'?: string;
  [key: string]: unknown;
}

export async function getSystemSettings(): Promise<SystemSettings> {
  const response = await fetch(`${API_BASE}/settings/system`);

  if (!response.ok) {
    throw new Error('Ошибка получения настроек');
  }

  return response.json();
}