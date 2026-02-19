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
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const pageSize = 10;
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateAccountRequest>({
    user_id: 0,
    account_type: "checking",
    currency: "USD",
  });
  const [error, setError] = useState("");

  const fetchData = (offset = 0) => {
    const requests: Promise<unknown>[] = [
      client.get(`/accounts?limit=${pageSize}&offset=${offset}`).catch(() => ({ data: { accounts: [], total: 0 } })),
    ];
    if (isAdmin) {
      requests.push(
        client.get("/users").catch(() => ({ data: { users: [] } }))
      );
    }
    Promise.all(requests).then((results) => {
      const acctRes = results[0] as { data: { accounts: Account[]; total: number } };
      setAccounts(acctRes.data.accounts ?? []);
      setTotal(acctRes.data.total ?? 0);
      if (isAdmin && results[1]) {
        const userRes = results[1] as { data: { users: User[] } };
        setUsers(userRes.data.users ?? []);
      }
      setLoading(false);
    });
  };

  useEffect(() => fetchData(page * pageSize), [isAdmin, page]);

  const userMap = new Map(users.map((u) => [u.id, u]));
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

  if (loading) return <p className="text-gray-400">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">Accounts</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-4 py-2 rounded-lg text-sm hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
        >
          {showForm ? "Cancel" : "Create Account"}
        </button>
      </div>

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 p-4 rounded-xl space-y-3"
        >
          {error && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm p-3 rounded-lg">
              {error}
            </div>
          )}
          <div className="grid grid-cols-2 gap-3">
            {isAdmin && (
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  User
                </label>
                <select
                  required
                  value={form.user_id || ""}
                  onChange={(e) =>
                    setForm({ ...form, user_id: Number(e.target.value) })
                  }
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
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
              <label className="block text-sm font-medium text-gray-400 mb-1">
                Type
              </label>
              <select
                value={form.account_type}
                onChange={(e) =>
                  setForm({ ...form, account_type: e.target.value })
                }
                className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              >
                <option value="checking">Checking</option>
                <option value="savings">Savings</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">
                Currency
              </label>
              <input
                type="text"
                maxLength={3}
                value={form.currency}
                onChange={(e) =>
                  setForm({ ...form, currency: e.target.value.toUpperCase() })
                }
                className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              />
            </div>
          </div>
          <button
            type="submit"
            className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 transition-colors"
          >
            Create
          </button>
        </form>
      )}

      {accounts.length === 0 ? (
        <p className="text-sm text-gray-500">No accounts found.</p>
      ) : (
        <>
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-800/50">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Owner</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Account Number</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Type</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Balance</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Currency</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {accounts.map((a) => {
                  const owner = userMap.get(a.user_id);
                  return (
                    <tr key={a.id} className="hover:bg-gray-800/30">
                      <td className="px-4 py-3 font-medium text-gray-300">
                        {owner ? owner.username : `User #${a.user_id}`}
                      </td>
                      <td className="px-4 py-3">
                        <Link
                          to={`/accounts/${a.id}`}
                          className="text-amber-500 hover:text-amber-400 font-mono"
                        >
                          {a.account_number}
                        </Link>
                      </td>
                      <td className="px-4 py-3 capitalize text-gray-300">{a.account_type}</td>
                      <td className="px-4 py-3 font-medium text-white">{a.balance}</td>
                      <td className="px-4 py-3 text-gray-300">{a.currency}</td>
                      <td className="px-4 py-3">
                        <StatusBadge status={a.status} />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          {total > pageSize && (
            <div className="flex items-center justify-between mt-4">
              <span className="text-sm text-gray-400">
                Showing {page * pageSize + 1}–{Math.min((page + 1) * pageSize, total)} of {total}
              </span>
              <div className="flex gap-2">
                <button
                  disabled={page === 0}
                  onClick={() => setPage(page - 1)}
                  className="px-3 py-1.5 text-sm rounded-lg border border-gray-700/50 text-gray-300 hover:bg-gray-800/50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Previous
                </button>
                <button
                  disabled={(page + 1) * pageSize >= total}
                  onClick={() => setPage(page + 1)}
                  className="px-3 py-1.5 text-sm rounded-lg border border-gray-700/50 text-gray-300 hover:bg-gray-800/50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-500/20 text-green-400 border-green-500/30",
    frozen: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    closed: "bg-gray-500/20 text-gray-400 border-gray-500/30",
  };
  return (
    <span
      className={`inline-block px-2 py-0.5 rounded border text-xs font-medium ${
        colors[status] ?? "bg-gray-500/20 text-gray-400 border-gray-500/30"
      }`}
    >
      {status}
    </span>
  );
}
