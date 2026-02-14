import { useEffect, useState, type FormEvent } from "react";
import { useParams, useNavigate } from "react-router-dom";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Account, User } from "../types";

export default function AccountDetail() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const navigate = useNavigate();
  const [account, setAccount] = useState<Account | null>(null);
  const [owner, setOwner] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [showDeposit, setShowDeposit] = useState(false);
  const [showWithdraw, setShowWithdraw] = useState(false);
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");

  const fetchAccount = () => {
    client
      .get(`/accounts/${id}`)
      .then((res) => {
        setAccount(res.data);
        if (isAdmin && res.data.user_id) {
          client
            .get(`/users/${res.data.user_id}`)
            .then((uRes) => setOwner(uRes.data))
            .catch(() => setOwner(null));
        } else if (user) {
          setOwner({ id: user.id, username: user.username, first_name: "", last_name: "", email: "", phone: "", role: user.role, status: "active", created_at: "", updated_at: "" });
        }
      })
      .catch(() => navigate("/accounts"))
      .finally(() => setLoading(false));
  };

  useEffect(fetchAccount, [id, navigate, isAdmin]);

  const updateStatus = async (status: string) => {
    setUpdating(true);
    try {
      await client.put(`/accounts/${id}`, { status });
      fetchAccount();
    } catch {
      // ignore
    } finally {
      setUpdating(false);
    }
  };

  const handleDeposit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setUpdating(true);
    try {
      await client.post(`/accounts/${id}/deposit`, { amount });
      setShowDeposit(false);
      setAmount("");
      fetchAccount();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Deposit failed");
    } finally {
      setUpdating(false);
    }
  };

  const handleWithdraw = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setUpdating(true);
    try {
      await client.post(`/accounts/${id}/withdraw`, { amount });
      setShowWithdraw(false);
      setAmount("");
      fetchAccount();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Withdrawal failed");
    } finally {
      setUpdating(false);
    }
  };

  if (loading) return <p className="text-gray-400">Loading...</p>;
  if (!account) return <p className="text-gray-400">Account not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/accounts")}
        className="text-sm text-amber-500 hover:text-amber-400 flex items-center gap-1"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
        Back to Accounts
      </button>

      <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-white font-mono">
            {account.account_number}
          </h1>
          <StatusBadge status={account.status} />
        </div>

        {/* Balance Card */}
        <div className="bg-gradient-to-br from-amber-500/20 to-amber-600/10 border border-amber-500/30 rounded-xl p-6 mb-6">
          <p className="text-sm text-amber-400/70 mb-1">Available Balance</p>
          <p className="text-4xl font-bold text-white">
            {account.currency} {account.balance}
          </p>
        </div>

        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500">Owner</dt>
            <dd className="font-medium text-white">
              {owner
                ? owner.first_name
                  ? `${owner.username} (${owner.first_name} ${owner.last_name})`
                  : owner.username
                : `User #${account.user_id}`}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Type</dt>
            <dd className="capitalize font-medium text-white">{account.account_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Currency</dt>
            <dd className="font-medium text-white">{account.currency}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="text-gray-300">{new Date(account.created_at).toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd className="text-gray-300">{new Date(account.updated_at).toLocaleString()}</dd>
          </div>
        </dl>

        {/* Deposit / Withdraw buttons */}
        <div className="mt-6 pt-4 border-t border-gray-700/50 flex gap-2">
          <button
            onClick={() => { setShowDeposit(!showDeposit); setShowWithdraw(false); setAmount(""); setError(""); }}
            className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 transition-colors"
          >
            {showDeposit ? "Cancel" : "Deposit"}
          </button>
          <button
            onClick={() => { setShowWithdraw(!showWithdraw); setShowDeposit(false); setAmount(""); setError(""); }}
            className="bg-orange-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-orange-500 transition-colors"
          >
            {showWithdraw ? "Cancel" : "Withdraw"}
          </button>
        </div>

        {error && (
          <div className="mt-3 bg-red-500/10 border border-red-500/20 text-red-400 text-sm p-3 rounded-lg">
            {error}
          </div>
        )}

        {showDeposit && (
          <form onSubmit={handleDeposit} className="mt-3 flex gap-2 items-end">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Amount</label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white w-48 focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              />
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
            >
              Confirm Deposit
            </button>
          </form>
        )}

        {showWithdraw && (
          <form onSubmit={handleWithdraw} className="mt-3 flex gap-2 items-end">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Amount</label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white w-48 focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              />
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-orange-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-orange-500 disabled:opacity-50 transition-colors"
            >
              Confirm Withdrawal
            </button>
          </form>
        )}

        {/* Admin status controls */}
        {user?.role === "admin" && (
          <div className="mt-6 pt-4 border-t border-gray-700/50">
            <p className="text-sm font-medium text-gray-400 mb-2">Admin Actions</p>
            <div className="flex gap-2">
              {account.status !== "active" && (
                <button
                  onClick={() => updateStatus("active")}
                  disabled={updating}
                  className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
                >
                  Activate
                </button>
              )}
              {account.status !== "frozen" && (
                <button
                  onClick={() => updateStatus("frozen")}
                  disabled={updating}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-500 disabled:opacity-50 transition-colors"
                >
                  Freeze
                </button>
              )}
              {account.status !== "closed" && (
                <button
                  onClick={() => updateStatus("closed")}
                  disabled={updating}
                  className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-red-500 disabled:opacity-50 transition-colors"
                >
                  Close
                </button>
              )}
            </div>
          </div>
        )}
      </div>
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
      className={`inline-block px-3 py-1 rounded-lg border text-sm font-medium ${
        colors[status] ?? "bg-gray-500/20 text-gray-400 border-gray-500/30"
      }`}
    >
      {status}
    </span>
  );
}
