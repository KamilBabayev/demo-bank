import { useEffect, useState, type FormEvent } from "react";
import { useParams, useNavigate } from "react-router-dom";
import client from "../api/client";
import type { User } from "../types";

export default function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  // Edit form state
  const [editMode, setEditMode] = useState(false);
  const [editForm, setEditForm] = useState({
    first_name: "",
    last_name: "",
    phone: "",
  });

  // Password reset state
  const [showPasswordReset, setShowPasswordReset] = useState(false);
  const [newPassword, setNewPassword] = useState("");

  const fetchUser = () => {
    client
      .get(`/users/${id}`)
      .then((res) => {
        setUser(res.data);
        setEditForm({
          first_name: res.data.first_name || "",
          last_name: res.data.last_name || "",
          phone: res.data.phone || "",
        });
      })
      .catch(() => navigate("/users"))
      .finally(() => setLoading(false));
  };

  useEffect(fetchUser, [id, navigate]);

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setUpdating(true);
    try {
      await client.put(`/users/${id}`, {
        first_name: editForm.first_name,
        last_name: editForm.last_name,
        phone: editForm.phone || null,
      });
      setSuccess("User updated successfully!");
      setEditMode(false);
      fetchUser();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to update user");
    } finally {
      setUpdating(false);
    }
  };

  const handlePasswordReset = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setUpdating(true);
    try {
      await client.put(`/users/${id}`, { password: newPassword });
      setSuccess("Password reset successfully!");
      setShowPasswordReset(false);
      setNewPassword("");
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to reset password");
    } finally {
      setUpdating(false);
    }
  };

  const handleStatusChange = async (status: string) => {
    setError("");
    setSuccess("");
    setUpdating(true);
    try {
      await client.put(`/users/${id}`, { status });
      setSuccess(`User ${status === "active" ? "activated" : status === "suspended" ? "suspended" : "closed"} successfully!`);
      fetchUser();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to update status");
    } finally {
      setUpdating(false);
    }
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;
  if (!user) return <p className="text-gray-500">User not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/users")}
        className="text-sm text-blue-600 hover:underline"
      >
        &larr; Back to Users
      </button>

      <div className="bg-white rounded shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">{user.username}</h1>
          <StatusBadge status={user.status} />
        </div>

        {error && <p className="text-red-600 text-sm mb-4">{error}</p>}
        {success && <p className="text-green-600 text-sm mb-4">{success}</p>}

        {!editMode ? (
          <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
            <div>
              <dt className="text-gray-500">ID</dt>
              <dd className="font-medium">{user.id}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Email</dt>
              <dd className="font-medium">{user.email}</dd>
            </div>
            <div>
              <dt className="text-gray-500">First Name</dt>
              <dd className="font-medium">{user.first_name}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Last Name</dt>
              <dd className="font-medium">{user.last_name}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Phone</dt>
              <dd className="font-medium">{user.phone || "-"}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Role</dt>
              <dd className="font-medium capitalize">{user.role}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Created</dt>
              <dd>{new Date(user.created_at).toLocaleString()}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Updated</dt>
              <dd>{new Date(user.updated_at).toLocaleString()}</dd>
            </div>
          </dl>
        ) : (
          <form onSubmit={handleUpdate} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  First Name
                </label>
                <input
                  type="text"
                  required
                  value={editForm.first_name}
                  onChange={(e) => setEditForm({ ...editForm, first_name: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Last Name
                </label>
                <input
                  type="text"
                  required
                  value={editForm.last_name}
                  onChange={(e) => setEditForm({ ...editForm, last_name: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Phone
                </label>
                <input
                  type="text"
                  value={editForm.phone}
                  onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="Optional"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={updating}
                className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
              >
                Save Changes
              </button>
              <button
                type="button"
                onClick={() => {
                  setEditMode(false);
                  setEditForm({
                    first_name: user.first_name,
                    last_name: user.last_name,
                    phone: user.phone || "",
                  });
                }}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {/* Action Buttons */}
        <div className="mt-6 pt-4 border-t flex flex-wrap gap-2">
          {!editMode && (
            <button
              onClick={() => setEditMode(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
            >
              Edit User
            </button>
          )}
          <button
            onClick={() => {
              setShowPasswordReset(!showPasswordReset);
              setNewPassword("");
              setError("");
              setSuccess("");
            }}
            className="bg-orange-600 text-white px-4 py-2 rounded text-sm hover:bg-orange-700"
          >
            {showPasswordReset ? "Cancel" : "Reset Password"}
          </button>
        </div>

        {/* Password Reset Form */}
        {showPasswordReset && (
          <form onSubmit={handlePasswordReset} className="mt-4 p-4 bg-gray-50 rounded">
            <h3 className="text-sm font-medium text-gray-700 mb-2">Reset Password</h3>
            <div className="flex gap-2 items-end">
              <div className="flex-1">
                <input
                  type="password"
                  required
                  minLength={8}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="New password (min 8 characters)"
                />
              </div>
              <button
                type="submit"
                disabled={updating}
                className="bg-orange-600 text-white px-4 py-2 rounded text-sm hover:bg-orange-700 disabled:opacity-50"
              >
                Confirm Reset
              </button>
            </div>
          </form>
        )}

        {/* Status Controls */}
        <div className="mt-4 pt-4 border-t">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Status Actions</h3>
          <div className="flex flex-wrap gap-2">
            {user.status !== "active" && (
              <button
                onClick={() => handleStatusChange("active")}
                disabled={updating}
                className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
              >
                Activate
              </button>
            )}
            {user.status !== "suspended" && (
              <button
                onClick={() => handleStatusChange("suspended")}
                disabled={updating}
                className="bg-yellow-600 text-white px-4 py-2 rounded text-sm hover:bg-yellow-700 disabled:opacity-50"
              >
                Suspend
              </button>
            )}
            {user.status !== "closed" && (
              <button
                onClick={() => handleStatusChange("closed")}
                disabled={updating}
                className="bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700 disabled:opacity-50"
              >
                Close Account
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-100 text-green-800",
    suspended: "bg-yellow-100 text-yellow-800",
    closed: "bg-gray-100 text-gray-800",
  };
  return (
    <span
      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
        colors[status] ?? "bg-gray-100 text-gray-800"
      }`}
    >
      {status}
    </span>
  );
}
