import { useEffect, useState } from "react";
import client from "../api/client";
import type { Notification } from "../types";

export default function Notifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unread, setUnread] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetchNotifications = () => {
    client
      .get("/notifications")
      .then((res) => {
        setNotifications(res.data.notifications ?? []);
        setUnread(res.data.unread ?? 0);
      })
      .catch(() => setNotifications([]))
      .finally(() => setLoading(false));
  };

  useEffect(fetchNotifications, []);

  const markAsRead = async (id: number) => {
    try {
      await client.put(`/notifications/${id}/read`);
      fetchNotifications();
    } catch {
      // ignore
    }
  };

  const markAllRead = async () => {
    try {
      await client.put("/notifications/read-all");
      fetchNotifications();
    } catch {
      // ignore
    }
  };

  const deleteNotification = async (id: number) => {
    try {
      await client.delete(`/notifications/${id}`);
      fetchNotifications();
    } catch {
      // ignore
    }
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">
          Notifications{" "}
          {unread > 0 && (
            <span className="text-sm font-normal text-gray-500">
              ({unread} unread)
            </span>
          )}
        </h1>
        {unread > 0 && (
          <button
            onClick={markAllRead}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
          >
            Mark All as Read
          </button>
        )}
      </div>

      {notifications.length === 0 ? (
        <p className="text-sm text-gray-500">No notifications.</p>
      ) : (
        <div className="space-y-2">
          {notifications.map((n) => {
            const isUnread = n.status !== "read";
            return (
              <div
                key={n.id}
                className={`bg-white rounded shadow p-4 flex items-start justify-between ${
                  isUnread ? "border-l-4 border-blue-500" : ""
                }`}
              >
                <div className="flex-1">
                  <h3
                    className={`text-sm ${
                      isUnread ? "font-semibold" : "font-normal text-gray-600"
                    }`}
                  >
                    {n.title}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">{n.content}</p>
                  <p className="text-xs text-gray-400 mt-1">
                    {new Date(n.created_at).toLocaleString()} &middot;{" "}
                    {n.channel} &middot; {n.type}
                  </p>
                </div>
                <div className="flex gap-2 ml-4">
                  {isUnread && (
                    <button
                      onClick={() => markAsRead(n.id)}
                      className="text-xs text-blue-600 hover:underline"
                    >
                      Mark read
                    </button>
                  )}
                  <button
                    onClick={() => deleteNotification(n.id)}
                    className="text-xs text-red-600 hover:underline"
                  >
                    Delete
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
