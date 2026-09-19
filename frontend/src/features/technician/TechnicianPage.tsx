/* Technician screen: workforce roster, onboarding and access lifecycle. */
import { useState } from 'react';

import { ApiError } from '../../services/api_client';
import {
  createTechnician,
  listTechnician,
  setTechnicianAccess,
} from '../../services/technician_service';
import { DataState, ErrorBanner, SuccessBanner } from '../../shared/DataState';
import { AccountStatusBadge, AvailabilityBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useSession, useToken } from '../../shared/SessionContext';

const EMPTY_FORM = {
  fullName: '',
  username: '',
  password: '',
  specialty: '',
};

export function TechnicianPage() {
  const token = useToken();
  const { isAdministrator } = useSession();
  const { data, loading, error, reload } = useAsyncData(() => listTechnician(token), [token]);

  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [formError, setFormError] = useState('');
  const [actionError, setActionError] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [sending, setSending] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const update = (field: keyof typeof EMPTY_FORM) => (event: React.ChangeEvent<HTMLInputElement>) =>
    setForm((previous) => ({ ...previous, [field]: event.target.value }));

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setFormError('');
    setConfirmation('');
    setSending(true);
    try {
      await createTechnician(token, form);
      setForm(EMPTY_FORM);
      setShowForm(false);
      setConfirmation('Tecnico registrado exitosamente con acceso activo.');
      reload();
    } catch (failure) {
      setFormError(
        failure instanceof ApiError ? failure.message : 'No se pudo registrar el tecnico.',
      );
    } finally {
      setSending(false);
    }
  };

  const handleToggleAccess = async (technicianId: string, currentActive: boolean, isBusy: boolean) => {
    const nextState = !currentActive;
    const actionName = nextState ? 'reactivar' : 'revocar';

    let confirmMsg = `¿Esta seguro de que desea ${actionName} el acceso a este tecnico?`;
    if (!nextState && isBusy) {
      confirmMsg += '\n\nNota: El tecnico tiene actualmente una orden en proceso. Al revocarle el acceso, su sesion se cerrara de inmediato y la orden quedara pendiente de reasignacion en el panel.';
    }

    if (!window.confirm(confirmMsg)) {
      return;
    }

    setActionError('');
    setConfirmation('');
    setActionLoadingId(technicianId);
    try {
      await setTechnicianAccess(token, technicianId, nextState);
      setConfirmation(
        nextState
          ? 'Acceso reactivado correctamente.'
          : 'Acceso revocado inmediatamente. Las sesiones previas han sido invalidadas.',
      );
      reload();
    } catch (failure) {
      setActionError(
        failure instanceof ApiError ? failure.message : 'No se pudo actualizar el acceso del tecnico.',
      );
    } finally {
      setActionLoadingId(null);
    }
  };

  return (
    <section>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h2 className="screen-title" style={{ margin: 0 }}>Tecnicos</h2>
        {isAdministrator && (
          <button
            type="button"
            className="button button--primary"
            onClick={() => setShowForm(!showForm)}
          >
            {showForm ? 'Cerrar formulario' : '+ Registrar nuevo tecnico'}
          </button>
        )}
      </div>

      <ErrorBanner message={actionError} />
      <SuccessBanner message={confirmation} />

      {showForm && isAdministrator && (
        <section className="card" style={{ marginBottom: '1.5rem' }}>
          <h3 className="card__title">Alta de nuevo tecnico</h3>
          <form onSubmit={submit} noValidate>
            <ErrorBanner message={formError} />
            <div className="form-grid">
              <div className="field">
                <label className="field__label" htmlFor="techFullName">
                  Nombre completo
                </label>
                <input
                  className="field__input"
                  id="techFullName"
                  placeholder="Ej: Carlos Mendoza"
                  value={form.fullName}
                  onChange={update('fullName')}
                  required
                />
              </div>
              <div className="field">
                <label className="field__label" htmlFor="techUsername">
                  Nombre de usuario
                </label>
                <input
                  className="field__input"
                  id="techUsername"
                  placeholder="Ej: cmendoza"
                  value={form.username}
                  onChange={update('username')}
                  required
                />
              </div>
              <div className="field">
                <label className="field__label" htmlFor="techPassword">
                  Contrasena inicial (min. 8 caracteres)
                </label>
                <input
                  className="field__input"
                  id="techPassword"
                  type="password"
                  placeholder="••••••••"
                  value={form.password}
                  onChange={update('password')}
                  required
                />
              </div>
              <div className="field">
                <label className="field__label" htmlFor="techSpecialty">
                  Especialidad tecnica
                </label>
                <input
                  className="field__input"
                  id="techSpecialty"
                  placeholder="Ej: Transmisiones y cajas"
                  value={form.specialty}
                  onChange={update('specialty')}
                  required
                />
              </div>
            </div>
            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem' }}>
              <button type="submit" className="button button--primary" disabled={sending}>
                {sending ? 'Guardando...' : 'Registrar tecnico'}
              </button>
              <button
                type="button"
                className="button button--secondary"
                onClick={() => {
                  setShowForm(false);
                  setForm(EMPTY_FORM);
                  setFormError('');
                }}
              >
                Cancelar
              </button>
            </div>
          </form>
        </section>
      )}

      <section className="card">
        <DataState
          loading={loading}
          error={error}
          empty={(data ?? []).length === 0}
          emptyMessage="No hay tecnicos registrados."
        >
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th scope="col">Tecnico</th>
                  <th scope="col">Especialidad</th>
                  <th scope="col">Acceso al sistema</th>
                  <th scope="col">Disponibilidad</th>
                  <th scope="col">Orden activa</th>
                  {isAdministrator && <th scope="col">Acciones</th>}
                </tr>
              </thead>
              <tbody>
                {(data ?? []).map((technician) => (
                  <tr key={technician.id}>
                    <td><strong>{technician.fullName}</strong></td>
                    <td>{technician.specialty}</td>
                    <td>
                      <AccountStatusBadge active={technician.isActive} />
                    </td>
                    <td>
                      <AvailabilityBadge busy={technician.busy} />
                    </td>
                    <td>{technician.activeOrderNumber || 'Sin orden'}</td>
                    {isAdministrator && (
                      <td>
                        <button
                          type="button"
                          className={`button button--small ${technician.isActive ? 'button--secondary' : 'button--primary'}`}
                          disabled={actionLoadingId === technician.id}
                          onClick={() => handleToggleAccess(technician.id, technician.isActive, technician.busy)}
                        >
                          {actionLoadingId === technician.id
                            ? 'Procesando...'
                            : technician.isActive
                            ? 'Revocar acceso'
                            : 'Reactivar acceso'}
                        </button>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </DataState>
      </section>
    </section>
  );
}
