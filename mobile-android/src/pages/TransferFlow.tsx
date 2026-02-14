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
        icon: "user",
        description: "User initiates transfer",
        status: "completed",
      },
      {
        id: "transfer-service",
        label: "Transfer Service",
        icon: "transfer",
        description: "Creates transfer record, publishes to Kafka",
        status: animationStep >= 1 ? "completed" : animationStep === 0 && isAnimating ? "active" : "pending",
      },
      {
        id: "kafka",
        label: "Kafka Queue",
        icon: "queue",
        description: "Message queued for processing",
        status: animationStep >= 2 ? "completed" : animationStep === 1 && isAnimating ? "active" : "pending",
      },
      {
        id: "account-service",
        label: "Account Service",
        icon: "account",
        description: isFailed ? "Transfer failed: " + (transfer.failure_reason || "Unknown error") : "Debits sender, credits receiver",
        status: isFailed ? "pending" : animationStep >= 3 ? "completed" : animationStep === 2 && isAnimating ? "active" : "pending",
      },
      {
        id: "notification",
        label: "Notification Service",
        icon: "notification",
        description: "Sends notifications to both parties",
        status: isFailed ? "pending" : animationStep >= 4 ? "completed" : animationStep === 3 && isAnimating ? "active" : "pending",
      },
      {
        id: "complete",
        label: isCompleted ? "Completed" : isFailed ? "Failed" : "Processing",
        icon: isCompleted ? "check" : isFailed ? "x" : "clock",
        description: isCompleted ? "Transfer successful" : isFailed ? "Transfer failed" : "Awaiting completion",
        status: animationStep >= 5 ? "completed" : "pending",
      },
    ];
  };

  const startAnimation = (transfer: Transfer) => {
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
    const stepDuration = 1000;

    const moneyTargets = isFailed
      ? { 2: 20, 3: 50 }
      : { 2: 20, 3: 60, 4: 100 };

    let step = 0;
    let currentMoneyPos = 0;

    animationIntervalRef.current = window.setInterval(() => {
      step++;
      setAnimationStep(step);

      const targetPos = moneyTargets[step as keyof typeof moneyTargets];
      if (targetPos !== undefined) {
        const startPos = currentMoneyPos;
        const distance = targetPos - startPos;
        const incrementCount = 25;
        const incrementDelay = stepDuration / incrementCount;
        let incrementsDone = 0;

        if (moneyIntervalRef.current) {
          clearInterval(moneyIntervalRef.current);
        }

        moneyIntervalRef.current = window.setInterval(() => {
          incrementsDone++;
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

  const getStepIcon = (icon: string, isActive: boolean) => {
    const color = isActive ? "text-white" : "text-gray-400";
    switch (icon) {
      case "user":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>;
      case "transfer":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>;
      case "queue":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>;
      case "account":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>;
      case "notification":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>;
      case "check":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>;
      case "x":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>;
      case "clock":
        return <svg className={`w-5 h-5 ${color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
      default:
        return null;
    }
  };

  if (loading) return <p className="text-gray-400">Loading...</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">Transfer Flow Visualization</h1>
        {selectedTransfer && (
          <button
            onClick={resetAnimation}
            className="text-amber-500 hover:text-amber-400 flex items-center gap-1 text-sm"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to List
          </button>
        )}
      </div>

      {!selectedTransfer ? (
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-700/50 bg-gray-800/50">
            <h2 className="font-medium text-white">Select a Transfer to Visualize</h2>
          </div>
          <div className="divide-y divide-gray-700/50">
            {transfers.length === 0 ? (
              <p className="p-6 text-gray-500 text-center">No transfers found</p>
            ) : (
              transfers.map((t) => {
                const fromAcc = getAccount(t.from_account_id);
                const toAcc = getAccount(t.to_account_id);
                return (
                  <div
                    key={t.id}
                    onClick={() => startAnimation(t)}
                    className="p-4 hover:bg-gray-800/30 cursor-pointer flex items-center justify-between transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                        t.status === "completed" ? "bg-green-500/20" :
                        t.status === "failed" ? "bg-red-500/20" : "bg-yellow-500/20"
                      }`}>
                        {t.status === "completed" ? (
                          <svg className="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                        ) : t.status === "failed" ? (
                          <svg className="w-5 h-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        ) : (
                          <svg className="w-5 h-5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        )}
                      </div>
                      <div>
                        <div className="font-medium text-white">
                          {fromAcc?.account_number || `Account #${t.from_account_id}`}
                          <span className="text-gray-500 mx-2">→</span>
                          {toAcc?.account_number || `Account #${t.to_account_id}`}
                        </div>
                        <div className="text-sm text-gray-500">
                          Ref: {t.reference_id.slice(0, 8)}... | {new Date(t.created_at).toLocaleString()}
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-bold text-lg text-white">{t.currency} {t.amount}</div>
                      <div className={`text-sm ${
                        t.status === "completed" ? "text-green-400" :
                        t.status === "failed" ? "text-red-400" : "text-yellow-400"
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
        <div className="space-y-6">
          {/* Transfer Summary */}
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4">
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm text-gray-500">Reference:</span>
                <span className="ml-2 font-mono text-white">{selectedTransfer.reference_id}</span>
              </div>
              <div className={`px-3 py-1 rounded-lg text-sm font-medium ${
                selectedTransfer.status === "completed" ? "bg-green-500/20 text-green-400 border border-green-500/30" :
                selectedTransfer.status === "failed" ? "bg-red-500/20 text-red-400 border border-red-500/30" :
                "bg-yellow-500/20 text-yellow-400 border border-yellow-500/30"
              }`}>
                {selectedTransfer.status.toUpperCase()}
              </div>
            </div>
          </div>

          {/* Flow Visualization */}
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6">
            {/* Account boxes */}
            <div className="flex justify-between mb-6">
              <div className={`w-44 p-4 border-2 rounded-xl text-center transition-all duration-500 ${
                animationStep >= 2
                  ? moneyFailed
                    ? "border-yellow-500/50 bg-yellow-500/10"
                    : "border-red-500/50 bg-red-500/10"
                  : "border-gray-700"
              }`}>
                <div className="text-2xl mb-2">🏦</div>
                <div className="font-medium text-sm text-white">
                  {getAccount(selectedTransfer.from_account_id)?.account_number || `#${selectedTransfer.from_account_id}`}
                </div>
                <div className="text-xs text-gray-500">Sender</div>
                {animationStep >= 2 && (
                  <div className={`mt-2 text-sm font-bold ${moneyFailed ? "text-yellow-400" : "text-red-400"}`}>
                    {moneyFailed ? "Refunded" : `-${selectedTransfer.currency} ${selectedTransfer.amount}`}
                  </div>
                )}
              </div>

              <div className={`w-44 p-4 border-2 rounded-xl text-center transition-all duration-500 ${
                moneyFailed
                  ? "border-gray-700 bg-gray-800/50"
                  : moneyPosition >= 100
                    ? "border-green-500/50 bg-green-500/10"
                    : "border-gray-700"
              }`}>
                <div className="text-2xl mb-2">🏦</div>
                <div className="font-medium text-sm text-white">
                  {getAccount(selectedTransfer.to_account_id)?.account_number || `#${selectedTransfer.to_account_id}`}
                </div>
                <div className="text-xs text-gray-500">Receiver</div>
                {moneyPosition >= 100 && !moneyFailed && (
                  <div className="mt-2 text-sm text-green-400 font-bold">
                    +{selectedTransfer.currency} {selectedTransfer.amount}
                  </div>
                )}
                {moneyFailed && (
                  <div className="mt-2 text-gray-500 text-xs">Not received</div>
                )}
              </div>
            </div>

            {/* Money Flow Line */}
            <div className="relative h-12 mb-6">
              <div className="absolute top-1/2 left-0 right-0 border-t-2 border-dashed border-gray-700"></div>
              <div className="absolute inset-0 flex items-center justify-between">
                {getFlowSteps(selectedTransfer).map((step, index) => {
                  const isActive = animationStep >= index;
                  const isFailed = selectedTransfer.status === "failed";
                  const failedAndPastPoint = isFailed && index > 3;

                  return (
                    <div
                      key={step.id}
                      className={`w-4 h-4 rounded-full transition-all duration-300 border-2 z-10 ${
                        failedAndPastPoint
                          ? "bg-gray-700 border-gray-600"
                          : isActive
                            ? "bg-green-500 border-green-400"
                            : "bg-gray-800 border-gray-600"
                      }`}
                    />
                  );
                })}
              </div>

              {animationStep >= 2 && (
                <div
                  className="absolute top-1/2 transition-all duration-75 ease-linear z-20"
                  style={{ left: `${moneyPosition}%`, transform: `translateX(-50%) translateY(-50%)` }}
                >
                  <div className={`px-3 py-1.5 rounded-full text-xs font-bold shadow-lg whitespace-nowrap ${
                    moneyFailed
                      ? "bg-red-500 text-white"
                      : "bg-green-500 text-white"
                  }`}>
                    {moneyFailed ? "Failed" : `${selectedTransfer.currency} ${selectedTransfer.amount}`}
                  </div>
                </div>
              )}
            </div>

            {/* Service Steps */}
            <div className="flex items-start justify-between">
              {getFlowSteps(selectedTransfer).map((step) => (
                <div key={step.id} className="flex flex-col items-center flex-1">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center transition-all duration-300 ${
                    step.status === "completed" ? "bg-green-500/20 border-2 border-green-500/50" :
                    step.status === "active" ? "bg-blue-500/20 border-2 border-blue-500/50 animate-pulse" :
                    "bg-gray-800 border-2 border-gray-700"
                  }`}>
                    {getStepIcon(step.icon, step.status !== "pending")}
                  </div>
                  <div className="mt-2 text-center">
                    <div className={`text-xs font-medium ${
                      step.status === "completed" ? "text-green-400" :
                      step.status === "active" ? "text-blue-400" :
                      "text-gray-500"
                    }`}>
                      {step.label}
                    </div>
                    <div className="text-xs text-gray-600 mt-1 max-w-20 leading-tight">
                      {step.description}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Progress Bar */}
            <div className="mt-6 relative">
              <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
                <div
                  className={`h-2 transition-all duration-500 ${
                    selectedTransfer.status === "failed" ? "bg-red-500" : "bg-green-500"
                  }`}
                  style={{ width: `${(animationStep / 5) * 100}%` }}
                ></div>
              </div>
            </div>
          </div>

          {/* Timeline */}
          <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6">
            <h3 className="font-medium text-white mb-4">Timeline</h3>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 rounded-full bg-green-500"></div>
                <span className="text-sm text-gray-400">Created:</span>
                <span className="text-sm font-medium text-white">{new Date(selectedTransfer.created_at).toLocaleString()}</span>
              </div>
              {selectedTransfer.completed_at && (
                <div className="flex items-center gap-3">
                  <div className={`w-2 h-2 rounded-full ${selectedTransfer.status === "completed" ? "bg-green-500" : "bg-red-500"}`}></div>
                  <span className="text-sm text-gray-400">{selectedTransfer.status === "completed" ? "Completed:" : "Failed:"}</span>
                  <span className="text-sm font-medium text-white">{new Date(selectedTransfer.completed_at).toLocaleString()}</span>
                </div>
              )}
              {selectedTransfer.failure_reason && (
                <div className="flex items-start gap-3 mt-2">
                  <div className="w-2 h-2 rounded-full bg-red-500 mt-1.5"></div>
                  <span className="text-sm text-gray-400">Reason:</span>
                  <span className="text-sm text-red-400">{selectedTransfer.failure_reason}</span>
                </div>
              )}
            </div>
          </div>

          {/* Replay Button */}
          {!isAnimating && (
            <div className="text-center">
              <button
                onClick={() => startAnimation(selectedTransfer)}
                className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-6 py-2 rounded-lg hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
              >
                Replay Animation
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
