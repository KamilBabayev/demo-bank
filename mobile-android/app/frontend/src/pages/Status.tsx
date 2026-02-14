import { useEffect, useState, useCallback } from "react";
import client from "../api/client";

interface ServiceStatus {
  name: string;
  status: string;
  response_time_ms: number;
  error?: string;
}

interface ConnectionStatus {
  from: string;
  to: string;
  status: string;
  response_time_ms: number;
  error?: string;
}

interface SystemStatus {
  services: ServiceStatus[];
  connections: ConnectionStatus[];
  timestamp: string;
}

// Node positions for the diagram
const nodePositions: Record<string, { x: number; y: number }> = {
  frontend: { x: 80, y: 300 },
  "api-gateway": { x: 240, y: 300 },
  user: { x: 420, y: 50 },
  account: { x: 420, y: 125 },
  card: { x: 420, y: 200 },
  transfer: { x: 420, y: 275 },
  payment: { x: 420, y: 350 },
  notification: { x: 420, y: 425 },
  "user-db": { x: 600, y: 50 },
  "account-db": { x: 600, y: 125 },
  "card-db": { x: 600, y: 200 },
  "transfer-db": { x: 600, y: 275 },
  "payment-db": { x: 600, y: 350 },
  "notification-db": { x: 600, y: 425 },
  kafka: { x: 720, y: 240 },
};

export default function Status() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
  const [apiGatewayTime, setApiGatewayTime] = useState<number>(0);

  const fetchStatus = useCallback(() => {
    const startTime = performance.now();
    client
      .get("/admin/status")
      .then((res) => {
        const endTime = performance.now();
        const responseTime = Math.round(endTime - startTime);
        setApiGatewayTime(responseTime);

        // Update the frontend → api-gateway connection with measured time
        const data = res.data as SystemStatus;
        const feToGwConn = data.connections.find(
          (c) => c.from === "frontend" && c.to === "api-gateway"
        );
        if (feToGwConn) {
          feToGwConn.response_time_ms = responseTime;
        }

        setStatus(data);
        setLastUpdate(new Date());
        setError("");
        setLoading(false);
      })
      .catch((err) => {
        setError(err.response?.data?.error || "Failed to fetch status");
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 15000); // Refresh every 15s
    return () => clearInterval(interval);
  }, [fetchStatus]);

  const getConnectionColor = (conn: ConnectionStatus) => {
    if (conn.status === "connected") return "#22c55e"; // green
    return "#ef4444"; // red
  };

  const getNodeColor = (name: string) => {
    if (!status) return "#6b7280"; // gray

    // Check if it's a service
    const service = status.services.find((s) => s.name === name);
    if (service) {
      return service.status === "up" ? "#22c55e" : "#ef4444";
    }

    // Check connections for db/kafka status
    const conn = status.connections.find((c) => c.to === name);
    if (conn) {
      return conn.status === "connected" ? "#22c55e" : "#ef4444";
    }

    // Frontend and api-gateway are always up if we got here
    if (name === "frontend" || name === "api-gateway") return "#22c55e";

    return "#6b7280";
  };

  const renderConnection = (from: string, to: string, key: string) => {
    const fromPos = nodePositions[from];
    const toPos = nodePositions[to];
    if (!fromPos || !toPos) return null;

    const conn = status?.connections.find((c) => c.from === from && c.to === to);
    const color = conn ? getConnectionColor(conn) : "#6b7280";
    const responseTime = conn?.response_time_ms;

    // Calculate midpoint for label
    const midX = (fromPos.x + toPos.x) / 2 + 40;
    const midY = (fromPos.y + toPos.y) / 2;

    return (
      <g key={key}>
        <line
          x1={fromPos.x + 60}
          y1={fromPos.y + 20}
          x2={toPos.x}
          y2={toPos.y + 20}
          stroke={color}
          strokeWidth="2"
          markerEnd="url(#arrowhead)"
        />
        {responseTime !== undefined && responseTime > 0 && (
          <text
            x={midX}
            y={midY}
            fontSize="10"
            fill="#9ca3af"
            textAnchor="middle"
          >
            {responseTime}ms
          </text>
        )}
      </g>
    );
  };

  const renderNode = (name: string, label: string, icon: string) => {
    const pos = nodePositions[name];
    if (!pos) return null;

    const color = getNodeColor(name);
    const service = status?.services.find((s) => s.name === name);

    return (
      <g key={name}>
        <rect
          x={pos.x}
          y={pos.y}
          width="120"
          height="40"
          rx="6"
          fill="#1f2937"
          stroke={color}
          strokeWidth="2"
        />
        <text
          x={pos.x + 10}
          y={pos.y + 18}
          fontSize="14"
          fill="#e5e7eb"
        >
          {icon} {label}
        </text>
        {service && (
          <text
            x={pos.x + 10}
            y={pos.y + 32}
            fontSize="10"
            fill={color}
          >
            {service.status === "up" ? "UP" : "DOWN"}
            {service.response_time_ms > 0 && ` (${service.response_time_ms}ms)`}
          </text>
        )}
        <circle
          cx={pos.x + 110}
          cy={pos.y + 10}
          r="5"
          fill={color}
        />
      </g>
    );
  };

  if (loading) return <p className="text-gray-400">Loading status...</p>;

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
        <p className="text-red-400">{error}</p>
        <button
          onClick={fetchStatus}
          className="mt-2 text-sm text-red-400 hover:text-red-300 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">System Status</h1>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-500">
            Last update: {lastUpdate?.toLocaleTimeString()}
          </span>
          <button
            onClick={fetchStatus}
            className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-4 py-2 rounded-lg text-sm hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Legend */}
      <div className="flex gap-6 text-sm">
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 bg-green-500 rounded-full"></span>
          <span className="text-gray-400">Connected / Up</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 bg-red-500 rounded-full"></span>
          <span className="text-gray-400">Disconnected / Down</span>
        </div>
      </div>

      {/* Visual Diagram */}
      <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-4 overflow-x-auto">
        <svg width="880" height="520" className="mx-auto">
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#6b7280" />
            </marker>
          </defs>

          {/* Section Labels */}
          <text x="80" y="25" fontSize="14" fontWeight="bold" fill="#9ca3af">
            Client
          </text>
          <text x="220" y="25" fontSize="14" fontWeight="bold" fill="#9ca3af">
            Gateway
          </text>
          <text x="400" y="25" fontSize="14" fontWeight="bold" fill="#9ca3af">
            Microservices
          </text>
          <text x="620" y="25" fontSize="14" fontWeight="bold" fill="#9ca3af">
            Data Layer
          </text>

          {/* Vertical dividers */}
          <line x1="190" y1="40" x2="190" y2="490" stroke="#374151" strokeDasharray="4" />
          <line x1="370" y1="40" x2="370" y2="490" stroke="#374151" strokeDasharray="4" />
          <line x1="560" y1="40" x2="560" y2="490" stroke="#374151" strokeDasharray="4" />

          {/* Connections */}
          {renderConnection("frontend", "api-gateway", "fe-gw")}
          {renderConnection("api-gateway", "user", "gw-user")}
          {renderConnection("api-gateway", "account", "gw-account")}
          {renderConnection("api-gateway", "card", "gw-card")}
          {renderConnection("api-gateway", "transfer", "gw-transfer")}
          {renderConnection("api-gateway", "payment", "gw-payment")}
          {renderConnection("api-gateway", "notification", "gw-notif")}
          {renderConnection("user", "user-db", "user-db")}
          {renderConnection("account", "account-db", "account-db")}
          {renderConnection("card", "card-db", "card-db")}
          {renderConnection("transfer", "transfer-db", "transfer-db")}
          {renderConnection("payment", "payment-db", "payment-db")}
          {renderConnection("notification", "notification-db", "notif-db")}

          {/* Kafka connections - render separately with curved paths */}
          <g>
            <path
              d={`M ${nodePositions.account.x + 120} ${nodePositions.account.y + 30}
                  Q ${nodePositions.account.x + 200} ${nodePositions.account.y + 80}
                  ${nodePositions.kafka.x} ${nodePositions.kafka.y - 30}`}
              fill="none"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
              strokeDasharray="4"
            />
            <path
              d={`M ${nodePositions.card.x + 120} ${nodePositions.card.y + 20}
                  L ${nodePositions.kafka.x} ${nodePositions.kafka.y}`}
              fill="none"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
              strokeDasharray="4"
            />
            <path
              d={`M ${nodePositions.transfer.x + 120} ${nodePositions.transfer.y + 20}
                  L ${nodePositions.kafka.x} ${nodePositions.kafka.y + 30}`}
              fill="none"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
              strokeDasharray="4"
            />
            <path
              d={`M ${nodePositions.payment.x + 120} ${nodePositions.payment.y + 10}
                  Q ${nodePositions.payment.x + 200} ${nodePositions.payment.y - 40}
                  ${nodePositions.kafka.x} ${nodePositions.kafka.y + 60}`}
              fill="none"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
              strokeDasharray="4"
            />
            <path
              d={`M ${nodePositions.notification.x + 120} ${nodePositions.notification.y + 10}
                  Q ${nodePositions.notification.x + 220} ${nodePositions.notification.y - 100}
                  ${nodePositions.kafka.x} ${nodePositions.kafka.y + 90}`}
              fill="none"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
              strokeDasharray="4"
            />
          </g>

          {/* Nodes */}
          {renderNode("frontend", "Frontend", "🌐")}
          {renderNode("api-gateway", "API Gateway", "🚪")}
          {renderNode("user", "User", "👤")}
          {renderNode("account", "Account", "🏦")}
          {renderNode("card", "Card", "💳")}
          {renderNode("transfer", "Transfer", "💸")}
          {renderNode("payment", "Payment", "💰")}
          {renderNode("notification", "Notification", "🔔")}

          {/* Database nodes */}
          <g>
            {["user", "account", "card", "transfer", "payment", "notification"].map((svc) => {
              const dbName = `${svc}-db`;
              const pos = nodePositions[dbName];
              const color = getNodeColor(dbName);
              return (
                <g key={dbName}>
                  <rect
                    x={pos.x}
                    y={pos.y}
                    width="80"
                    height="40"
                    rx="6"
                    fill="#1f2937"
                    stroke={color}
                    strokeWidth="2"
                  />
                  <text x={pos.x + 10} y={pos.y + 25} fontSize="12" fill="#e5e7eb">
                    🗄️ {svc} DB
                  </text>
                  <circle cx={pos.x + 70} cy={pos.y + 10} r="4" fill={color} />
                </g>
              );
            })}
          </g>

          {/* Kafka node */}
          <g>
            <rect
              x={nodePositions.kafka.x}
              y={nodePositions.kafka.y - 40}
              width="90"
              height="140"
              rx="6"
              fill="#1f2937"
              stroke={getNodeColor("kafka")}
              strokeWidth="2"
            />
            <text x={nodePositions.kafka.x + 15} y={nodePositions.kafka.y + 30} fontSize="12" fill="#e5e7eb">
              📨 Kafka
            </text>
            <circle cx={nodePositions.kafka.x + 80} cy={nodePositions.kafka.y - 30} r="4" fill={getNodeColor("kafka")} />
          </g>
        </svg>
      </div>

      {/* Services Table */}
      <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-800/50">
            <tr>
              <th className="text-left px-4 py-3 text-gray-400 font-medium">Service</th>
              <th className="text-left px-4 py-3 text-gray-400 font-medium">Status</th>
              <th className="text-left px-4 py-3 text-gray-400 font-medium">Response Time</th>
              <th className="text-left px-4 py-3 text-gray-400 font-medium">Error</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700/50">
            {/* API Gateway row - always first */}
            <tr className="hover:bg-gray-800/30">
              <td className="px-4 py-3 font-medium text-white">api-gateway</td>
              <td className="px-4 py-3">
                <span className="inline-block px-2 py-0.5 rounded border text-xs font-medium bg-green-500/20 text-green-400 border-green-500/30">
                  UP
                </span>
              </td>
              <td className="px-4 py-3 text-gray-300">{apiGatewayTime}ms</td>
              <td className="px-4 py-3 text-red-400 text-xs">-</td>
            </tr>
            {status?.services.map((svc) => (
              <tr key={svc.name} className="hover:bg-gray-800/30">
                <td className="px-4 py-3 font-medium text-white">{svc.name}</td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-block px-2 py-0.5 rounded border text-xs font-medium ${
                      svc.status === "up"
                        ? "bg-green-500/20 text-green-400 border-green-500/30"
                        : "bg-red-500/20 text-red-400 border-red-500/30"
                    }`}
                  >
                    {svc.status.toUpperCase()}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-300">{svc.response_time_ms}ms</td>
                <td className="px-4 py-3 text-red-400 text-xs">{svc.error || "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
