import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DiagnosticPanel } from './DiagnosticPanel';
import { SessionProvider } from '../../shared/SessionContext';

function storeSession(role: 'ADMINISTRATOR' | 'TECHNICIAN' = 'TECHNICIAN') {
  window.localStorage.setItem(
    'workshop.session',
    JSON.stringify({
      token: 'test-token',
      expiresAt: new Date(Date.now() + 3600000).toISOString(),
      userId: 'user-1',
      username: 'tech',
      fullName: 'Technician User',
      role,
    }),
  );
}

function stubDiagnosticApi(diagnosticData: any = null) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockImplementation(() => {
      if (diagnosticData) {
        return Promise.resolve(new Response(JSON.stringify(diagnosticData), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ status: 404 }), { status: 404 }));
    }),
  );
}

describe('DiagnosticPanel', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('enables the diagnostic form when orderPermissions.canAddDiagnostic is true', async () => {
    storeSession('TECHNICIAN');
    stubDiagnosticApi(null);

    render(
      <SessionProvider>
        <DiagnosticPanel
          serviceOrderId="order-123"
          onChange={() => {}}
          orderPermissions={{
            canAdvance: false,
            canAddDiagnostic: true,
            canAddIntervention: false,
            canAssign: false,
          }}
        />
      </SessionProvider>,
    );

    expect(await screen.findByLabelText(/Hallazgo/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Componentes a reparar/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Guardar diagnostico/i })).toBeInTheDocument();
  });

  it('displays read-only informative notice when orderPermissions.canAddDiagnostic is false', async () => {
    storeSession('TECHNICIAN');
    stubDiagnosticApi(null);

    render(
      <SessionProvider>
        <DiagnosticPanel
          serviceOrderId="order-123"
          onChange={() => {}}
          orderPermissions={{
            canAdvance: false,
            canAddDiagnostic: false,
            canAddIntervention: false,
            canAssign: false,
          }}
        />
      </SessionProvider>,
    );

    expect(
      await screen.findByText(/El diagnóstico ya fue registrado o no tienes permisos para editarlo/i),
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Guardar diagnostico/i })).not.toBeInTheDocument();
  });
});
