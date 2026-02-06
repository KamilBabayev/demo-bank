import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import client from "../api/client";
import { useAuth } from "../context/AuthContext";
import type { Card, Account, User, CreateCardRequest } from "../types";

export default function Cards() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const [cards, setCards] = useState<Card[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateCardRequest>({
    account_id: 0,
    card_type: "debit",
    cardholder_name: "",
  });
  const [error, setError] = useState("");

  const fetchData = () => {
    const requests: Promise<unknown>[] = [
      client.get("/cards").catch(() => ({ data: { cards: [] } })),
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
    ];
    if (isAdmin) {
      requests.push(client.get("/users").catch(() => ({ data: { users: [] } })));
    }
    Promise.all(requests).then((results) => {
      const cardRes = results[0] as { data: { cards: Card[] } };
      const acctRes = results[1] as { data: { accounts: Account[] } };
      setCards(cardRes.data.cards ?? []);
      setAccounts(acctRes.data.accounts ?? []);
      if (isAdmin && results[2]) {
        const userRes = results[2] as { data: { users: User[] } };
        setUsers(userRes.data.users ?? []);
      }
      setLoading(false);
    });
  };

  useEffect(fetchData, [isAdmin]);

  const accountMap = new Map(accounts.map((a) => [a.id, a]));
  const userMap = new Map(users.map((u) => [u.id, u]));

  // Get display name for account owner
  const getAccountOwnerName = (account: Account): string => {
    const owner = userMap.get(account.user_id);
    if (owner) {
      return owner.first_name && owner.last_name
        ? `${owner.first_name} ${owner.last_name}`
        : owner.username;
    }
    // For non-admin, use current user info
    if (user && account.user_id === user.id) {
      return user.username;
    }
    return `User #${account.user_id}`;
  };

  // Handle account selection - auto-fill cardholder name
  const handleAccountChange = (accountId: number) => {
    const account = accounts.find((a) => a.id === accountId);
    let cardholderName = form.cardholder_name;

    if (account) {
      const owner = userMap.get(account.user_id);
      if (owner && owner.first_name && owner.last_name) {
        cardholderName = `${owner.first_name} ${owner.last_name}`.toUpperCase();
      } else if (owner) {
        cardholderName = owner.username.toUpperCase();
      } else if (user && account.user_id === user.id) {
        cardholderName = user.username.toUpperCase();
      }
    }

    setForm({ ...form, account_id: accountId, cardholder_name: cardholderName });
  };

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await client.post("/cards", form);
      setShowForm(false);
      setForm({ account_id: 0, card_type: "debit", cardholder_name: "" });
      fetchData();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } })
        .response;
      setError(resp?.data?.error ?? "Failed to create card");
    }
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Cards</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {showForm ? "Cancel" : "Create Card"}
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
                Account
              </label>
              <select
                required
                value={form.account_id || ""}
                onChange={(e) => handleAccountChange(Number(e.target.value))}
                className="w-full border rounded px-3 py-2 text-sm"
              >
                <option value="">Select account...</option>
                {accounts.map((a) => (
                  <option key={a.id} value={a.id}>
                    {getAccountOwnerName(a)} - {a.account_number} ({a.account_type})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Card Type
              </label>
              <select
                value={form.card_type}
                onChange={(e) =>
                  setForm({ ...form, card_type: e.target.value })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              >
                <option value="debit">Debit</option>
                <option value="credit">Credit</option>
                <option value="virtual">Virtual</option>
              </select>
            </div>
            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Cardholder Name
              </label>
              <input
                type="text"
                required
                placeholder="JOHN DOE"
                value={form.cardholder_name}
                onChange={(e) =>
                  setForm({ ...form, cardholder_name: e.target.value.toUpperCase() })
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

      {cards.length === 0 ? (
        <p className="text-sm text-gray-500">No cards found.</p>
      ) : (
        <table className="w-full text-sm bg-white rounded shadow">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-2">Card Number</th>
              <th className="text-left px-4 py-2">Type</th>
              <th className="text-left px-4 py-2">Cardholder</th>
              <th className="text-left px-4 py-2">Account</th>
              <th className="text-right px-4 py-2">Available Balance</th>
              <th className="text-left px-4 py-2">Expires</th>
              <th className="text-left px-4 py-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {cards.map((c) => {
              const account = accountMap.get(c.account_id);
              return (
                <tr key={c.id} className="border-t hover:bg-gray-50">
                  <td className="px-4 py-2 font-mono">
                    <Link
                      to={`/cards/${c.id}`}
                      className="text-blue-600 hover:underline"
                    >
                      {c.card_number}
                    </Link>
                  </td>
                  <td className="px-4 py-2 capitalize">{c.card_type}</td>
                  <td className="px-4 py-2">{c.cardholder_name}</td>
                  <td className="px-4 py-2">
                    {account ? (
                      <Link
                        to={`/accounts/${account.id}`}
                        className="text-blue-600 hover:underline"
                      >
                        {account.account_number}
                      </Link>
                    ) : (
                      `Account #${c.account_id}`
                    )}
                  </td>
                  <td className="px-4 py-2 text-right font-medium">
                    {account ? `${account.currency} ${account.balance}` : "-"}
                  </td>
                  <td className="px-4 py-2">
                    {String(c.expiration_month).padStart(2, "0")}/{c.expiration_year}
                  </td>
                  <td className="px-4 py-2">
                    <StatusBadge status={c.status} />
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
