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

  if (loading) return <p className="text-gray-500">Loading...</p>;
  if (!account) return <p className="text-gray-500">Account not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/accounts")}
        className="text-sm text-blue-600 hover:underline"
      >
        &larr; Back to Accounts
      </button>

      <div className="bg-white rounded shadow p-6">
        <h1 className="text-2xl font-bold mb-4">
          Account {account.account_number}
        </h1>
        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500">Owner</dt>
            <dd className="font-medium">
              {owner
                ? owner.first_name
                  ? `${owner.username} (${owner.first_name} ${owner.last_name})`
                  : owner.username
                : `User #${account.user_id}`}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Type</dt>
            <dd className="capitalize font-medium">{account.account_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Status</dt>
            <dd>
              <StatusBadge status={account.status} />
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Balance</dt>
            <dd className="text-xl font-bold">
              {account.currency} {account.balance}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Currency</dt>
            <dd className="font-medium">{account.currency}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd>{new Date(account.created_at).toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd>{new Date(account.updated_at).toLocaleString()}</dd>
          </div>
        </dl>

        {/* Deposit / Withdraw buttons */}
        <div className="mt-6 pt-4 border-t flex gap-2">
          <button
            onClick={() => { setShowDeposit(!showDeposit); setShowWithdraw(false); setAmount(""); setError(""); }}
            className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
          >
            {showDeposit ? "Cancel" : "Deposit"}
          </button>
          <button
            onClick={() => { setShowWithdraw(!showWithdraw); setShowDeposit(false); setAmount(""); setError(""); }}
            className="bg-orange-600 text-white px-4 py-2 rounded text-sm hover:bg-orange-700"
          >
            {showWithdraw ? "Cancel" : "Withdraw"}
          </button>
        </div>

        {error && <p className="text-red-600 text-sm mt-2">{error}</p>}

        {showDeposit && (
          <form onSubmit={handleDeposit} className="mt-3 flex gap-2 items-end">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Amount</label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="border rounded px-3 py-2 text-sm w-48"
              />
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
            >
              Confirm Deposit
            </button>
          </form>
        )}

        {showWithdraw && (
          <form onSubmit={handleWithdraw} className="mt-3 flex gap-2 items-end">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Amount</label>
              <input
                type="text"
                required
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="border rounded px-3 py-2 text-sm w-48"
              />
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-orange-600 text-white px-4 py-2 rounded text-sm hover:bg-orange-700 disabled:opacity-50"
            >
              Confirm Withdrawal
            </button>
          </form>
        )}

        {/* Admin status controls */}
        {user?.role === "admin" && (
          <div className="mt-4 pt-4 border-t flex gap-2">
            {account.status !== "active" && (
              <button
                onClick={() => updateStatus("active")}
                disabled={updating}
                className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
              >
                Activate
              </button>
            )}
            {account.status !== "frozen" && (
              <button
                onClick={() => updateStatus("frozen")}
                disabled={updating}
                className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                Freeze
              </button>
            )}
            {account.status !== "closed" && (
              <button
                onClick={() => updateStatus("closed")}
                disabled={updating}
                className="bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700 disabled:opacity-50"
              >
                Close
              </button>
            )}
          </div>
        )}
      </div>
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
