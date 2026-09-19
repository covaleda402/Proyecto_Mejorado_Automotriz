/* Technician allocation and employee lifecycle service. */
import { request } from './api_client';

export interface Technician {
  id: string;
  userId: string;
  fullName: string;
  specialty: string;
  isActive: boolean;
  busy: boolean;
  canReceiveAssignment: boolean;
  activeOrderId: string;
  activeOrderNumber: string;
  activeVehiclePlate: string;
}

export interface CreateTechnicianPayload {
  fullName: string;
  username: string;
  password: string;
  specialty: string;
}

export function listTechnician(token: string): Promise<Technician[]> {
  return request<Technician[]>('/technician', { token });
}

export function createTechnician(token: string, payload: CreateTechnicianPayload): Promise<Technician> {
  return request<Technician>('/technician', {
    method: 'POST',
    token,
    body: payload,
  });
}

export function setTechnicianAccess(token: string, technicianId: string, active: boolean): Promise<Technician> {
  return request<Technician>(`/technician/${technicianId}/access`, {
    method: 'PATCH',
    token,
    body: { active },
  });
}
