import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Account, User, CreateAccountRequest } from "../types";

export default function Accounts() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateAccountRequest>({
    user_id: 0,
    account_type: "checking",
    currency: "USD",
  });
  const [error, setError] = useState("");

  const fetchData = () => {
    const requests: Promise<unknown>[] = [
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
    ];
    if (isAdmin) {
      requests.push(
        client.get("/users").catch(() => ({ data: { users: [] } }))
      );
    }
    Promise.all(requests).then((results) => {
      const acctRes = results[0] as { data: { accounts: Account[] } };
      setAccounts(acctRes.data.accounts ?? []);
      if (isAdmin && results[1]) {
        const userRes = results[1] as { data: { users: User[] } };
        setUsers(userRes.data.users ?? []);
      }
      setLoading(false);
    });
  };

  useEffect(fetchData, [isAdmin]);

  const userMap = new Map(users.map((u) => [u.id, u]));
  // For customers, map their own user info so the Owner column works
  if (!isAdmin && user) {
    userMap.set(user.id, { id: user.id, username: user.username, first_name: "", last_name: "", email: "", phone: "", role: user.role, status: "active", created_at: "", updated_at: "" });
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await client.post("/accounts", form);
      setShowForm(false);
      setForm({ user_id: 0, account_type: "checking", currency: "USD" });
      fetchData();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } })
        .response;
      setError(resp?.data?.error ?? "Failed to create account");
    }
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Accounts</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {showForm ? "Cancel" : "Create Account"}
        </button>
      </div>

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="bg-white p-4 rounded shadow space-y-3"
        >
          {error && <p className="text-red-600 text-sm">{error}</p>}
          <div className="grid grid-cols-2 gap-3">
            {isAdmin && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  User
                </label>
                <select
                  required
                  value={form.user_id || ""}
                  onChange={(e) =>
                    setForm({ ...form, user_id: Number(e.target.value) })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                >
                  <option value="">Select user...</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.username} ({u.first_name} {u.last_name} - {u.role})
                    </option>
                  ))}
                </select>
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Type
              </label>
              <select
                value={form.account_type}
                onChange={(e) =>
                  setForm({ ...form, account_type: e.target.value })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              >
                <option value="checking">Checking</option>
                <option value="savings">Savings</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Currency
              </label>
              <input
                type="text"
                maxLength={3}
                value={form.currency}
                onChange={(e) =>
                  setForm({ ...form, currency: e.target.value.toUpperCase() })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              />
            </div>
          </div>
          <button
            type="submit"
            className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
          >
            Create
          </button>
        </form>
      )}

      {accounts.length === 0 ? (
        <p className="text-sm text-gray-500">No accounts found.</p>
      ) : (
        <table className="w-full text-sm bg-white rounded shadow">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-2">Owner</th>
              <th className="text-left px-4 py-2">Account Number</th>
              <th className="text-left px-4 py-2">Type</th>
              <th className="text-left px-4 py-2">Balance</th>
              <th className="text-left px-4 py-2">Currency</th>
              <th className="text-left px-4 py-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {accounts.map((a) => {
              const owner = userMap.get(a.user_id);
              return (
                <tr key={a.id} className="border-t hover:bg-gray-50">
                  <td className="px-4 py-2 font-medium">
                    {owner ? owner.username : `User #${a.user_id}`}
                  </td>
                  <td className="px-4 py-2">
                    <Link
                      to={`/accounts/${a.id}`}
                      className="text-blue-600 hover:underline"
                    >
                      {a.account_number}
                    </Link>
                  </td>
                  <td className="px-4 py-2 capitalize">{a.account_type}</td>
                  <td className="px-4 py-2 font-medium">{a.balance}</td>
                  <td className="px-4 py-2">{a.currency}</td>
                  <td className="px-4 py-2">
                    <StatusBadge status={a.status} />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-100 text-green-800",
    frozen: "bg-blue-100 text-blue-800",
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
