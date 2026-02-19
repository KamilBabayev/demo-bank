import { useEffect, useState, type FormEvent } from "react";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Account, Transfer, User } from "../types";

export default function Transfers() {
  useAuth();
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [allAccounts, setAllAccounts] = useState<Account[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const pageSize = 10;
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [fromAccountId, setFromAccountId] = useState("");
  const [toAccountId, setToAccountId] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");

  const fetchData = (offset = 0) => {
    Promise.all([
      client.get(`/transfers?limit=${pageSize}&offset=${offset}`).catch(() => ({ data: { transfers: [], total: 0 } })),
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
      client.get("/accounts/directory").catch(() => ({ data: { accounts: [] } })),
      client.get("/users").catch(() => ({ data: { users: [] } })),
    ])
      .then(([txRes, acctRes, dirRes, userRes]) => {
        const txData = txRes?.data as { transfers?: Transfer[]; total?: number } | undefined;
        const acctData = acctRes?.data as { accounts?: Account[] } | undefined;
        const dirData = dirRes?.data as { accounts?: Account[] } | undefined;
        const userData = userRes?.data as { users?: User[] } | undefined;
        setTransfers(txData?.transfers ?? []);
        setTotal(txData?.total ?? 0);
        setAccounts(acctData?.accounts ?? []);
        setAllAccounts(dirData?.accounts ?? []);
        setUsers(userData?.users ?? []);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Error fetching data:", err);
        setLoading(false);
      });
  };

  useEffect(() => fetchData(page * pageSize), [page]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const from = accounts.find((a) => a.id === Number(fromAccountId));
      await client.post("/transfers", {
        from_account_id: Number(fromAccountId),
        to_account_id: Number(toAccountId),
        amount,
        currency: from?.currency ?? "USD",
      });
      setShowForm(false);
      setFromAccountId("");
      setToAccountId("");
      setAmount("");
      fetchData();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } })
        .response;
      setError(resp?.data?.error ?? "Failed to create transfer");
    }
  };

  const userMap = new Map(users.map((u) => [u.id, u]));

  const accountLabel = (a: Account) => {
    const owner = userMap.get(a.user_id);
    const ownerStr = owner ? `${owner.username} - ` : "";
    return `${ownerStr}${a.account_number} (${a.account_type} - ${a.currency} ${a.balance})`;
  };

  const accountMap = new Map(allAccounts.map((a) => [a.id, a]));

  const activeAccounts = accounts.filter((a) => a.status === "active");
  const toAccounts = allAccounts
    .filter((a) => a.status === "active")
    .filter((a) => a.id !== Number(fromAccountId));

  if (loading) return <p className="text-gray-400">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">Transfers</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-4 py-2 rounded-lg text-sm hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
        >
          {showForm ? "Cancel" : "New Transfer"}
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
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">
                From Account
              </label>
              <select
                required
                value={fromAccountId}
                onChange={(e) => setFromAccountId(e.target.value)}
                className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              >
                <option value="">Select account...</option>
                {activeAccounts.map((a) => (
                  <option key={a.id} value={a.id}>
                    {accountLabel(a)}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">
                To Account
              </label>
              <select
                required
                value={toAccountId}
                onChange={(e) => setToAccountId(e.target.value)}
                className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              >
                <option value="">Select account...</option>
                {toAccounts.map((a) => (
                  <option key={a.id} value={a.id}>
                    {accountLabel(a)}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">
                Amount
              </label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              />
            </div>
          </div>
          <button
            type="submit"
            className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 transition-colors"
          >
            Submit Transfer
          </button>
        </form>
      )}

      {transfers.length === 0 ? (
        <p className="text-sm text-gray-500">No transfers found.</p>
      ) : (
        <>
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-800/50">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Reference</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">From</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">To</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Amount</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {transfers.map((t) => {
                  const fromAcct = accountMap.get(t.from_account_id);
                  const toAcct = accountMap.get(t.to_account_id);
                  return (
                    <tr key={t.id} className="hover:bg-gray-800/30">
                      <td className="px-4 py-3 font-mono text-xs text-gray-300">
                        {t.reference_id.slice(0, 8)}...
                      </td>
                      <td className="px-4 py-3 text-gray-300">
                        {fromAcct ? fromAcct.account_number : t.from_account_id}
                      </td>
                      <td className="px-4 py-3 text-gray-300">
                        {toAcct ? toAcct.account_number : t.to_account_id}
                      </td>
                      <td className="px-4 py-3 text-white font-medium">
                        {t.currency} {t.amount}
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge status={t.status} />
                      </td>
                      <td className="px-4 py-3 text-gray-400">
                        {new Date(t.created_at).toLocaleString()}
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
    pending: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    processing: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    completed: "bg-green-500/20 text-green-400 border-green-500/30",
    failed: "bg-red-500/20 text-red-400 border-red-500/30",
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
