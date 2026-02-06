import { Routes, Route, Navigate } from "react-router-dom";
import Layout from "./components/Layout";
import ProtectedRoute from "./components/ProtectedRoute";
import RoleRoute from "./components/RoleRoute";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Accounts from "./pages/Accounts";
import AccountDetail from "./pages/AccountDetail";
import Cards from "./pages/Cards";
import CardDetail from "./pages/CardDetail";
import Transfers from "./pages/Transfers";
import Payments from "./pages/Payments";
import Notifications from "./pages/Notifications";
import Users from "./pages/Users";
import UserDetail from "./pages/UserDetail";
import Status from "./pages/Status";
import TransferFlow from "./pages/TransferFlow";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route path="/accounts" element={<Accounts />} />
        <Route path="/accounts/:id" element={<AccountDetail />} />
        <Route path="/cards" element={<Cards />} />
        <Route path="/cards/:id" element={<CardDetail />} />
        <Route path="/transfers" element={<Transfers />} />
        <Route path="/payments" element={<Payments />} />
        <Route path="/notifications" element={<Notifications />} />
        <Route
          path="/users"
          element={
            <RoleRoute allowedRoles={["admin"]}>
              <Users />
            </RoleRoute>
          }
        />
        <Route
          path="/users/:id"
          element={
            <RoleRoute allowedRoles={["admin"]}>
              <UserDetail />
            </RoleRoute>
          }
        />
        <Route
          path="/status"
          element={
            <RoleRoute allowedRoles={["admin"]}>
              <Status />
            </RoleRoute>
          }
        />
        <Route
          path="/transfer-flow"
          element={
            <RoleRoute allowedRoles={["admin"]}>
              <TransferFlow />
            </RoleRoute>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
