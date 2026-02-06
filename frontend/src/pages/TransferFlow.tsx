import { useEffect, useState, useRef } from "react";
import { useLocation } from "react-router-dom";
import client from "../api/client";
import type { Transfer, Account } from "../types";

interface FlowStep {
  id: string;
  label: string;
  icon: string;
  description: string;
  status: "pending" | "active" | "completed";
}

export default function TransferFlow() {
  const location = useLocation();
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedTransfer, setSelectedTransfer] = useState<Transfer | null>(null);
  const [loading, setLoading] = useState(true);
  const [animationStep, setAnimationStep] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [moneyPosition, setMoneyPosition] = useState(0);
  const [moneyFailed, setMoneyFailed] = useState(false);
  const animationIntervalRef = useRef<number | null>(null);
  const moneyIntervalRef = useRef<number | null>(null);

  // Reset to list view when navigating to this page via sidebar
  useEffect(() => {
    if (animationIntervalRef.current) {
      clearInterval(animationIntervalRef.current);
      animationIntervalRef.current = null;
    }
    if (moneyIntervalRef.current) {
      clearInterval(moneyIntervalRef.current);
      moneyIntervalRef.current = null;
    }
    setSelectedTransfer(null);
    setAnimationStep(0);
    setMoneyPosition(0);
    setMoneyFailed(false);
    setIsAnimating(false);
  }, [location.key]);

  useEffect(() => {
    Promise.all([
      client.get("/transfers?limit=20"),
      client.get("/accounts"),
    ]).then(([txRes, accRes]) => {
      const txData = txRes.data as { transfers?: Transfer[] };
      const accData = accRes.data as { accounts?: Account[] };
      setTransfers(txData.transfers ?? []);
      setAccounts(accData.accounts ?? []);
      setLoading(false);
    }).catch(() => setLoading(false));
  }, []);

  const getAccount = (id: number) => accounts.find((a) => a.id === id);

  const getFlowSteps = (transfer: Transfer): FlowStep[] => {
    const isFailed = transfer.status === "failed";
    const isCompleted = transfer.status === "completed";

    return [
      {
        id: "request",
        label: "Request",
        icon: "👤",
        description: "User initiates transfer",
        status: "completed",
      },
      {
        id: "transfer-service",
        label: "Transfer Service",
        icon: "💸",
        description: "Creates transfer record, publishes to Kafka",
        status: animationStep >= 1 ? "completed" : animationStep === 0 && isAnimating ? "active" : "pending",
      },
      {
        id: "kafka",
        label: "Kafka Queue",
        icon: "📨",
        description: "Message queued for processing",
        status: animationStep >= 2 ? "completed" : animationStep === 1 && isAnimating ? "active" : "pending",
      },
      {
        id: "account-service",
        label: "Account Service",
        icon: "💳",
        description: isFailed ? "Transfer failed: " + (transfer.failure_reason || "Unknown error") : "Debits sender, credits receiver",
        status: isFailed ? "pending" : animationStep >= 3 ? "completed" : animationStep === 2 && isAnimating ? "active" : "pending",
      },
      {
        id: "notification",
        label: "Notification Service",
        icon: "🔔",
        description: "Sends notifications to both parties",
        status: isFailed ? "pending" : animationStep >= 4 ? "completed" : animationStep === 3 && isAnimating ? "active" : "pending",
      },
      {
        id: "complete",
        label: isCompleted ? "Completed" : isFailed ? "Failed" : "Processing",
        icon: isCompleted ? "✅" : isFailed ? "❌" : "⏳",
        description: isCompleted ? "Transfer successful" : isFailed ? "Transfer failed" : "Awaiting completion",
        status: animationStep >= 5 ? "completed" : "pending",
      },
    ];
  };

  const startAnimation = (transfer: Transfer) => {
    // Clear any existing animations
    if (animationIntervalRef.current) {
      clearInterval(animationIntervalRef.current);
    }
    if (moneyIntervalRef.current) {
      clearInterval(moneyIntervalRef.current);
    }

    setSelectedTransfer(transfer);
    setAnimationStep(0);
    setMoneyPosition(0);
    setMoneyFailed(false);
    setIsAnimating(true);

    const maxSteps = transfer.status === "failed" ? 3 : 5;
    const isFailed = transfer.status === "failed";
    const stepDuration = 1000; // ms per step

    // Money position targets synced with service flow steps
    // 6 dots at: 0, 20, 40, 60, 80, 100
    // Step 2 (Kafka): 0 -> 20, Step 3 (Account): 20 -> 60, Step 4 (Notification): 60 -> 100
    const moneyTargets = isFailed
      ? { 2: 20, 3: 50 } // Failed: moves to 50% then stops
      : { 2: 20, 3: 60, 4: 100 }; // Success: progresses to 100%

    let step = 0;
    let currentMoneyPos = 0;

    // Animate through steps
    animationIntervalRef.current = window.setInterval(() => {
      step++;
      setAnimationStep(step);

      // Animate money for steps 2, 3, 4
      const targetPos = moneyTargets[step as keyof typeof moneyTargets];
      if (targetPos !== undefined) {
        const startPos = currentMoneyPos;
        const distance = targetPos - startPos;
        const incrementCount = 25; // Smooth animation
        const incrementDelay = stepDuration / incrementCount;
        let incrementsDone = 0;

        // Clear previous money interval
        if (moneyIntervalRef.current) {
          clearInterval(moneyIntervalRef.current);
        }

        moneyIntervalRef.current = window.setInterval(() => {
          incrementsDone++;
          // Ease-out animation
          const progress = 1 - Math.pow(1 - (incrementsDone / incrementCount), 2);
          const newPos = startPos + (distance * progress);
          currentMoneyPos = newPos;
          setMoneyPosition(newPos);

          if (incrementsDone >= incrementCount) {
            currentMoneyPos = targetPos;
            setMoneyPosition(targetPos);
            if (moneyIntervalRef.current) {
              clearInterval(moneyIntervalRef.current);
              moneyIntervalRef.current = null;
            }
          }
        }, incrementDelay);
      }

      if (step >= maxSteps) {
        if (animationIntervalRef.current) {
          clearInterval(animationIntervalRef.current);
          animationIntervalRef.current = null;
        }
        setIsAnimating(false);
        if (isFailed) {
          // Delay the failed state slightly for visual effect
          setTimeout(() => setMoneyFailed(true), 300);
        }
      }
    }, stepDuration);
  };

  const resetAnimation = () => {
    if (animationIntervalRef.current) {
      clearInterval(animationIntervalRef.current);
      animationIntervalRef.current = null;
    }
    if (moneyIntervalRef.current) {
      clearInterval(moneyIntervalRef.current);
      moneyIntervalRef.current = null;
    }
    setSelectedTransfer(null);
    setAnimationStep(0);
    setMoneyPosition(0);
    setMoneyFailed(false);
    setIsAnimating(false);
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Transfer Flow Visualization</h1>
        {selectedTransfer && (
          <button
            onClick={resetAnimation}
            className="bg-gray-600 text-white px-4 py-2 rounded text-sm hover:bg-gray-700"
          >
            ← Back to List
          </button>
        )}
      </div>

      {!selectedTransfer ? (
        /* Transfer Selection */
        <div className="bg-white rounded shadow">
          <div className="px-4 py-3 border-b bg-gray-50">
            <h2 className="font-medium">Select a Transfer to Visualize</h2>
          </div>
          <div className="divide-y">
            {transfers.length === 0 ? (
              <p className="p-4 text-gray-500">No transfers found</p>
            ) : (
              transfers.map((t) => {
                const fromAcc = getAccount(t.from_account_id);
                const toAcc = getAccount(t.to_account_id);
                return (
                  <div
                    key={t.id}
                    onClick={() => startAnimation(t)}
                    className="p-4 hover:bg-blue-50 cursor-pointer flex items-center justify-between"
                  >
                    <div className="flex items-center gap-4">
                      <div className="text-2xl">
                        {t.status === "completed" ? "✅" : t.status === "failed" ? "❌" : "⏳"}
                      </div>
                      <div>
                        <div className="font-medium">
                          {fromAcc?.account_number || `Account #${t.from_account_id}`} →{" "}
                          {toAcc?.account_number || `Account #${t.to_account_id}`}
                        </div>
                        <div className="text-sm text-gray-500">
                          Ref: {t.reference_id.slice(0, 8)}... | {new Date(t.created_at).toLocaleString()}
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-bold text-lg">{t.currency} {t.amount}</div>
                      <div className={`text-sm ${
                        t.status === "completed" ? "text-green-600" :
                        t.status === "failed" ? "text-red-600" : "text-yellow-600"
                      }`}>
                        {t.status.toUpperCase()}
                      </div>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      ) : (
        /* Flow Visualization */
        <div className="space-y-6">
          {/* Transfer Summary */}
          <div className="bg-white rounded shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm text-gray-500">Reference:</span>
                <span className="ml-2 font-mono">{selectedTransfer.reference_id}</span>
              </div>
              <div className={`px-3 py-1 rounded text-sm font-medium ${
                selectedTransfer.status === "completed" ? "bg-green-100 text-green-800" :
                selectedTransfer.status === "failed" ? "bg-red-100 text-red-800" :
                "bg-yellow-100 text-yellow-800"
              }`}>
                {selectedTransfer.status.toUpperCase()}
              </div>
            </div>
          </div>

          {/* Combined Flow Visualization */}
          <div className="bg-white rounded shadow p-6">
            {/* Account boxes at top */}
            <div className="flex justify-between mb-2">
              {/* From Account */}
              <div className={`w-40 p-3 border-2 rounded-lg text-center transition-all duration-500 ${
                animationStep >= 2
                  ? moneyFailed
                    ? "border-yellow-400 bg-yellow-50"
                    : "border-red-400 bg-red-50"
                  : "border-gray-300"
              }`}>
                <div className="text-2xl mb-1">🏦</div>
                <div className="font-medium text-sm">
                  {getAccount(selectedTransfer.from_account_id)?.account_number || `#${selectedTransfer.from_account_id}`}
                </div>
                <div className="text-xs text-gray-500">Sender</div>
                {animationStep >= 2 && (
                  <div className={`mt-1 text-sm font-bold ${moneyFailed ? "text-yellow-600" : "text-red-600 animate-pulse"}`}>
                    {moneyFailed ? "Refunded" : `-${selectedTransfer.currency} ${selectedTransfer.amount}`}
                  </div>
                )}
              </div>

              {/* To Account */}
              <div className={`w-40 p-3 border-2 rounded-lg text-center transition-all duration-500 ${
                moneyFailed
                  ? "border-gray-300 bg-gray-50"
                  : moneyPosition >= 100
                    ? "border-green-400 bg-green-50"
                    : "border-gray-300"
              }`}>
                <div className="text-2xl mb-1">🏦</div>
                <div className="font-medium text-sm">
                  {getAccount(selectedTransfer.to_account_id)?.account_number || `#${selectedTransfer.to_account_id}`}
                </div>
                <div className="text-xs text-gray-500">Receiver</div>
                {moneyPosition >= 100 && !moneyFailed && (
                  <div className="mt-1 text-sm text-green-600 font-bold animate-pulse">
                    +{selectedTransfer.currency} {selectedTransfer.amount}
                  </div>
                )}
                {moneyFailed && (
                  <div className="mt-1 text-gray-500 text-xs">
                    Not received
                  </div>
                )}
              </div>
            </div>

            {/* Connection lines from accounts to flow */}
            <div className="flex justify-between px-16 mb-2">
              <div className="w-px h-4 bg-gray-300"></div>
              <div className="w-px h-4 bg-gray-300"></div>
            </div>

            {/* Money Flow - dots aligned with service steps */}
            <div className="relative h-16 mb-2">
              {/* Dashed line */}
              <div className="absolute top-1/2 left-0 right-0 border-t-2 border-dashed border-gray-300"></div>

              {/* 6 dots matching 6 service steps */}
              <div className="absolute inset-0 flex items-center justify-between">
                {getFlowSteps(selectedTransfer).map((step, index) => {
                  const pos = (index / 5) * 100; // 0, 20, 40, 60, 80, 100
                  const isActive = animationStep >= index || (animationStep >= 2 && moneyPosition >= pos);
                  const isFailed = selectedTransfer.status === "failed";
                  const failedAndPastPoint = isFailed && index > 3; // Past account service

                  return (
                    <div
                      key={step.id}
                      className={`w-4 h-4 rounded-full transition-all duration-300 border-2 z-10 ${
                        failedAndPastPoint
                          ? "bg-gray-200 border-gray-300"
                          : isActive
                            ? "bg-green-500 border-green-600"
                            : "bg-white border-gray-300"
                      }`}
                    />
                  );
                })}
              </div>

              {/* Moving money */}
              {animationStep >= 2 && (
                <div
                  className="absolute top-1/2 transition-all duration-75 ease-linear z-20"
                  style={{ left: `${moneyPosition}%`, transform: `translateX(-50%) translateY(-50%)` }}
                >
                  <div className={`px-2 py-1 rounded-full text-xs font-bold shadow-lg transition-all duration-300 whitespace-nowrap ${
                    moneyFailed
                      ? "bg-red-500 text-white animate-pulse"
                      : "bg-green-500 text-white animate-bounce"
                  }`}>
                    {moneyFailed ? (
                      <>❌ Failed</>
                    ) : (
                      <>💵 {selectedTransfer.currency} {selectedTransfer.amount}</>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Service Flow Steps - aligned with dots above */}
            <div className="flex items-start justify-between">
              {getFlowSteps(selectedTransfer).map((step) => (
                <div key={step.id} className="flex flex-col items-center flex-1">
                  {/* Step Circle */}
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center text-xl transition-all duration-300 ${
                    step.status === "completed" ? "bg-green-100 border-2 border-green-500" :
                    step.status === "active" ? "bg-blue-100 border-2 border-blue-500 animate-pulse" :
                    "bg-gray-100 border-2 border-gray-300"
                  }`}>
                    {step.icon}
                  </div>

                  {/* Step Label */}
                  <div className="mt-2 text-center">
                    <div className={`text-xs font-medium ${
                      step.status === "completed" ? "text-green-700" :
                      step.status === "active" ? "text-blue-700" :
                      "text-gray-500"
                    }`}>
                      {step.label}
                    </div>
                    <div className="text-xs text-gray-400 mt-1 max-w-20 leading-tight">
                      {step.description}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Progress Bar */}
            <div className="mt-6 relative">
              <div className="h-2 bg-gray-200 rounded-full">
                <div
                  className={`h-2 rounded-full transition-all duration-500 ${
                    selectedTransfer.status === "failed" ? "bg-red-500" : "bg-green-500"
                  }`}
                  style={{ width: `${(animationStep / 5) * 100}%` }}
                ></div>
              </div>
            </div>
          </div>

          {/* Timeline */}
          <div className="bg-white rounded shadow p-6">
            <h3 className="font-medium mb-4">Timeline</h3>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 rounded-full bg-green-500"></div>
                <span className="text-sm text-gray-600">Created:</span>
                <span className="text-sm font-medium">{new Date(selectedTransfer.created_at).toLocaleString()}</span>
              </div>
              {selectedTransfer.completed_at && (
                <div className="flex items-center gap-3">
                  <div className={`w-2 h-2 rounded-full ${selectedTransfer.status === "completed" ? "bg-green-500" : "bg-red-500"}`}></div>
                  <span className="text-sm text-gray-600">{selectedTransfer.status === "completed" ? "Completed:" : "Failed:"}</span>
                  <span className="text-sm font-medium">{new Date(selectedTransfer.completed_at).toLocaleString()}</span>
                </div>
              )}
              {selectedTransfer.failure_reason && (
                <div className="flex items-start gap-3 mt-2">
                  <div className="w-2 h-2 rounded-full bg-red-500 mt-1.5"></div>
                  <span className="text-sm text-gray-600">Reason:</span>
                  <span className="text-sm text-red-600">{selectedTransfer.failure_reason}</span>
                </div>
              )}
            </div>
          </div>

          {/* Replay Button */}
          {!isAnimating && (
            <div className="text-center">
              <button
                onClick={() => startAnimation(selectedTransfer)}
                className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700"
              >
                🔄 Replay Animation
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
