import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ServiceOrderDetailPage } from './ServiceOrderDetailPage';
import { SessionProvider } from '../../shared/SessionContext';

const BASE_ORDER = {
  id: 'order-123',
  orderNumber: 'OS-0001',
  vehicleId: 'vehicle-1',
  vehiclePlate: 'ABC123',
  technicianName: 'Carlos Gomez',
  reportedFailure: 'Falla en frenos',
  status: 'IN_DIAGNOSIS',
  receivedAt: '2026-03-01T10:00:00Z',
  updatedAt: '2026-03-01T10:00:00Z',
  permissions: {
    canAdvance: true,
    canAddDiagnostic: true,
    canAddIntervention: false,
    canAssign: false,
  },
};

function storeSession(role: 'ADMINISTRATOR' | 'TECHNICIAN' = 'ADMINISTRATOR') {
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

function stubApis(orderData: any) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/service-order/order-123/transition')) {
        return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
      }
      if (url.includes('/api/service-order/order-123/diagnostic')) {
        return Promise.resolve(new Response(JSON.stringify({ status: 404 }), { status: 404 }));
      }
      if (url.includes('/api/service-order/order-123/intervention')) {
        return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
      }
      if (url.includes('/api/service-order/order-123/assignment')) {
        return Promise.resolve(new Response(JSON.stringify(null), { status: 404 }));
      }
      if (url.includes('/api/service-order/order-123')) {
        return Promise.resolve(new Response(JSON.stringify(orderData), { status: 200 }));
      }
      if (url.includes('/api/technician')) {
        return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
      }
      if (url.includes('/api/warranty')) {
        return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify([]), { status: 200 }));
    }),
  );
}

function renderDetailPage() {
  return render(
    <SessionProvider>
      <MemoryRouter initialEntries={['/service-orders/order-123']}>
        <Routes>
          <Route path="/service-orders/:serviceOrderId" element={<ServiceOrderDetailPage />} />
        </Routes>
      </MemoryRouter>
    </SessionProvider>,
  );
}

describe('ServiceOrderDetailPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('displays the assigned technician name prominently in the main card', async () => {
    storeSession('ADMINISTRATOR');
    stubApis(BASE_ORDER);
    renderDetailPage();

    expect(await screen.findByText('OS-0001')).toBeInTheDocument();
    expect(screen.getByText(/Técnico responsable asignado:/i)).toBeInTheDocument();
    expect(screen.getByText('Carlos Gomez')).toBeInTheDocument();
  });

  it('displays "Sin asignar" when technicianName is empty', async () => {
    storeSession('ADMINISTRATOR');
    stubApis({ ...BASE_ORDER, technicianName: '' });
    renderDetailPage();

    expect(await screen.findByText('OS-0001')).toBeInTheDocument();
    expect(screen.getByText(/Sin asignar/i)).toBeInTheDocument();
  });

  it('shows the advance status button when permissions.canAdvance is true', async () => {
    storeSession('TECHNICIAN');
    stubApis({
      ...BASE_ORDER,
      status: 'IN_DIAGNOSIS',
      permissions: { canAdvance: true, canAddDiagnostic: true, canAddIntervention: false, canAssign: false },
    });
    renderDetailPage();

    expect(await screen.findByText('OS-0001')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Marcar en reparacion/i })).toBeInTheDocument();
  });

  it('hides the advance status button when permissions.canAdvance is false', async () => {
    storeSession('TECHNICIAN');
    stubApis({
      ...BASE_ORDER,
      status: 'IN_DIAGNOSIS',
      permissions: { canAdvance: false, canAddDiagnostic: false, canAddIntervention: false, canAssign: false },
    });
    renderDetailPage();

    expect(await screen.findByText('OS-0001')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Marcar en reparacion/i })).not.toBeInTheDocument();
  });

  it('displays warning and badge when assigned technician account has access disabled', async () => {
    storeSession('ADMINISTRATOR');
    stubApis({
      ...BASE_ORDER,
      technicianIsActive: false,
    });
    renderDetailPage();

    expect(await screen.findByText('OS-0001')).toBeInTheDocument();
    expect(screen.getByText('Sin acceso')).toBeInTheDocument();
    expect(screen.getByText(/El técnico responsable asignado tiene el acceso inhabilitado/i)).toBeInTheDocument();
  });
});
