/*
 * Session service: it signs the user in and is the only place that reads or
 * writes the stored session, so no component touches browser storage.
 */
import { request } from './api_client';

const STORAGE_KEY = 'workshop.session';

export interface Session {
  token?: string;
  expiresAt: string;
  userId: string;
  username: string;
  fullName: string;
  role: 'ADMINISTRATOR' | 'TECHNICIAN';
}

export async function signIn(username: string, password: string): Promise<Session> {
  return request<Session>('/session', { method: 'POST', body: { username, password } });
}

export function readStoredSession(): Session | null {
  try {
    // Ensure localStorage contains no session tokens
    window.localStorage?.removeItem(STORAGE_KEY);
    const raw = window.sessionStorage?.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const stored = JSON.parse(raw) as Session;
    if (new Date(stored.expiresAt).getTime() <= Date.now()) {
      window.sessionStorage?.removeItem(STORAGE_KEY);
      return null;
    }
    return stored;
  } catch {
    return null;
  }
}

export function storeSession(session: Session): void {
  try {
    // Ensure localStorage contains no session tokens
    window.localStorage?.removeItem(STORAGE_KEY);
    // Only store safe display metadata in sessionStorage (the real auth token is in the HttpOnly cookie)
    const safeData: Session = {
      expiresAt: session.expiresAt,
      userId: session.userId,
      username: session.username,
      fullName: session.fullName,
      role: session.role,
    };
    window.sessionStorage?.setItem(STORAGE_KEY, JSON.stringify(safeData));
  } catch {
    // A browser with storage disabled still works in-memory
  }
}

export function clearStoredSession(): void {
  try {
    window.localStorage?.removeItem(STORAGE_KEY);
    window.sessionStorage?.removeItem(STORAGE_KEY);
    void request('/session/logout', { method: 'POST' }).catch(() => {});
  } catch {
    // Nothing to clean
  }
}
