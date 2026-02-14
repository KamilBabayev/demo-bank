import { useAuth } from "../context/AuthContext";
import type { ReactNode } from "react";

interface Props {
  allowedRoles: string[];
  children: ReactNode;
}

export default function RoleRoute({ allowedRoles, children }: Props) {
  const { user } = useAuth();

  if (!user || !allowedRoles.includes(user.role)) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-lg text-gray-500">Access Denied</p>
      </div>
    );
  }

  return <>{children}</>;
}
