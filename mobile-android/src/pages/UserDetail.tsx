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

  if (loading) return <p className="text-gray-400">Loading...</p>;
  if (!user) return <p className="text-gray-400">User not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/users")}
        className="text-sm text-amber-500 hover:text-amber-400 flex items-center gap-1"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
        Back to Users
      </button>

      <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 bg-gradient-to-br from-amber-400 to-amber-600 rounded-xl flex items-center justify-center text-black font-bold text-xl">
              {user.username.charAt(0).toUpperCase()}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">{user.username}</h1>
              <p className="text-gray-400">{user.email}</p>
            </div>
          </div>
          <StatusBadge status={user.status} />
        </div>

        {error && (
          <div className="mb-4 bg-red-500/10 border border-red-500/20 text-red-400 text-sm p-3 rounded-lg">
            {error}
          </div>
        )}
        {success && (
          <div className="mb-4 bg-green-500/10 border border-green-500/20 text-green-400 text-sm p-3 rounded-lg">
            {success}
          </div>
        )}

        {!editMode ? (
          <dl className="grid grid-cols-2 gap-x-8 gap-y-4 text-sm">
            <div>
              <dt className="text-gray-500">ID</dt>
              <dd className="font-medium text-white">{user.id}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Role</dt>
              <dd className="font-medium capitalize text-white">{user.role}</dd>
            </div>
            <div>
              <dt className="text-gray-500">First Name</dt>
              <dd className="font-medium text-white">{user.first_name}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Last Name</dt>
              <dd className="font-medium text-white">{user.last_name}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Phone</dt>
              <dd className="font-medium text-white">{user.phone || "-"}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Created</dt>
              <dd className="text-gray-300">{new Date(user.created_at).toLocaleString()}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Updated</dt>
              <dd className="text-gray-300">{new Date(user.updated_at).toLocaleString()}</dd>
            </div>
          </dl>
        ) : (
          <form onSubmit={handleUpdate} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  First Name
                </label>
                <input
                  type="text"
                  required
                  value={editForm.first_name}
                  onChange={(e) => setEditForm({ ...editForm, first_name: e.target.value })}
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Last Name
                </label>
                <input
                  type="text"
                  required
                  value={editForm.last_name}
                  onChange={(e) => setEditForm({ ...editForm, last_name: e.target.value })}
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Phone
                </label>
                <input
                  type="text"
                  value={editForm.phone}
                  onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })}
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                  placeholder="Optional"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={updating}
                className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
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
                className="px-4 py-2 text-sm text-gray-400 hover:text-gray-300"
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {/* Action Buttons */}
        <div className="mt-6 pt-4 border-t border-gray-700/50 flex flex-wrap gap-2">
          {!editMode && (
            <button
              onClick={() => setEditMode(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-500 transition-colors"
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
            className="bg-orange-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-orange-500 transition-colors"
          >
            {showPasswordReset ? "Cancel" : "Reset Password"}
          </button>
        </div>

        {/* Password Reset Form */}
        {showPasswordReset && (
          <form onSubmit={handlePasswordReset} className="mt-4 p-4 bg-gray-800/50 border border-gray-700/50 rounded-xl">
            <h3 className="text-sm font-medium text-white mb-2">Reset Password</h3>
            <div className="flex gap-2 items-end">
              <div className="flex-1">
                <input
                  type="password"
                  required
                  minLength={8}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                  placeholder="New password (min 8 characters)"
                />
              </div>
              <button
                type="submit"
                disabled={updating}
                className="bg-orange-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-orange-500 disabled:opacity-50 transition-colors"
              >
                Confirm Reset
              </button>
            </div>
          </form>
        )}

        {/* Status Controls */}
        <div className="mt-6 pt-4 border-t border-gray-700/50">
          <h3 className="text-sm font-medium text-gray-400 mb-3">Status Actions</h3>
          <div className="flex flex-wrap gap-2">
            {user.status !== "active" && (
              <button
                onClick={() => handleStatusChange("active")}
                disabled={updating}
                className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
              >
                Activate
              </button>
            )}
            {user.status !== "suspended" && (
              <button
                onClick={() => handleStatusChange("suspended")}
                disabled={updating}
                className="bg-yellow-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-yellow-500 disabled:opacity-50 transition-colors"
              >
                Suspend
              </button>
            )}
            {user.status !== "closed" && (
              <button
                onClick={() => handleStatusChange("closed")}
                disabled={updating}
                className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-red-500 disabled:opacity-50 transition-colors"
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
    active: "bg-green-500/20 text-green-400 border-green-500/30",
    suspended: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    closed: "bg-gray-500/20 text-gray-400 border-gray-500/30",
  };
  return (
    <span
      className={`inline-block px-3 py-1 rounded-lg border text-sm font-medium ${
        colors[status] ?? "bg-gray-500/20 text-gray-400 border-gray-500/30"
      }`}
    >
      {status}
    </span>
  );
}
