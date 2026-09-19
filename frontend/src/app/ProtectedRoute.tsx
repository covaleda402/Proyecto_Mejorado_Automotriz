/* Route guard: without a session the user is sent back to the login screen. */
import { Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';

import { useSession } from '../shared/SessionContext';

interface ProtectedRouteProps {
  children: ReactNode;
  administratorOnly?: boolean;
}

export function ProtectedRoute({ children, administratorOnly = false }: ProtectedRouteProps) {
  const { session, isAdministrator } = useSession();
  if (!session) {
    return <Navigate to="/login" replace />;
  }
  if (administratorOnly && !isAdministrator) {
    return <Navigate to="/service-orders" replace />;
  }
  return <>{children}</>;
}
