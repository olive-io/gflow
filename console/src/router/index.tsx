import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../layouts/MainLayout';
import Login from '../pages/Login';
import Dashboard from '../pages/Dashboard';
import Definitions from '../pages/Definitions';
import DefinitionDetail from '../pages/DefinitionDetail';
import Processes from '../pages/Processes';
import ProcessDetail from '../pages/ProcessDetail';
import Runners from '../pages/Runners';
import Endpoints from '../pages/Endpoints';
import Users from '../pages/Users';
import AuditLogs from '../pages/AuditLogs';
import Designer from '../pages/Designer';
import { useUserStore } from '../stores';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const token = useUserStore((state) => state.token);
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
};

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/designer/new',
    element: (
      <ProtectedRoute>
        <Designer />
      </ProtectedRoute>
    ),
  },
  {
    path: '/designer/:uid',
    element: (
      <ProtectedRoute>
        <Designer />
      </ProtectedRoute>
    ),
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: <Dashboard />,
      },
      {
        path: 'definitions',
        element: <Definitions />,
      },
      {
        path: 'definitions/:uid',
        element: <DefinitionDetail />,
      },
      {
        path: 'processes',
        element: <Processes />,
      },
      {
        path: 'processes/:id',
        element: <ProcessDetail />,
      },
      {
        path: 'runners',
        element: <Runners />,
      },
      {
        path: 'endpoints',
        element: <Endpoints />,
      },
      {
        path: 'users',
        element: <Users />,
      },
      {
        path: 'audit-logs',
        element: <AuditLogs />,
      },
    ],
  },
]);
