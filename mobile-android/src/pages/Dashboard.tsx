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

  if (loading) return <p className="text-gray-400">Loading...</p>;

  const totalBalance = accounts.reduce((sum, a) => sum + parseFloat(a.balance || "0"), 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">
          {user?.role === "admin" ? "Admin Dashboard" : "Dashboard"}
        </h1>
        <div className="text-right">
          <p className="text-sm text-gray-400">Total Balance</p>
          <p className="text-2xl font-bold text-amber-400">USD {totalBalance.toFixed(2)}</p>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500/20 to-blue-600/20 rounded-lg flex items-center justify-center">
              <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
              </svg>
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{accounts.length}</p>
              <p className="text-sm text-gray-400">Accounts</p>
            </div>
          </div>
        </div>
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-green-500/20 to-green-600/20 rounded-lg flex items-center justify-center">
              <svg className="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{transfers.length}</p>
              <p className="text-sm text-gray-400">Recent Transfers</p>
            </div>
          </div>
        </div>
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-purple-500/20 to-purple-600/20 rounded-lg flex items-center justify-center">
              <svg className="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{payments.length}</p>
              <p className="text-sm text-gray-400">Recent Payments</p>
            </div>
          </div>
        </div>
      </div>

      {/* Accounts summary */}
      <section>
        <h2 className="text-lg font-semibold text-white mb-3">
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
                className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4 hover:border-amber-500/50 transition-all duration-200 group"
              >
                <div className="flex items-center justify-between mb-2">
                  <p className="text-xs text-gray-500 font-mono">{a.account_number}</p>
                  <StatusBadge status={a.status} />
                </div>
                <p className="text-sm capitalize text-gray-400">
                  {a.account_type}
                </p>
                <p className="text-xl font-bold text-white mt-1 group-hover:text-amber-400 transition-colors">
                  {a.currency} {a.balance}
                </p>
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* Recent transfers */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-white">Recent Transfers</h2>
          <Link to="/transfers" className="text-sm text-amber-500 hover:text-amber-400">View all</Link>
        </div>
        {transfers.length === 0 ? (
          <p className="text-sm text-gray-500">No transfers yet.</p>
        ) : (
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-800/50">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Reference</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Amount</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {transfers.map((t) => (
                  <tr key={t.id} className="hover:bg-gray-800/30">
                    <td className="px-4 py-3 font-mono text-xs text-gray-300">
                      {t.reference_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 text-white">
                      {t.currency} {t.amount}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={t.status} />
                    </td>
                    <td className="px-4 py-3 text-gray-400">
                      {new Date(t.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Recent payments */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-white">Recent Payments</h2>
          <Link to="/payments" className="text-sm text-amber-500 hover:text-amber-400">View all</Link>
        </div>
        {payments.length === 0 ? (
          <p className="text-sm text-gray-500">No payments yet.</p>
        ) : (
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-800/50">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Reference</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Type</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Amount</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {payments.map((p) => (
                  <tr key={p.id} className="hover:bg-gray-800/30">
                    <td className="px-4 py-3 font-mono text-xs text-gray-300">
                      {p.reference_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 capitalize text-gray-300">{p.payment_type}</td>
                    <td className="px-4 py-3 text-white">
                      {p.currency} {p.amount}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={p.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-500/20 text-green-400 border-green-500/30",
    completed: "bg-green-500/20 text-green-400 border-green-500/30",
    pending: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    processing: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    failed: "bg-red-500/20 text-red-400 border-red-500/30",
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
