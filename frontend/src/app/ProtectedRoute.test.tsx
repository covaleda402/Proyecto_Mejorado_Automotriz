import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it } from 'vitest';

import { ProtectedRoute } from './ProtectedRoute';
import { SessionProvider } from '../shared/SessionContext';

function storeSession(role: 'ADMINISTRATOR' | 'TECHNICIAN') {
  window.localStorage.setItem(
    'workshop.session',
    JSON.stringify({
      token: 'test-token',
      expiresAt: new Date(Date.now() + 3600000).toISOString(),
      userId: 'user-1',
      username: 'testuser',
      fullName: 'Test User',
      role,
    }),
  );
}

function clearSession() {
  window.localStorage.removeItem('workshop.session');
}

function renderWithRoute(element: React.ReactNode, initialEntry = '/protected') {
  return render(
    <SessionProvider>
      <MemoryRouter initialEntries={[initialEntry]}>
        <Routes>
          <Route path="/login" element={<div>Login Page</div>} />
          <Route path="/dashboard" element={<div>Dashboard Page</div>} />
          <Route path="/protected" element={element} />
        </Routes>
      </MemoryRouter>
    </SessionProvider>,
  );
}

describe('ProtectedRoute', () => {
  beforeEach(() => {
    clearSession();
  });

  it('redirects to /login when no session exists', () => {
    renderWithRoute(
      <ProtectedRoute>
        <div>Secret Content</div>
      </ProtectedRoute>,
    );

    expect(screen.getByText('Login Page')).toBeInTheDocument();
    expect(screen.queryByText('Secret Content')).not.toBeInTheDocument();
  });

  it('renders children when session exists and no specific role is required', () => {
    storeSession('TECHNICIAN');
    renderWithRoute(
      <ProtectedRoute>
        <div>Secret Content</div>
      </ProtectedRoute>,
    );

    expect(screen.getByText('Secret Content')).toBeInTheDocument();
  });

  it('allows ADMINISTRATOR to access route with requiredRole="ADMINISTRATOR"', () => {
    storeSession('ADMINISTRATOR');
    renderWithRoute(
      <ProtectedRoute requiredRole="ADMINISTRATOR">
        <div>Admin Content</div>
      </ProtectedRoute>,
    );

    expect(screen.getByText('Admin Content')).toBeInTheDocument();
  });

  it('redirects non-administrators to /dashboard when requiredRole="ADMINISTRATOR"', () => {
    storeSession('TECHNICIAN');
    renderWithRoute(
      <ProtectedRoute requiredRole="ADMINISTRATOR">
        <div>Admin Content</div>
      </ProtectedRoute>,
    );

    expect(screen.getByText('Dashboard Page')).toBeInTheDocument();
    expect(screen.queryByText('Admin Content')).not.toBeInTheDocument();
  });
});
