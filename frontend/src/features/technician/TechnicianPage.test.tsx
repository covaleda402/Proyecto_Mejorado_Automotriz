import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TechnicianPage } from './TechnicianPage';
import { SessionProvider } from '../../shared/SessionContext';

const MOCK_TECHNICIANS = [
  {
    id: 'tech-1',
    userId: 'user-tech-1',
    fullName: 'Juan Perez',
    specialty: 'Frenos y suspension',
    isActive: true,
    busy: false,
    canReceiveAssignment: true,
    activeOrderId: '',
    activeOrderNumber: '',
    activeVehiclePlate: '',
  },
  {
    id: 'tech-2',
    userId: 'user-tech-2',
    fullName: 'Laura Ramirez',
    specialty: 'Electricidad y diagnostico',
    isActive: false,
    busy: true,
    canReceiveAssignment: false,
    activeOrderId: 'order-1',
    activeOrderNumber: 'OS-0001',
    activeVehiclePlate: 'ABC123',
  },
];

function storeSession(role: 'ADMINISTRATOR' | 'TECHNICIAN' = 'ADMINISTRATOR') {
  window.localStorage.setItem(
    'workshop.session',
    JSON.stringify({
      token: 'test-token',
      expiresAt: new Date(Date.now() + 3600000).toISOString(),
      userId: 'admin-1',
      username: 'admin',
      fullName: 'Administrador General',
      role,
    }),
  );
}

function jsonResponse(status: number, payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function renderTechnicianPage() {
  return render(
    <SessionProvider>
      <TechnicianPage />
    </SessionProvider>,
  );
}

describe('TechnicianPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders technician list with active and revoked access badges', async () => {
    storeSession('ADMINISTRATOR');
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation((url: string) => {
        if (url.includes('/api/technician')) {
          return Promise.resolve(jsonResponse(200, MOCK_TECHNICIANS));
        }
        return Promise.resolve(jsonResponse(200, []));
      }),
    );

    renderTechnicianPage();

    expect(await screen.findByText('Juan Perez')).toBeInTheDocument();
    expect(screen.getByText('Laura Ramirez')).toBeInTheDocument();
    expect(screen.getByText('Activo')).toBeInTheDocument();
    expect(screen.getByText('Sin acceso')).toBeInTheDocument();
    expect(screen.getByText('Disponible')).toBeInTheDocument();
    expect(screen.getByText('Ocupado')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Revocar acceso' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Reactivar acceso' })).toBeInTheDocument();
  });

  it('opens onboarding form and successfully registers a new technician', async () => {
    storeSession('ADMINISTRATOR');
    const fetchMock = vi.fn().mockImplementation((url: string, init?: RequestInit) => {
      if (url.includes('/api/technician') && init?.method === 'POST') {
        return Promise.resolve(
          jsonResponse(201, {
            id: 'tech-3',
            userId: 'user-tech-3',
            fullName: 'Carlos Mendoza',
            specialty: 'Transmisiones',
            isActive: true,
            busy: false,
            canReceiveAssignment: true,
          }),
        );
      }
      return Promise.resolve(jsonResponse(200, MOCK_TECHNICIANS));
    });
    vi.stubGlobal('fetch', fetchMock);

    renderTechnicianPage();

    expect(await screen.findByText('Juan Perez')).toBeInTheDocument();

    const openBtn = screen.getByRole('button', { name: '+ Registrar nuevo tecnico' });
    await userEvent.click(openBtn);

    expect(screen.getByText('Alta de nuevo tecnico')).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText(/Nombre completo/i), 'Carlos Mendoza');
    await userEvent.type(screen.getByLabelText(/Nombre de usuario/i), 'cmendoza');
    await userEvent.type(screen.getByLabelText(/Contrasena inicial/i), 'Password#2026!');
    await userEvent.type(screen.getByLabelText(/Especialidad tecnica/i), 'Transmisiones');

    await userEvent.click(screen.getByRole('button', { name: 'Registrar tecnico' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining('/api/technician'),
        expect.objectContaining({
          method: 'POST',
        }),
      );
    });
  });

  it('toggles technician access with confirmation', async () => {
    storeSession('ADMINISTRATOR');
    vi.spyOn(window, 'confirm').mockReturnValue(true);

    const fetchMock = vi.fn().mockImplementation((url: string, init?: RequestInit) => {
      if (url.includes('/api/technician/tech-1/access') && init?.method === 'PATCH') {
        return Promise.resolve(
          jsonResponse(200, {
            id: 'tech-1',
            isActive: false,
          }),
        );
      }
      return Promise.resolve(jsonResponse(200, MOCK_TECHNICIANS));
    });
    vi.stubGlobal('fetch', fetchMock);

    renderTechnicianPage();

    expect(await screen.findByText('Juan Perez')).toBeInTheDocument();
    const revokeBtn = screen.getByRole('button', { name: 'Revocar acceso' });
    await userEvent.click(revokeBtn);

    expect(window.confirm).toHaveBeenCalled();
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining('/api/technician/tech-1/access'),
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ active: false }),
        }),
      );
    });
  });
});
