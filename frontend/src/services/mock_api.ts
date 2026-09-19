/*
 * Standalone mock API engine for Vercel and offline deployments.
 * Enables 100% functionality without an external database or backend server.
 * Persists changes in localStorage so reloads and user flows work seamlessly.
 */

import type { Customer, NewCustomer } from './customer_service';
import type { Vehicle, NewVehicle } from './vehicle_service';
import type { Technician } from './technician_service';
import type {
  ServiceOrder,
  ServiceOrderStatus,
  StatusTransition,
  Assignment,
  Diagnostic,
  Intervention,
  PartUsage,
} from './service_order_service';
import type { Warranty, WarrantyKind } from './warranty_service';
import type { Dashboard } from './dashboard_service';
import type { Timeline, TimelineEntry } from './timeline_service';
import type { Session } from './session_service';

const DB_KEY = 'workshop.demo.db.v1';

interface MockDatabase {
  customers: Customer[];
  vehicles: Vehicle[];
  technicians: Technician[];
  orders: ServiceOrder[];
  assignments: Assignment[];
  diagnostics: Diagnostic[];
  interventions: Intervention[];
  warranties: Warranty[];
  transitions: StatusTransition[];
}

function getInitialData(): MockDatabase {
  const customer1: Customer = {
    id: 'c1000000-0000-4000-8000-000000000001',
    fullName: 'Carlos Mendoza',
    documentNumber: '1020304050',
    phone: '3101234567',
    email: 'carlos.mendoza@email.com',
    createdAt: '2026-03-01T08:00:00Z',
  };

  const customer2: Customer = {
    id: 'c2000000-0000-4000-8000-000000000002',
    fullName: 'Ana María Gómez',
    documentNumber: '1098765432',
    phone: '3159876543',
    email: 'ana.gomez@email.com',
    createdAt: '2026-03-05T09:30:00Z',
  };

  const vehicle1: Vehicle = {
    id: 'v1000000-0000-4000-8000-000000000001',
    customerId: customer1.id,
    ownerName: customer1.fullName,
    plate: 'ABC123',
    vin: '1HGCR2F83HA000001',
    brand: 'Toyota',
    model: 'Corolla',
    modelYear: 2022,
    createdAt: '2026-03-01T08:15:00Z',
  };

  const vehicle2: Vehicle = {
    id: 'v2000000-0000-4000-8000-000000000002',
    customerId: customer2.id,
    ownerName: customer2.fullName,
    plate: 'XYZ789',
    vin: '3N1AB7AP4HY000002',
    brand: 'Chevrolet',
    model: 'Onix',
    modelYear: 2021,
    createdAt: '2026-03-05T09:45:00Z',
  };

  const tech1: Technician = {
    id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
    userId: '22222222-2222-4222-8222-222222222222',
    fullName: 'Juan Perez',
    specialty: 'Motor y transmision',
    busy: true,
    activeOrderId: 'o1000000-0000-4000-8000-000000000001',
    activeOrderNumber: 'OS-0001',
    activeVehiclePlate: 'ABC123',
  };

  const tech2: Technician = {
    id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
    userId: '33333333-3333-4333-8333-333333333333',
    fullName: 'Laura Ramirez',
    specialty: 'Frenos y suspension',
    busy: false,
    activeOrderId: '',
    activeOrderNumber: '',
    activeVehiclePlate: '',
  };

  const order1: ServiceOrder = {
    id: 'o1000000-0000-4000-8000-000000000001',
    orderNumber: 'OS-0001',
    vehicleId: vehicle1.id,
    vehiclePlate: vehicle1.plate,
    technicianName: tech1.fullName,
    reportedFailure: 'Ruido anormal en el compartimento de motor y perdida leve de potencia al acelerar.',
    status: 'IN_DIAGNOSIS',
    receivedAt: '2026-03-10T08:30:00Z',
    updatedAt: '2026-03-10T09:15:00Z',
  };

  const assignment1: Assignment = {
    id: 'a1000000-0000-4000-8000-000000000001',
    serviceOrderId: order1.id,
    technicianId: tech1.id,
    isActive: true,
    assignedAt: '2026-03-10T08:45:00Z',
  };

  const diagnostic1: Diagnostic = {
    id: 'd1000000-0000-4000-8000-000000000001',
    serviceOrderId: order1.id,
    technicianId: tech1.id,
    finding: 'Fuga en empaque de valvulas y bujias con desgaste prematuro.',
    componentToRepair: 'Empaque de tapa de valvulas y juego de bujias de iridio.',
    createdAt: '2026-03-10T09:15:00Z',
  };

  const transition1: StatusTransition = {
    id: 't1000000-0000-4000-8000-000000000001',
    fromStatus: 'RECEIVED',
    toStatus: 'IN_DIAGNOSIS',
    changedByName: 'Administrador del taller',
    changedAt: '2026-03-10T09:15:00Z',
  };

  return {
    customers: [customer1, customer2],
    vehicles: [vehicle1, vehicle2],
    technicians: [tech1, tech2],
    orders: [order1],
    assignments: [assignment1],
    diagnostics: [diagnostic1],
    interventions: [],
    warranties: [],
    transitions: [transition1],
  };
}

function loadDB(): MockDatabase {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) {
      return JSON.parse(raw) as MockDatabase;
    }
  } catch {
    // Local storage not accessible
  }
  const initial = getInitialData();
  saveDB(initial);
  return initial;
}

function saveDB(data: MockDatabase): void {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(data));
  } catch {
    // Storage full or unavailable
  }
}

function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function errorResponse(status: number, code: string, message: string): Response {
  return new Response(JSON.stringify({ code, message }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

export function setupBrowserMockApi(): void {
  if (typeof window === 'undefined') return;

  const originalFetch = window.fetch;

  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    let urlString = '';
    if (typeof input === 'string') {
      urlString = input;
    } else if (input instanceof URL) {
      urlString = input.toString();
    } else if (input && typeof (input as Request).url === 'string') {
      urlString = (input as Request).url;
    }

    // Only intercept local /api/ routes
    const isApiCall =
      urlString.startsWith('/api') ||
      urlString.includes('/api/') ||
      (urlString.startsWith(window.location.origin) && urlString.includes('/api/'));

    if (!isApiCall) {
      return originalFetch(input, init);
    }

    // If an external backend is explicitly configured and not localhost, try real fetch first
    const customApiUrl = (import.meta.env.VITE_API_URL as string | undefined)?.trim();
    if (customApiUrl && !customApiUrl.startsWith('/api') && !customApiUrl.includes('localhost')) {
      try {
        const response = await originalFetch(input, init);
        if (response.ok || response.status < 500) {
          return response;
        }
      } catch {
        // Fallback to local mock if remote server is unreachable
      }
    }

    // Parse path and query
    let path = urlString;
    if (path.startsWith(window.location.origin)) {
      path = path.slice(window.location.origin.length);
    }
    const [pathname, search = ''] = path.split('?');
    const searchParams = new URLSearchParams(search);
    const method = (init?.method || 'GET').toUpperCase();

    let body: any = null;
    if (init?.body && typeof init.body === 'string') {
      try {
        body = JSON.parse(init.body);
      } catch {
        body = null;
      }
    }

    const db = loadDB();

    // 1. Session / Auth
    if (pathname === '/api/session' && method === 'POST') {
      const username = body?.username?.trim();
      const password = body?.password?.trim();

      if (username === 'admin' && password === 'Admin2026*') {
        const session: Session = {
          token: 'token-admin-session-mock',
          expiresAt: new Date(Date.now() + 8 * 3600 * 1000).toISOString(),
          userId: '11111111-1111-4111-8111-111111111111',
          username: 'admin',
          fullName: 'Administrador del taller',
          role: 'ADMINISTRATOR',
        };
        return jsonResponse(session);
      }
      if (username === 'jperez' && password === 'JPerez2026*') {
        const session: Session = {
          token: 'token-jperez-session-mock',
          expiresAt: new Date(Date.now() + 8 * 3600 * 1000).toISOString(),
          userId: '22222222-2222-4222-8222-222222222222',
          username: 'jperez',
          fullName: 'Juan Perez',
          role: 'TECHNICIAN',
        };
        return jsonResponse(session);
      }
      if (username === 'lramirez' && password === 'LRamirez2026*') {
        const session: Session = {
          token: 'token-lramirez-session-mock',
          expiresAt: new Date(Date.now() + 8 * 3600 * 1000).toISOString(),
          userId: '33333333-3333-4333-8333-333333333333',
          username: 'lramirez',
          fullName: 'Laura Ramirez',
          role: 'TECHNICIAN',
        };
        return jsonResponse(session);
      }
      return errorResponse(401, 'unauthorized', 'Credenciales invalidas. Verifique su usuario y contrasena.');
    }

    if (pathname === '/api/session/logout' && method === 'POST') {
      return jsonResponse({ ok: true });
    }

    if (pathname === '/api/health') {
      return jsonResponse({ status: 'ok' });
    }

    // 2. Customers
    if (pathname === '/api/customer') {
      if (method === 'GET') {
        return jsonResponse(db.customers);
      }
      if (method === 'POST') {
        const { fullName, documentNumber, phone, email } = body as NewCustomer;
        if (!fullName || !documentNumber) {
          return errorResponse(400, 'bad_request', 'Nombre y documento son obligatorios.');
        }
        if (db.customers.some((c) => c.documentNumber === documentNumber)) {
          return errorResponse(409, 'conflict', 'Ya existe un cliente con ese numero de documento.');
        }
        const newCustomer: Customer = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'c-' + Date.now(),
          fullName,
          documentNumber,
          phone: phone || '',
          email: email || '',
          createdAt: new Date().toISOString(),
        };
        db.customers.unshift(newCustomer);
        saveDB(db);
        return jsonResponse(newCustomer, 201);
      }
    }

    // 3. Vehicles
    if (pathname === '/api/vehicle') {
      if (method === 'GET') {
        return jsonResponse(db.vehicles);
      }
      if (method === 'POST') {
        const { customerId, plate, vin, brand, model, modelYear } = body as NewVehicle;
        if (!customerId || !plate || !vin) {
          return errorResponse(400, 'bad_request', 'Todos los datos del vehiculo son obligatorios.');
        }
        if (db.vehicles.some((v) => v.plate.toUpperCase() === plate.toUpperCase())) {
          return errorResponse(409, 'conflict', 'Ya existe un vehiculo con esa placa.');
        }
        const owner = db.customers.find((c) => c.id === customerId);
        const newVehicle: Vehicle = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'v-' + Date.now(),
          customerId,
          ownerName: owner ? owner.fullName : 'Propietario',
          plate: plate.toUpperCase(),
          vin: vin.toUpperCase(),
          brand,
          model,
          modelYear: Number(modelYear),
          createdAt: new Date().toISOString(),
        };
        db.vehicles.unshift(newVehicle);
        saveDB(db);
        return jsonResponse(newVehicle, 201);
      }
    }

    // 4. Vehicle Timeline
    const timelineMatch = pathname.match(/^\/api\/vehicle\/([^/]+)\/timeline$/);
    if (timelineMatch && method === 'GET') {
      const vehicleId = timelineMatch[1];
      const vehicle = db.vehicles.find((v) => v.id === vehicleId);
      if (!vehicle) {
        return errorResponse(404, 'not_found', 'Vehiculo no encontrado.');
      }
      const entries: TimelineEntry[] = [];
      const vehicleOrders = db.orders.filter((o) => o.vehicleId === vehicleId);

      for (const ord of vehicleOrders) {
        entries.push({
          kind: 'ORDER',
          occurredAt: ord.receivedAt,
          title: `Apertura de orden ${ord.orderNumber}`,
          description: ord.reportedFailure,
          reference: ord.orderNumber,
        });

        const diag = db.diagnostics.find((d) => d.serviceOrderId === ord.id);
        if (diag) {
          entries.push({
            kind: 'DIAGNOSTIC',
            occurredAt: diag.createdAt,
            title: `Diagnostico tecnico (${ord.orderNumber})`,
            description: `${diag.finding} | Componente: ${diag.componentToRepair}`,
            reference: ord.orderNumber,
          });
        }

        const intervs = db.interventions.filter((i) => i.serviceOrderId === ord.id);
        for (const itv of intervs) {
          const partsStr = itv.part.map((p) => `${p.partName} (x${p.quantity})`).join(', ');
          entries.push({
            kind: 'INTERVENTION',
            occurredAt: itv.performedAt,
            title: `Intervencion: ${itv.description}`,
            description: `${itv.laborHourCount} horas de labor. ${partsStr ? 'Repuestos: ' + partsStr : ''}`,
            reference: ord.orderNumber,
          });
        }
      }

      entries.sort((a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime());

      const timeline: Timeline = { vehicle, entry: entries };
      return jsonResponse(timeline);
    }

    // 5. Technicians
    if (pathname === '/api/technician' && method === 'GET') {
      const updatedTechnicians = db.technicians.map((tech) => {
        const activeAssignment = db.assignments.find((a) => a.technicianId === tech.id && a.isActive);
        if (activeAssignment) {
          const relatedOrder = db.orders.find((o) => o.id === activeAssignment.serviceOrderId);
          return {
            ...tech,
            busy: true,
            activeOrderId: relatedOrder?.id || '',
            activeOrderNumber: relatedOrder?.orderNumber || '',
            activeVehiclePlate: relatedOrder?.vehiclePlate || '',
          };
        }
        return {
          ...tech,
          busy: false,
          activeOrderId: '',
          activeOrderNumber: '',
          activeVehiclePlate: '',
        };
      });
      return jsonResponse(updatedTechnicians);
    }

    // 6. Service Orders List & Create
    if (pathname === '/api/service-order') {
      if (method === 'GET') {
        const filterStatus = searchParams.get('status');
        let result = db.orders;
        if (filterStatus) {
          result = result.filter((o) => o.status === filterStatus);
        }
        return jsonResponse(result);
      }
      if (method === 'POST') {
        const { vehicleId, reportedFailure } = body || {};
        const vehicle = db.vehicles.find((v) => v.id === vehicleId);
        if (!vehicle) {
          return errorResponse(400, 'bad_request', 'Vehiculo no encontrado.');
        }

        const nextNum = String(db.orders.length + 1).padStart(4, '0');
        const orderNumber = `OS-${nextNum}`;
        const newOrder: ServiceOrder = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'o-' + Date.now(),
          orderNumber,
          vehicleId,
          vehiclePlate: vehicle.plate,
          technicianName: '',
          reportedFailure: reportedFailure || 'Revision general',
          status: 'RECEIVED',
          receivedAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };

        db.orders.unshift(newOrder);
        db.transitions.push({
          id: crypto.randomUUID ? crypto.randomUUID() : 't-' + Date.now(),
          fromStatus: 'RECEIVED',
          toStatus: 'RECEIVED',
          changedByName: 'Recepcion',
          changedAt: new Date().toISOString(),
        });
        saveDB(db);
        return jsonResponse(newOrder, 201);
      }
    }

    // 7. Single Service Order Details
    const orderMatch = pathname.match(/^\/api\/service-order\/([^/]+)$/);
    if (orderMatch && method === 'GET') {
      const orderId = orderMatch[1];
      const order = db.orders.find((o) => o.id === orderId);
      if (!order) {
        return errorResponse(404, 'not_found', 'Orden no encontrada.');
      }
      return jsonResponse(order);
    }

    // 8. Service Order Status Advance
    const statusMatch = pathname.match(/^\/api\/service-order\/([^/]+)\/status$/);
    if (statusMatch && method === 'POST') {
      const orderId = statusMatch[1];
      const nextStatus = body?.status as ServiceOrderStatus;
      const order = db.orders.find((o) => o.id === orderId);
      if (!order) {
        return errorResponse(404, 'not_found', 'Orden no encontrada.');
      }

      const prevStatus = order.status;
      order.status = nextStatus;
      order.updatedAt = new Date().toISOString();

      // If delivered, release active assignment
      if (nextStatus === 'DELIVERED') {
        const activeAss = db.assignments.find((a) => a.serviceOrderId === orderId && a.isActive);
        if (activeAss) {
          activeAss.isActive = false;
        }
      }

      db.transitions.push({
        id: crypto.randomUUID ? crypto.randomUUID() : 't-' + Date.now(),
        fromStatus: prevStatus,
        toStatus: nextStatus,
        changedByName: 'Administrador del taller',
        changedAt: new Date().toISOString(),
      });

      saveDB(db);
      return jsonResponse(order);
    }

    // 9. Status Transitions
    const transitionsMatch = pathname.match(/^\/api\/service-order\/([^/]+)\/transition$/);
    if (transitionsMatch && method === 'GET') {
      const orderId = transitionsMatch[1];
      const history = (db.transitions as (StatusTransition & { serviceOrderId?: string })[])
        .filter((t) => !t.serviceOrderId || t.serviceOrderId === orderId);
      return jsonResponse(history);
    }

    // 10. Assignment
    const assignmentMatch = pathname.match(/^\/api\/service-order\/([^/]+)\/assignment$/);
    if (assignmentMatch) {
      const orderId = assignmentMatch[1];
      if (method === 'GET') {
        const assignment = db.assignments.find((a) => a.serviceOrderId === orderId && a.isActive);
        if (!assignment) {
          return errorResponse(404, 'not_found', 'No hay tecnico asignado.');
        }
        return jsonResponse(assignment);
      }
      if (method === 'POST') {
        const { technicianId } = body || {};
        const tech = db.technicians.find((t) => t.id === technicianId);
        if (!tech) {
          return errorResponse(404, 'not_found', 'Tecnico no encontrado.');
        }

        // Deactivate previous active assignment on this order
        db.assignments.forEach((a) => {
          if (a.serviceOrderId === orderId) a.isActive = false;
        });

        const newAssignment: Assignment = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'a-' + Date.now(),
          serviceOrderId: orderId,
          technicianId,
          isActive: true,
          assignedAt: new Date().toISOString(),
        };
        db.assignments.push(newAssignment);

        // Update technician name on order
        const order = db.orders.find((o) => o.id === orderId);
        if (order) {
          order.technicianName = tech.fullName;
          order.updatedAt = new Date().toISOString();
        }

        saveDB(db);
        return jsonResponse(newAssignment);
      }
    }

    // 11. Diagnostic
    const diagnosticMatch = pathname.match(/^\/api\/service-order\/([^/]+)\/diagnostic$/);
    if (diagnosticMatch) {
      const orderId = diagnosticMatch[1];
      if (method === 'GET') {
        const diagnostic = db.diagnostics.find((d) => d.serviceOrderId === orderId);
        if (!diagnostic) {
          return errorResponse(404, 'not_found', 'Sin diagnostico.');
        }
        return jsonResponse(diagnostic);
      }
      if (method === 'POST') {
        const { finding, componentToRepair } = body || {};
        const assignment = db.assignments.find((a) => a.serviceOrderId === orderId && a.isActive);
        const newDiagnostic: Diagnostic = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'd-' + Date.now(),
          serviceOrderId: orderId,
          technicianId: assignment?.technicianId || 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
          finding,
          componentToRepair,
          createdAt: new Date().toISOString(),
        };
        db.diagnostics.push(newDiagnostic);
        saveDB(db);
        return jsonResponse(newDiagnostic, 201);
      }
    }

    // 12. Intervention
    const interventionMatch = pathname.match(/^\/api\/service-order\/([^/]+)\/intervention$/);
    if (interventionMatch) {
      const orderId = interventionMatch[1];
      if (method === 'GET') {
        const list = db.interventions.filter((i) => i.serviceOrderId === orderId);
        return jsonResponse(list);
      }
      if (method === 'POST') {
        const { description, laborHourCount, part } = body || {};
        const assignment = db.assignments.find((a) => a.serviceOrderId === orderId && a.isActive);
        const newIntervention: Intervention = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'i-' + Date.now(),
          serviceOrderId: orderId,
          technicianId: assignment?.technicianId || 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
          description,
          laborHourCount: Number(laborHourCount) || 1,
          performedAt: new Date().toISOString(),
          part: Array.isArray(part) ? (part as PartUsage[]) : [],
        };
        db.interventions.push(newIntervention);
        saveDB(db);
        return jsonResponse(newIntervention, 201);
      }
    }

    // 13. Warranties
    if (pathname === '/api/warranty') {
      if (method === 'GET') {
        return jsonResponse(db.warranties);
      }
      if (method === 'POST') {
        const { interventionId, kind, coverageMonthCount } = body || {};
        const months = Number(coverageMonthCount) || 3;
        const now = new Date();
        const exp = new Date(now);
        exp.setMonth(exp.getMonth() + months);

        const interv = db.interventions.find((i) => i.id === interventionId);
        const order = db.orders.find((o) => o.id === interv?.serviceOrderId);

        const newWarranty: Warranty = {
          id: crypto.randomUUID ? crypto.randomUUID() : 'w-' + Date.now(),
          interventionId,
          orderNumber: order?.orderNumber || 'OS-0001',
          vehiclePlate: order?.vehiclePlate || 'ABC123',
          kind: (kind as WarrantyKind) || 'LABOR',
          coverageMonthCount: months,
          issuedAt: now.toISOString(),
          expirationDate: exp.toISOString(),
          valid: true,
        };
        db.warranties.unshift(newWarranty);
        saveDB(db);
        return jsonResponse(newWarranty, 201);
      }
    }

    // 14. Dashboard
    if (pathname === '/api/dashboard' && method === 'GET') {
      const openOrders = db.orders.filter((o) => o.status !== 'DELIVERED');
      const allStatuses: ServiceOrderStatus[] = [
        'RECEIVED',
        'IN_DIAGNOSIS',
        'IN_REPAIR',
        'READY',
        'DELIVERED',
      ];
      const statusCount = allStatuses.map((st) => ({
        status: st,
        count: db.orders.filter((o) => o.status === st).length,
      }));

      const busyTechnicians = db.technicians
        .filter((t) => db.assignments.some((a) => a.technicianId === t.id && a.isActive))
        .map((t) => {
          const a = db.assignments.find((asg) => asg.technicianId === t.id && asg.isActive);
          const ord = db.orders.find((o) => o.id === a?.serviceOrderId);
          return {
            ...t,
            busy: true,
            activeOrderId: ord?.id || '',
            activeOrderNumber: ord?.orderNumber || '',
            activeVehiclePlate: ord?.vehiclePlate || '',
          };
        });

      const dashboard: Dashboard = {
        openOrderCount: openOrders.length,
        statusCount,
        busyTechnician: busyTechnicians,
      };
      return jsonResponse(dashboard);
    }

    return originalFetch(input, init);
  };
}
