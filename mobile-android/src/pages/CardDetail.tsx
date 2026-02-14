import { useEffect, useState, type FormEvent } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Card, Account } from "../types";

export default function CardDetail() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const navigate = useNavigate();
  const [card, setCard] = useState<Card | null>(null);
  const [account, setAccount] = useState<Account | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [showLimits, setShowLimits] = useState(false);
  const [showPIN, setShowPIN] = useState(false);
  const [limits, setLimits] = useState({
    daily_limit: "",
    monthly_limit: "",
    per_transaction_limit: "",
  });
  const [pin, setPIN] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const fetchCard = () => {
    client
      .get(`/cards/${id}`)
      .then((res) => {
        setCard(res.data);
        setLimits({
          daily_limit: res.data.daily_limit,
          monthly_limit: res.data.monthly_limit,
          per_transaction_limit: res.data.per_transaction_limit,
        });
        if (res.data.account_id) {
          client
            .get(`/accounts/${res.data.account_id}`)
            .then((aRes) => setAccount(aRes.data))
            .catch(() => setAccount(null));
        }
      })
      .catch(() => navigate("/cards"))
      .finally(() => setLoading(false));
  };

  useEffect(fetchCard, [id, navigate]);

  const handleBlock = async () => {
    setUpdating(true);
    setError("");
    try {
      await client.post(`/cards/${id}/block`);
      fetchCard();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to block card");
    } finally {
      setUpdating(false);
    }
  };

  const handleUnblock = async () => {
    setUpdating(true);
    setError("");
    try {
      await client.post(`/cards/${id}/unblock`);
      fetchCard();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to unblock card");
    } finally {
      setUpdating(false);
    }
  };

  const handleCancel = async () => {
    if (!confirm("Are you sure you want to cancel this card? This cannot be undone.")) {
      return;
    }
    setUpdating(true);
    setError("");
    try {
      await client.delete(`/cards/${id}`);
      navigate("/cards");
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to cancel card");
      setUpdating(false);
    }
  };

  const handleUpdateLimits = async (e: FormEvent) => {
    e.preventDefault();
    setUpdating(true);
    setError("");
    try {
      await client.put(`/cards/${id}`, limits);
      setShowLimits(false);
      setSuccess("Limits updated successfully");
      fetchCard();
      setTimeout(() => setSuccess(""), 3000);
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to update limits");
    } finally {
      setUpdating(false);
    }
  };

  const handleSetPIN = async (e: FormEvent) => {
    e.preventDefault();
    setUpdating(true);
    setError("");
    try {
      await client.post(`/cards/${id}/pin`, { pin });
      setShowPIN(false);
      setPIN("");
      setSuccess("PIN set successfully");
      setTimeout(() => setSuccess(""), 3000);
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to set PIN");
    } finally {
      setUpdating(false);
    }
  };

  if (loading) return <p className="text-gray-400">Loading...</p>;
  if (!card) return <p className="text-gray-400">Card not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/cards")}
        className="text-sm text-amber-500 hover:text-amber-400 flex items-center gap-1"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
        Back to Cards
      </button>

      <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6">
        <div className="flex justify-between items-start mb-6">
          <h1 className="text-2xl font-bold font-mono text-white">{card.card_number}</h1>
          <StatusBadge status={card.status} />
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

        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500">Card Type</dt>
            <dd className="capitalize font-medium text-white">{card.card_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Cardholder Name</dt>
            <dd className="font-medium text-white">{card.cardholder_name}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Linked Account</dt>
            <dd>
              {account ? (
                <Link
                  to={`/accounts/${account.id}`}
                  className="text-amber-500 hover:text-amber-400 font-medium"
                >
                  {account.account_number}
                </Link>
              ) : (
                <span className="text-gray-400">Account #{card.account_id}</span>
              )}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Expiration</dt>
            <dd className="font-medium text-white">
              {String(card.expiration_month).padStart(2, "0")}/{card.expiration_year}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="text-gray-300">{new Date(card.created_at).toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd className="text-gray-300">{new Date(card.updated_at).toLocaleString()}</dd>
          </div>
        </dl>

        {/* Available Balance from linked account */}
        {account && (
          <div className="mt-6 pt-4 border-t border-gray-700/50">
            <h2 className="text-lg font-semibold text-white mb-3">Available Balance</h2>
            <div className="bg-gradient-to-br from-green-500/20 to-green-600/10 border border-green-500/30 p-4 rounded-xl">
              <div className="text-3xl font-bold text-white">
                {account.currency} {account.balance}
              </div>
              <div className="text-sm text-green-400/70 mt-1">
                From linked {account.account_type} account
              </div>
            </div>
          </div>
        )}

        <div className="mt-6 pt-4 border-t border-gray-700/50">
          <h2 className="text-lg font-semibold text-white mb-3">Spending Limits</h2>
          <dl className="grid grid-cols-3 gap-4 text-sm">
            <div className="bg-gray-800/50 border border-gray-700/50 p-3 rounded-lg">
              <dt className="text-gray-500">Per Transaction</dt>
              <dd className="text-lg font-bold text-white">${card.per_transaction_limit}</dd>
            </div>
            <div className="bg-gray-800/50 border border-gray-700/50 p-3 rounded-lg">
              <dt className="text-gray-500">Daily Limit</dt>
              <dd className="text-lg font-bold text-white">${card.daily_limit}</dd>
              <dd className="text-xs text-gray-500">Used: ${card.daily_used}</dd>
            </div>
            <div className="bg-gray-800/50 border border-gray-700/50 p-3 rounded-lg">
              <dt className="text-gray-500">Monthly Limit</dt>
              <dd className="text-lg font-bold text-white">${card.monthly_limit}</dd>
              <dd className="text-xs text-gray-500">Used: ${card.monthly_used}</dd>
            </div>
          </dl>
        </div>

        {/* Action buttons */}
        <div className="mt-6 pt-4 border-t border-gray-700/50 flex gap-2 flex-wrap">
          {card.status === "active" && (
            <button
              onClick={handleBlock}
              disabled={updating}
              className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-red-500 disabled:opacity-50 transition-colors"
            >
              Block Card
            </button>
          )}
          {card.status === "blocked" && isAdmin && (
            <button
              onClick={handleUnblock}
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
            >
              Unblock Card
            </button>
          )}
          <button
            onClick={() => { setShowLimits(!showLimits); setShowPIN(false); setError(""); }}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-500 transition-colors"
          >
            {showLimits ? "Cancel" : "Update Limits"}
          </button>
          <button
            onClick={() => { setShowPIN(!showPIN); setShowLimits(false); setError(""); }}
            className="bg-purple-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-purple-500 transition-colors"
          >
            {showPIN ? "Cancel" : "Set PIN"}
          </button>
          {isAdmin && card.status !== "cancelled" && (
            <button
              onClick={handleCancel}
              disabled={updating}
              className="bg-gray-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-gray-500 disabled:opacity-50 transition-colors"
            >
              Cancel Card
            </button>
          )}
        </div>

        {showLimits && (
          <form onSubmit={handleUpdateLimits} className="mt-4 p-4 bg-gray-800/50 border border-gray-700/50 rounded-xl space-y-3">
            <h3 className="font-medium text-white">Update Spending Limits</h3>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Per Transaction
                </label>
                <input
                  type="text"
                  value={limits.per_transaction_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, per_transaction_limit: e.target.value })
                  }
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Daily Limit
                </label>
                <input
                  type="text"
                  value={limits.daily_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, daily_limit: e.target.value })
                  }
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Monthly Limit
                </label>
                <input
                  type="text"
                  value={limits.monthly_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, monthly_limit: e.target.value })
                  }
                  className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                />
              </div>
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-500 disabled:opacity-50 transition-colors"
            >
              Save Limits
            </button>
          </form>
        )}

        {showPIN && (
          <form onSubmit={handleSetPIN} className="mt-4 p-4 bg-gray-800/50 border border-gray-700/50 rounded-xl space-y-3">
            <h3 className="font-medium text-white">Set Card PIN</h3>
            <div className="max-w-xs">
              <label className="block text-sm font-medium text-gray-400 mb-1">
                4-Digit PIN
              </label>
              <input
                type="password"
                maxLength={4}
                pattern="[0-9]{4}"
                required
                placeholder="****"
                value={pin}
                onChange={(e) => setPIN(e.target.value.replace(/\D/g, ""))}
                className="w-32 bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-2 text-sm text-center tracking-widest text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
              />
            </div>
            <button
              type="submit"
              disabled={updating || pin.length !== 4}
              className="bg-purple-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-purple-500 disabled:opacity-50 transition-colors"
            >
              Set PIN
            </button>
          </form>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-500/20 text-green-400 border-green-500/30",
    blocked: "bg-red-500/20 text-red-400 border-red-500/30",
    expired: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    cancelled: "bg-gray-500/20 text-gray-400 border-gray-500/30",
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
