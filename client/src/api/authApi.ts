const API_BASE = '/api/v1';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface UserInfo {
  id: string;
  username: string;
  display_name: string;
  email?: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: UserInfo;
}

export interface MeResponse {
  user: UserInfo;
  permissions: string[];
}

export async function login(data: LoginRequest): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || 'Ошибка входа');
  }

  return response.json();
}

export async function getMe(token: string): Promise<MeResponse> {
  const response = await fetch(`${API_BASE}/auth/me`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });

  if (!response.ok) {
    throw new Error('Сессия недействительна');
  }

  return response.json();
}

export async function logout(token: string): Promise<void> {
  const response = await fetch(`${API_BASE}/auth/logout`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
  });

  if (!response.ok) {
    throw new Error('Ошибка выхода');
  }
}