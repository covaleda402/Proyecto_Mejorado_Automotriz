/* Allocation panel: pick an available technician for this order. A busy
   or inactive technician is shown with badges and disabled in selection. */
import { useState } from 'react';

import { ApiError } from '../../services/api_client';
import { assignTechnician } from '../../services/service_order_service';
import { listTechnician } from '../../services/technician_service';
import { DataState, ErrorBanner, SuccessBanner } from '../../shared/DataState';
import { AccountStatusBadge, AvailabilityBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useSession, useToken } from '../../shared/SessionContext';

interface AssignmentPanelProps {
  serviceOrderId: string;
  assignedTechnicianId?: string;
  technicianIsActive?: boolean;
  onChange: () => void;
}

export function AssignmentPanel({
  serviceOrderId,
  assignedTechnicianId,
  technicianIsActive,
  onChange,
}: AssignmentPanelProps) {
  const token = useToken();
  const { isAdministrator } = useSession();
  const technician = useAsyncData(
    () => (isAdministrator ? listTechnician(token) : Promise.resolve([])),
    [token, isAdministrator],
  );
  const [technicianId, setTechnicianId] = useState('');
  const [error, setError] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [sending, setSending] = useState(false);

  if (!isAdministrator) {
    return null;
  }

  // Check if assigned technician has revoked/inactive access
  const hasInactiveAssignedTech =
    assignedTechnicianId &&
    (technicianIsActive === false ||
      technician.data?.some((t) => t.id === assignedTechnicianId && !t.isActive));

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    setConfirmation('');
    setSending(true);
    try {
      await assignTechnician(token, serviceOrderId, technicianId);
      setConfirmation('Técnico asignado exitosamente.');
      setTechnicianId('');
      technician.reload();
      onChange();
    } catch (failure) {
      setError(failure instanceof ApiError ? failure.message : 'No se pudo asignar el técnico.');
    } finally {
      setSending(false);
    }
  };

  return (
    <section className="card">
      <h3 className="card__title">Asignación de Técnico</h3>
      {hasInactiveAssignedTech ? (
        <div
          role="alert"
          style={{
            background: '#fef2f2',
            border: '1px solid #f87171',
            borderRadius: '6px',
            padding: '12px 16px',
            marginBottom: '16px',
            color: '#991b1b',
            fontSize: '0.925rem',
            lineHeight: 1.4,
          }}
        >
          <strong>⚠️ Alerta de asignación:</strong> El técnico responsable asignado a esta orden
          tiene su cuenta <em>inhabilitada / sin acceso</em>. Para garantizar la continuidad
          operativa y el registro de diagnósticos o reparaciones, reasigne la orden a un técnico activo.
        </div>
      ) : null}
      <DataState
        loading={technician.loading}
        error={technician.error}
        empty={(technician.data ?? []).length === 0}
        emptyMessage="No hay técnicos registrados."
      >
        {isAdministrator ? (
          <form onSubmit={submit} noValidate>
            <ErrorBanner message={error} />
            <SuccessBanner message={confirmation} />
            <div className="field">
              <label className="field__label" htmlFor="technicianId">
                Técnico
              </label>
              <select
                className="field__input"
                id="technicianId"
                value={technicianId}
                onChange={(event) => setTechnicianId(event.target.value)}
                required
              >
                <option value="">Seleccione un técnico</option>
                {(technician.data ?? []).map((item) => {
                  const statusNote = !item.isActive
                    ? ' (sin acceso)'
                    : item.busy
                    ? ' (ocupado)'
                    : '';
                  return (
                    <option
                      key={item.id}
                      value={item.id}
                      disabled={!item.canReceiveAssignment}
                    >
                      {item.fullName + statusNote}
                    </option>
                  );
                })}
              </select>
            </div>
            <button
              type="submit"
              className="button button--primary"
              disabled={sending || !technicianId}
            >
              {sending ? 'Asignando...' : 'Asignar técnico'}
            </button>
          </form>
        ) : (
          <p className="state-message">Solo el jefe de taller asigna técnicos.</p>
        )}
        <div className="table-scroll">
          <table className="data-table">
            <thead>
              <tr>
                <th scope="col">Técnico</th>
                <th scope="col">Especialidad</th>
                <th scope="col">Acceso</th>
                <th scope="col">Disponibilidad</th>
              </tr>
            </thead>
            <tbody>
              {(technician.data ?? []).map((item) => (
                <tr key={item.id}>
                  <td>{item.fullName}</td>
                  <td>{item.specialty}</td>
                  <td>
                    <AccountStatusBadge active={item.isActive} />
                  </td>
                  <td>
                    {item.isActive ? (
                      <AvailabilityBadge busy={item.busy} />
                    ) : (
                      <span className="badge badge--delivered">Inhabilitado</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </DataState>
    </section>
  );
}
