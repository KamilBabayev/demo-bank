import { useEffect, useState, type FormEvent } from "react";
import client from "../api/client";
import type { Account, Payment } from "../types";

// Predefined payment services
const paymentServices = {
  internet: [
    { id: "spectrum", name: "Spectrum Internet", icon: "wifi" },
    { id: "att", name: "AT&T Internet", icon: "wifi" },
    { id: "xfinity", name: "Xfinity/Comcast", icon: "wifi" },
    { id: "verizon-fios", name: "Verizon Fios", icon: "wifi" },
    { id: "google-fiber", name: "Google Fiber", icon: "wifi" },
  ],
  mobile: [
    { id: "verizon", name: "Verizon Wireless", icon: "phone" },
    { id: "tmobile", name: "T-Mobile", icon: "phone" },
    { id: "att-mobile", name: "AT&T Wireless", icon: "phone" },
    { id: "sprint", name: "Sprint", icon: "phone" },
    { id: "mint", name: "Mint Mobile", icon: "phone" },
  ],
};

type PaymentCategory = "internet" | "mobile" | null;

export default function Payments() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<PaymentCategory>(null);
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [accountId, setAccountId] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

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

  const handlePay = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!selectedService || !accountId || !amount) {
      setError("Please fill in all fields");
      return;
    }

    const service = [...paymentServices.internet, ...paymentServices.mobile].find(
      (s) => s.id === selectedService
    );

    try {
      await client.post("/payments", {
        account_id: Number(accountId),
        payment_type: "bill",
        recipient_name: service?.name,
        amount: amount,
        currency: "USD",
        description: `${selectedCategory === "internet" ? "Internet" : "Mobile"} bill payment`,
      });
      setSuccess(`Payment of $${amount} to ${service?.name} initiated successfully!`);
      setShowForm(false);
      setSelectedCategory(null);
      setSelectedService(null);
      setAccountId("");
      setAmount("");
      fetchData();
      setTimeout(() => setSuccess(""), 5000);
    } catch (err: unknown) {
      const resp = (err as { response?: { data?: { error?: string } } }).response;
      setError(resp?.data?.error ?? "Failed to process payment");
    }
  };

  const resetForm = () => {
    setShowForm(false);
    setSelectedCategory(null);
    setSelectedService(null);
    setAccountId("");
    setAmount("");
    setError("");
  };

  const accountLabel = (a: Account) =>
    `${a.account_number} (${a.currency} ${a.balance})`;

  if (loading) return <p className="text-gray-400">Loading...</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">Payments</h1>
        <button
          onClick={() => showForm ? resetForm() : setShowForm(true)}
          className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-4 py-2 rounded-lg text-sm hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
        >
          {showForm ? "Cancel" : "Pay Bills"}
        </button>
      </div>

      {success && (
        <div className="bg-green-500/10 border border-green-500/20 text-green-400 text-sm p-4 rounded-xl flex items-center gap-3">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {success}
        </div>
      )}

      {showForm && (
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6 space-y-6">
          {error && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm p-3 rounded-lg">
              {error}
            </div>
          )}

          {/* Step 1: Select Category */}
          {!selectedCategory && (
            <div>
              <h3 className="text-lg font-semibold text-white mb-4">Select Service Type</h3>
              <div className="grid grid-cols-2 gap-4">
                <button
                  onClick={() => setSelectedCategory("internet")}
                  className="bg-gray-800/50 border border-gray-700/50 rounded-xl p-6 hover:border-amber-500/50 transition-all duration-200 group"
                >
                  <div className="w-14 h-14 bg-gradient-to-br from-blue-500/20 to-blue-600/20 rounded-xl flex items-center justify-center mb-4 mx-auto group-hover:from-blue-500/30 group-hover:to-blue-600/30">
                    <svg className="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.14 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
                    </svg>
                  </div>
                  <h4 className="text-white font-medium text-center">Internet</h4>
                  <p className="text-gray-500 text-sm text-center mt-1">Pay internet bills</p>
                </button>
                <button
                  onClick={() => setSelectedCategory("mobile")}
                  className="bg-gray-800/50 border border-gray-700/50 rounded-xl p-6 hover:border-amber-500/50 transition-all duration-200 group"
                >
                  <div className="w-14 h-14 bg-gradient-to-br from-green-500/20 to-green-600/20 rounded-xl flex items-center justify-center mb-4 mx-auto group-hover:from-green-500/30 group-hover:to-green-600/30">
                    <svg className="w-8 h-8 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <h4 className="text-white font-medium text-center">Mobile</h4>
                  <p className="text-gray-500 text-sm text-center mt-1">Pay mobile bills</p>
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Select Provider */}
          {selectedCategory && !selectedService && (
            <div>
              <div className="flex items-center gap-2 mb-4">
                <button
                  onClick={() => setSelectedCategory(null)}
                  className="text-amber-500 hover:text-amber-400"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                <h3 className="text-lg font-semibold text-white">
                  Select {selectedCategory === "internet" ? "Internet" : "Mobile"} Provider
                </h3>
              </div>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {paymentServices[selectedCategory].map((service) => (
                  <button
                    key={service.id}
                    onClick={() => setSelectedService(service.id)}
                    className="bg-gray-800/50 border border-gray-700/50 rounded-xl p-4 hover:border-amber-500/50 transition-all duration-200 text-left group"
                  >
                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center mb-2 ${
                      selectedCategory === "internet"
                        ? "bg-blue-500/20"
                        : "bg-green-500/20"
                    }`}>
                      {selectedCategory === "internet" ? (
                        <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.14 0" />
                        </svg>
                      ) : (
                        <svg className="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                        </svg>
                      )}
                    </div>
                    <h4 className="text-white font-medium text-sm">{service.name}</h4>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Step 3: Enter Payment Details */}
          {selectedService && (
            <form onSubmit={handlePay}>
              <div className="flex items-center gap-2 mb-4">
                <button
                  type="button"
                  onClick={() => setSelectedService(null)}
                  className="text-amber-500 hover:text-amber-400"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                <h3 className="text-lg font-semibold text-white">Payment Details</h3>
              </div>

              <div className="bg-gray-800/30 border border-gray-700/30 rounded-xl p-4 mb-6">
                <div className="flex items-center gap-3">
                  <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                    selectedCategory === "internet"
                      ? "bg-blue-500/20"
                      : "bg-green-500/20"
                  }`}>
                    {selectedCategory === "internet" ? (
                      <svg className="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.14 0" />
                      </svg>
                    ) : (
                      <svg className="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                      </svg>
                    )}
                  </div>
                  <div>
                    <p className="text-white font-medium">
                      {[...paymentServices.internet, ...paymentServices.mobile].find(s => s.id === selectedService)?.name}
                    </p>
                    <p className="text-gray-500 text-sm">
                      {selectedCategory === "internet" ? "Internet Service" : "Mobile Service"}
                    </p>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-2">
                    Pay From Account
                  </label>
                  <select
                    required
                    value={accountId}
                    onChange={(e) => setAccountId(e.target.value)}
                    className="w-full bg-gray-800/50 border border-gray-700 rounded-lg px-3 py-3 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
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
                  <label className="block text-sm font-medium text-gray-400 mb-2">
                    Amount (USD)
                  </label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
                    <input
                      type="text"
                      required
                      placeholder="0.00"
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                      className="w-full bg-gray-800/50 border border-gray-700 rounded-lg pl-8 pr-3 py-3 text-sm text-white focus:outline-none focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20"
                    />
                  </div>
                </div>
              </div>

              <button
                type="submit"
                className="mt-6 w-full bg-gradient-to-r from-amber-500 to-amber-600 text-black font-semibold py-3 rounded-lg hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
              >
                Pay Now
              </button>
            </form>
          )}
        </div>
      )}

      {/* Recent Payments */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-4">Payment History</h2>
        {payments.length === 0 ? (
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-8 text-center">
            <svg className="w-12 h-12 text-gray-600 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            <p className="text-gray-500">No payments yet.</p>
          </div>
        ) : (
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-800/50">
                <tr>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Reference</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Recipient</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Amount</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
                  <th className="text-left px-4 py-3 text-gray-400 font-medium">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {payments.map((p) => (
                  <tr key={p.id} className="hover:bg-gray-800/30">
                    <td className="px-4 py-3 font-mono text-xs text-gray-300">
                      {p.reference_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 text-gray-300">{p.recipient_name ?? "-"}</td>
                    <td className="px-4 py-3 text-white font-medium">
                      {p.currency} {p.amount}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={p.status} />
                    </td>
                    <td className="px-4 py-3 text-gray-400">
                      {new Date(p.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
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
