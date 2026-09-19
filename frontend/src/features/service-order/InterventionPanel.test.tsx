import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { InterventionPanel } from './InterventionPanel';
import { SessionProvider } from '../../shared/SessionContext';

const INTERVENTIONS_MOCK = [
  {
    id: 'int-1',
    serviceOrderId: 'order-123',
    technicianId: 'tech-1',
    description: 'Cambio de pastillas de freno',
    laborHourCount: 2,
    performedAt: '2026-03-01T11:00:00Z',
    part: [{ partName: 'Pastillas delanteras', quantity: 1 }],
  },
  {
    id: 'int-2',
    serviceOrderId: 'order-123',
    technicianId: 'tech-1',
    description: 'Alineación y balanceo',
    laborHourCount: 1,
    performedAt: '2026-03-01T12:00:00Z',
    part: [],
  },
];

const WARRANTIES_MOCK = [
  {
    id: 'war-1',
    interventionId: 'int-1',
    orderNumber: 'OS-0001',
    vehiclePlate: 'ABC123',
    kind: 'LABOR',
    coverageMonthCount: 12,
    issuedAt: '2026-03-01T12:00:00Z',
    expirationDate: '2027-03-01T12:00:00Z',
    valid: true,
  },
];

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

function stubInterventionApis(interventions = INTERVENTIONS_MOCK, warranties = WARRANTIES_MOCK) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/service-order/order-123/intervention')) {
        return Promise.resolve(new Response(JSON.stringify(interventions), { status: 200 }));
      }
      if (url.includes('/api/warranty')) {
        return Promise.resolve(new Response(JSON.stringify(warranties), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
    }),
  );
}

describe('InterventionPanel', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('shows the intervention registration form when orderPermissions.canAddIntervention is true', async () => {
    storeSession('TECHNICIAN');
    stubInterventionApis(INTERVENTIONS_MOCK, WARRANTIES_MOCK);

    render(
      <SessionProvider>
        <InterventionPanel
          serviceOrderId="order-123"
          onChange={() => {}}
          orderPermissions={{
            canAdvance: false,
            canAddDiagnostic: false,
            canAddIntervention: true,
            canAssign: false,
          }}
        />
      </SessionProvider>,
    );

    expect(await screen.findByLabelText(/Descripcion/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Horas de trabajo/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Registrar intervencion/i })).toBeInTheDocument();
  });

  it('shows read-only message when orderPermissions.canAddIntervention is false', async () => {
    storeSession('TECHNICIAN');
    stubInterventionApis(INTERVENTIONS_MOCK, WARRANTIES_MOCK);

    render(
      <SessionProvider>
        <InterventionPanel
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
      await screen.findByText(/No tienes permisos para registrar intervenciones en esta orden/i),
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Registrar intervencion/i })).not.toBeInTheDocument();
  });

  it('displays the warranty validity badge in the table for an intervention with warranty', async () => {
    storeSession('TECHNICIAN');
    stubInterventionApis(INTERVENTIONS_MOCK, WARRANTIES_MOCK);

    render(
      <SessionProvider>
        <InterventionPanel
          serviceOrderId="order-123"
          onChange={() => {}}
          orderPermissions={{
            canAdvance: false,
            canAddDiagnostic: false,
            canAddIntervention: true,
            canAssign: false,
          }}
        />
      </SessionProvider>,
    );

    expect(await screen.findByText('Cambio de pastillas de freno')).toBeInTheDocument();
    expect(screen.getByText('Vigente')).toBeInTheDocument();
    expect(screen.getByText('Sin garantía')).toBeInTheDocument();
  });
});
