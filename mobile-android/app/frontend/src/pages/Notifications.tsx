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

  if (loading) return <p className="text-gray-400">Loading...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white">
          Notifications{" "}
          {unread > 0 && (
            <span className="text-sm font-normal text-amber-500">
              ({unread} unread)
            </span>
          )}
        </h1>
        {unread > 0 && (
          <button
            onClick={markAllRead}
            className="bg-gradient-to-r from-amber-500 to-amber-600 text-black font-medium px-4 py-2 rounded-lg text-sm hover:from-amber-400 hover:to-amber-500 transition-all duration-200"
          >
            Mark All as Read
          </button>
        )}
      </div>

      {notifications.length === 0 ? (
        <div className="bg-gray-900/80 backdrop-blur-sm border border-gray-700/50 rounded-xl p-8 text-center">
          <svg className="w-12 h-12 text-gray-600 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <p className="text-gray-500">No notifications.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {notifications.map((n) => {
            const isUnread = n.status !== "read";
            return (
              <div
                key={n.id}
                className={`bg-gray-900/80 backdrop-blur-sm border rounded-xl p-4 flex items-start justify-between transition-all duration-200 ${
                  isUnread
                    ? "border-amber-500/50 bg-amber-500/5"
                    : "border-gray-700/50"
                }`}
              >
                <div className="flex-1">
                  <h3
                    className={`text-sm ${
                      isUnread ? "font-semibold text-white" : "font-normal text-gray-400"
                    }`}
                  >
                    {n.title}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">{n.content}</p>
                  <p className="text-xs text-gray-600 mt-2">
                    {new Date(n.created_at).toLocaleString()} &middot;{" "}
                    <span className="text-gray-500">{n.channel}</span> &middot;{" "}
                    <span className="text-gray-500">{n.type}</span>
                  </p>
                </div>
                <div className="flex gap-2 ml-4">
                  {isUnread && (
                    <button
                      onClick={() => markAsRead(n.id)}
                      className="text-xs text-amber-500 hover:text-amber-400 transition-colors"
                    >
                      Mark read
                    </button>
                  )}
                  <button
                    onClick={() => deleteNotification(n.id)}
                    className="text-xs text-red-500 hover:text-red-400 transition-colors"
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
