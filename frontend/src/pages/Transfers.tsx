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
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [fromAccountId, setFromAccountId] = useState("");
  const [toAccountId, setToAccountId] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");

  const fetchData = () => {
    Promise.all([
      client.get("/transfers").catch(() => ({ data: { transfers: [] } })),
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
      client.get("/accounts/directory").catch(() => ({ data: { accounts: [] } })),
      client.get("/users").catch(() => ({ data: { users: [] } })),
    ])
      .then(([txRes, acctRes, dirRes, userRes]) => {
        const txData = txRes?.data as { transfers?: Transfer[] } | undefined;
        const acctData = acctRes?.data as { accounts?: Account[] } | undefined;
        const dirData = dirRes?.data as { accounts?: Account[] } | undefined;
        const userData = userRes?.data as { users?: User[] } | undefined;
        setTransfers(txData?.transfers ?? []);
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

  useEffect(fetchData, []);

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

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Transfers</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {showForm ? "Cancel" : "New Transfer"}
        </button>
      </div>

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="bg-white p-4 rounded shadow space-y-3"
        >
          {error && <p className="text-red-600 text-sm">{error}</p>}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                From Account
              </label>
              <select
                required
                value={fromAccountId}
                onChange={(e) => setFromAccountId(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm"
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
              <label className="block text-sm font-medium text-gray-700 mb-1">
                To Account
              </label>
              <select
                required
                value={toAccountId}
                onChange={(e) => setToAccountId(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm"
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
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount
              </label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm"
              />
            </div>
          </div>
          <button
            type="submit"
            className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
          >
            Submit Transfer
          </button>
        </form>
      )}

      {transfers.length === 0 ? (
        <p className="text-sm text-gray-500">No transfers found.</p>
      ) : (
        <table className="w-full text-sm bg-white rounded shadow">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-2">Reference</th>
              <th className="text-left px-4 py-2">From</th>
              <th className="text-left px-4 py-2">To</th>
              <th className="text-left px-4 py-2">Amount</th>
              <th className="text-left px-4 py-2">Status</th>
              <th className="text-left px-4 py-2">Date</th>
            </tr>
          </thead>
          <tbody>
            {transfers.map((t) => {
              const fromAcct = accountMap.get(t.from_account_id);
              const toAcct = accountMap.get(t.to_account_id);
              return (
                <tr key={t.id} className="border-t">
                  <td className="px-4 py-2 font-mono text-xs">
                    {t.reference_id.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-2">
                    {fromAcct ? fromAcct.account_number : t.from_account_id}
                  </td>
                  <td className="px-4 py-2">
                    {toAcct ? toAcct.account_number : t.to_account_id}
                  </td>
                  <td className="px-4 py-2">
                    {t.currency} {t.amount}
                  </td>
                  <td className="px-4 py-2">
                    <StatusBadge status={t.status} />
                  </td>
                  <td className="px-4 py-2 text-gray-500">
                    {new Date(t.created_at).toLocaleString()}
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
    pending: "bg-yellow-100 text-yellow-800",
    processing: "bg-blue-100 text-blue-800",
    completed: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
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
