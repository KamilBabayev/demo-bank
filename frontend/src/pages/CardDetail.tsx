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

  if (loading) return <p className="text-gray-500">Loading...</p>;
  if (!card) return <p className="text-gray-500">Card not found.</p>;

  return (
    <div className="space-y-4">
      <button
        onClick={() => navigate("/cards")}
        className="text-sm text-blue-600 hover:underline"
      >
        &larr; Back to Cards
      </button>

      <div className="bg-white rounded shadow p-6">
        <div className="flex justify-between items-start mb-4">
          <h1 className="text-2xl font-bold font-mono">{card.card_number}</h1>
          <StatusBadge status={card.status} />
        </div>

        {error && <p className="text-red-600 text-sm mb-4">{error}</p>}
        {success && <p className="text-green-600 text-sm mb-4">{success}</p>}

        <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div>
            <dt className="text-gray-500">Card Type</dt>
            <dd className="capitalize font-medium">{card.card_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Cardholder Name</dt>
            <dd className="font-medium">{card.cardholder_name}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Linked Account</dt>
            <dd>
              {account ? (
                <Link
                  to={`/accounts/${account.id}`}
                  className="text-blue-600 hover:underline font-medium"
                >
                  {account.account_number}
                </Link>
              ) : (
                `Account #${card.account_id}`
              )}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Expiration</dt>
            <dd className="font-medium">
              {String(card.expiration_month).padStart(2, "0")}/{card.expiration_year}
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd>{new Date(card.created_at).toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd>{new Date(card.updated_at).toLocaleString()}</dd>
          </div>
        </dl>

        {/* Available Balance from linked account */}
        {account && (
          <div className="mt-6 pt-4 border-t">
            <h2 className="text-lg font-semibold mb-3">Available Balance</h2>
            <div className="bg-green-50 p-4 rounded-lg">
              <div className="text-3xl font-bold text-green-700">
                {account.currency} {account.balance}
              </div>
              <div className="text-sm text-green-600 mt-1">
                From linked {account.account_type} account
              </div>
            </div>
          </div>
        )}

        <div className="mt-6 pt-4 border-t">
          <h2 className="text-lg font-semibold mb-3">Spending Limits</h2>
          <dl className="grid grid-cols-3 gap-4 text-sm">
            <div className="bg-gray-50 p-3 rounded">
              <dt className="text-gray-500">Per Transaction</dt>
              <dd className="text-lg font-bold">${card.per_transaction_limit}</dd>
            </div>
            <div className="bg-gray-50 p-3 rounded">
              <dt className="text-gray-500">Daily Limit</dt>
              <dd className="text-lg font-bold">${card.daily_limit}</dd>
              <dd className="text-xs text-gray-500">Used: ${card.daily_used}</dd>
            </div>
            <div className="bg-gray-50 p-3 rounded">
              <dt className="text-gray-500">Monthly Limit</dt>
              <dd className="text-lg font-bold">${card.monthly_limit}</dd>
              <dd className="text-xs text-gray-500">Used: ${card.monthly_used}</dd>
            </div>
          </dl>
        </div>

        {/* Action buttons */}
        <div className="mt-6 pt-4 border-t flex gap-2 flex-wrap">
          {card.status === "active" && (
            <button
              onClick={handleBlock}
              disabled={updating}
              className="bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700 disabled:opacity-50"
            >
              Block Card
            </button>
          )}
          {card.status === "blocked" && isAdmin && (
            <button
              onClick={handleUnblock}
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
            >
              Unblock Card
            </button>
          )}
          <button
            onClick={() => { setShowLimits(!showLimits); setShowPIN(false); setError(""); }}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
          >
            {showLimits ? "Cancel" : "Update Limits"}
          </button>
          <button
            onClick={() => { setShowPIN(!showPIN); setShowLimits(false); setError(""); }}
            className="bg-purple-600 text-white px-4 py-2 rounded text-sm hover:bg-purple-700"
          >
            {showPIN ? "Cancel" : "Set PIN"}
          </button>
          {isAdmin && card.status !== "cancelled" && (
            <button
              onClick={handleCancel}
              disabled={updating}
              className="bg-gray-600 text-white px-4 py-2 rounded text-sm hover:bg-gray-700 disabled:opacity-50"
            >
              Cancel Card
            </button>
          )}
        </div>

        {showLimits && (
          <form onSubmit={handleUpdateLimits} className="mt-4 p-4 bg-gray-50 rounded space-y-3">
            <h3 className="font-medium">Update Spending Limits</h3>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Per Transaction
                </label>
                <input
                  type="text"
                  value={limits.per_transaction_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, per_transaction_limit: e.target.value })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Daily Limit
                </label>
                <input
                  type="text"
                  value={limits.daily_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, daily_limit: e.target.value })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Monthly Limit
                </label>
                <input
                  type="text"
                  value={limits.monthly_limit}
                  onChange={(e) =>
                    setLimits({ ...limits, monthly_limit: e.target.value })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
            </div>
            <button
              type="submit"
              disabled={updating}
              className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
            >
              Save Limits
            </button>
          </form>
        )}

        {showPIN && (
          <form onSubmit={handleSetPIN} className="mt-4 p-4 bg-gray-50 rounded space-y-3">
            <h3 className="font-medium">Set Card PIN</h3>
            <div className="max-w-xs">
              <label className="block text-sm font-medium text-gray-700 mb-1">
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
                className="w-32 border rounded px-3 py-2 text-sm text-center tracking-widest"
              />
            </div>
            <button
              type="submit"
              disabled={updating || pin.length !== 4}
              className="bg-purple-600 text-white px-4 py-2 rounded text-sm hover:bg-purple-700 disabled:opacity-50"
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
    active: "bg-green-100 text-green-800",
    blocked: "bg-red-100 text-red-800",
    expired: "bg-yellow-100 text-yellow-800",
    cancelled: "bg-gray-100 text-gray-800",
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
