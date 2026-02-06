import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Account, Transfer, Payment } from "../types";

export default function Dashboard() {
  const { user } = useAuth();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
      client.get("/transfers").catch(() => ({ data: { transfers: [] } })),
      client.get("/payments").catch(() => ({ data: { payments: [] } })),
    ]).then(([acctRes, txRes, payRes]) => {
      setAccounts(acctRes.data.accounts ?? []);
      setTransfers((txRes.data.transfers ?? []).slice(0, 5));
      setPayments((payRes.data.payments ?? []).slice(0, 5));
      setLoading(false);
    });
  }, []);

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">
        {user?.role === "admin" ? "Admin Dashboard" : "Dashboard"}
      </h1>

      {/* Accounts summary */}
      <section>
        <h2 className="text-lg font-semibold mb-3">
          {user?.role === "admin"
            ? `All Accounts (${accounts.length})`
            : "My Accounts"}
        </h2>
        {accounts.length === 0 ? (
          <p className="text-sm text-gray-500">No accounts found.</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {accounts.map((a) => (
              <Link
                key={a.id}
                to={`/accounts/${a.id}`}
                className="bg-white rounded shadow p-4 hover:shadow-md transition"
              >
                <p className="text-xs text-gray-400">{a.account_number}</p>
                <p className="text-sm capitalize text-gray-600">
                  {a.account_type}
                </p>
                <p className="text-xl font-bold mt-1">
                  {a.currency} {a.balance}
                </p>
                <StatusBadge status={a.status} />
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* Recent transfers */}
      <section>
        <h2 className="text-lg font-semibold mb-3">Recent Transfers</h2>
        {transfers.length === 0 ? (
          <p className="text-sm text-gray-500">No transfers yet.</p>
        ) : (
          <table className="w-full text-sm bg-white rounded shadow">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-left px-4 py-2">Reference</th>
                <th className="text-left px-4 py-2">Amount</th>
                <th className="text-left px-4 py-2">Status</th>
                <th className="text-left px-4 py-2">Date</th>
              </tr>
            </thead>
            <tbody>
              {transfers.map((t) => (
                <tr key={t.id} className="border-t">
                  <td className="px-4 py-2 font-mono text-xs">
                    {t.reference_id.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-2">
                    {t.currency} {t.amount}
                  </td>
                  <td className="px-4 py-2">
                    <StatusBadge status={t.status} />
                  </td>
                  <td className="px-4 py-2 text-gray-500">
                    {new Date(t.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* Recent payments */}
      <section>
        <h2 className="text-lg font-semibold mb-3">Recent Payments</h2>
        {payments.length === 0 ? (
          <p className="text-sm text-gray-500">No payments yet.</p>
        ) : (
          <table className="w-full text-sm bg-white rounded shadow">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-left px-4 py-2">Reference</th>
                <th className="text-left px-4 py-2">Type</th>
                <th className="text-left px-4 py-2">Amount</th>
                <th className="text-left px-4 py-2">Status</th>
              </tr>
            </thead>
            <tbody>
              {payments.map((p) => (
                <tr key={p.id} className="border-t">
                  <td className="px-4 py-2 font-mono text-xs">
                    {p.reference_id.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-2 capitalize">{p.payment_type}</td>
                  <td className="px-4 py-2">
                    {p.currency} {p.amount}
                  </td>
                  <td className="px-4 py-2">
                    <StatusBadge status={p.status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-100 text-green-800",
    completed: "bg-green-100 text-green-800",
    pending: "bg-yellow-100 text-yellow-800",
    processing: "bg-blue-100 text-blue-800",
    failed: "bg-red-100 text-red-800",
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
