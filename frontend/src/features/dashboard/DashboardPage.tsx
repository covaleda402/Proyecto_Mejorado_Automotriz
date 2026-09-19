/* Dashboard: open orders, the count per lifecycle status and the technicians
   currently holding a vehicle. */
import { DataState } from '../../shared/DataState';
import { AvailabilityBadge, StatusBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useToken } from '../../shared/SessionContext';
import { readDashboard } from '../../services/dashboard_service';

export function DashboardPage() {
  const token = useToken();
  const { data, loading, error } = useAsyncData(() => readDashboard(token), [token]);
  const technicianList = data?.technicians ?? data?.busyTechnician ?? [];

  return (
    <section>
      <h2 className="screen-title">Panel del taller</h2>
      <DataState loading={loading} error={error} empty={!data}>
        <div className="grid">
          <article className="stat-card">
            <div className="stat-card__value">{data?.openOrderCount ?? 0}</div>
            <div className="stat-card__label">Ordenes abiertas</div>
          </article>
          {(data?.statusCount ?? []).map((item) => (
            <article className="stat-card" key={item.status}>
              <div className="stat-card__value">{item.count}</div>
              <div className="stat-card__label">
                <StatusBadge status={item.status} />
              </div>
            </article>
          ))}
        </div>
        <section className="card">
          <h3 className="card__title">Estado de tecnicos</h3>
          {technicianList.length === 0 ? (
            <p className="state-message">No hay tecnicos registrados.</p>
          ) : (
            <div className="table-scroll">
              <table className="data-table">
                <thead>
                  <tr>
                    <th scope="col">Tecnico</th>
                    <th scope="col">Estado</th>
                    <th scope="col">Orden</th>
                    <th scope="col">Placa</th>
                  </tr>
                </thead>
                <tbody>
                  {technicianList.map((technician) => (
                    <tr key={technician.id}>
                      <td>{technician.fullName}</td>
                      <td>
                        <AvailabilityBadge busy={technician.busy} />
                      </td>
                      <td>{technician.activeOrderNumber || 'Sin orden'}</td>
                      <td>{technician.activeVehiclePlate || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </DataState>
    </section>
  );
}
