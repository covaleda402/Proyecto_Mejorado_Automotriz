/* Route guard: without a session the user is sent back to the login screen. */
import { Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';

import { useSession } from '../shared/SessionContext';

export interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: 'ADMINISTRATOR' | 'TECHNICIAN';
}

export function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { session, isAdministrator } = useSession();
  if (!session) {
    return <Navigate to="/login" replace />;
  }
  if (requiredRole === 'ADMINISTRATOR' && !isAdministrator) {
    return <Navigate to="/dashboard" replace />;
  }
  if (requiredRole === 'TECHNICIAN' && session.role !== 'TECHNICIAN') {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
}
