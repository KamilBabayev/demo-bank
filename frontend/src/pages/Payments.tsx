import { useEffect, useState, type FormEvent } from "react";
import client from "../api/client";
import type { Account, Payment } from "../types";

const defaultForm = {
  account_id: "",
  payment_type: "bill",
  recipient_name: "",
  recipient_account: "",
  recipient_bank: "",
  amount: "",
  currency: "USD",
  description: "",
};

export default function Payments() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ ...defaultForm });
  const [error, setError] = useState("");

  const fetchData = () => {
    Promise.all([
      client.get("/payments").catch(() => ({ data: { payments: [] } })),
      client.get("/accounts").catch(() => ({ data: { accounts: [] } })),
    ]).then(([payRes, acctRes]) => {
      setPayments(payRes.data.payments ?? []);
      const accts = acctRes.data.accounts ?? [];
      setAccounts(accts);
      setLoading(false);
    });
  };

  useEffect(fetchData, []);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const payload: Record<string, unknown> = {
        account_id: Number(form.account_id),
        payment_type: form.payment_type,
        amount: form.amount,
        currency: form.currency,
      };
      if (form.recipient_name) payload.recipient_name = form.recipient_name;
      if (form.recipient_account)
        payload.recipient_account = form.recipient_account;
      if (form.recipient_bank) payload.recipient_bank = form.recipient_bank;
      if (form.description) payload.description = form.description;

      await client.post("/payments", payload);
      setShowForm(false);
      setForm({ ...defaultForm });
      fetchData();
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } })
        .response;
      setError(resp?.data?.error ?? "Failed to create payment");
    }
  };

  const accountLabel = (a: Account) =>
    `${a.account_number} (${a.account_type} - ${a.currency} ${a.balance})`;

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Payments</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {showForm ? "Cancel" : "New Payment"}
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
                value={form.account_id}
                onChange={(e) =>
                  setForm({ ...form, account_id: e.target.value })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              >
                <option value="">Select account...</option>
                {accounts
                  .filter((a) => a.status === "active")
                  .map((a) => (
                    <option key={a.id} value={a.id}>
                      {accountLabel(a)}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Payment Type
              </label>
              <select
                value={form.payment_type}
                onChange={(e) =>
                  setForm({ ...form, payment_type: e.target.value })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              >
                <option value="bill">Bill</option>
                <option value="merchant">Merchant</option>
                <option value="external">External</option>
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
                value={form.amount}
                onChange={(e) =>
                  setForm({ ...form, amount: e.target.value })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Currency
              </label>
              <input
                type="text"
                maxLength={3}
                value={form.currency}
                onChange={(e) =>
                  setForm({ ...form, currency: e.target.value.toUpperCase() })
                }
                className="w-full border rounded px-3 py-2 text-sm"
              />
            </div>
          </div>

          {(form.payment_type === "merchant" ||
            form.payment_type === "external") && (
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Recipient Name
                </label>
                <input
                  type="text"
                  value={form.recipient_name}
                  onChange={(e) =>
                    setForm({ ...form, recipient_name: e.target.value })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Recipient Account
                </label>
                <input
                  type="text"
                  value={form.recipient_account}
                  onChange={(e) =>
                    setForm({ ...form, recipient_account: e.target.value })
                  }
                  className="w-full border rounded px-3 py-2 text-sm"
                />
              </div>
              {form.payment_type === "external" && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Recipient Bank
                  </label>
                  <input
                    type="text"
                    value={form.recipient_bank}
                    onChange={(e) =>
                      setForm({ ...form, recipient_bank: e.target.value })
                    }
                    className="w-full border rounded px-3 py-2 text-sm"
                  />
                </div>
              )}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <input
              type="text"
              value={form.description}
              onChange={(e) =>
                setForm({ ...form, description: e.target.value })
              }
              className="w-full border rounded px-3 py-2 text-sm"
            />
          </div>

          <button
            type="submit"
            className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
          >
            Submit Payment
          </button>
        </form>
      )}

      {payments.length === 0 ? (
        <p className="text-sm text-gray-500">No payments found.</p>
      ) : (
        <table className="w-full text-sm bg-white rounded shadow">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-2">Reference</th>
              <th className="text-left px-4 py-2">Type</th>
              <th className="text-left px-4 py-2">Recipient</th>
              <th className="text-left px-4 py-2">Amount</th>
              <th className="text-left px-4 py-2">Status</th>
              <th className="text-left px-4 py-2">Date</th>
            </tr>
          </thead>
          <tbody>
            {payments.map((p) => (
              <tr key={p.id} className="border-t">
                <td className="px-4 py-2 font-mono text-xs">
                  {p.reference_id.slice(0, 8)}...
                </td>
                <td className="px-4 py-2 capitalize">{p.payment_type}</td>
                <td className="px-4 py-2">{p.recipient_name ?? "-"}</td>
                <td className="px-4 py-2">
                  {p.currency} {p.amount}
                </td>
                <td className="px-4 py-2">
                  <StatusBadge status={p.status} />
                </td>
                <td className="px-4 py-2 text-gray-500">
                  {new Date(p.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
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
