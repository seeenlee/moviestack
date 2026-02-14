"use client";

import { useEffect, useState } from "react";

interface AdminUser {
  id: number;
  username: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const ACTIVE_USER_KEY = "moviestack_active_user";

export default function AdminLoginPage() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeUser, setActiveUser] = useState<AdminUser | null>(null);

  useEffect(() => {
    const raw = window.localStorage.getItem(ACTIVE_USER_KEY);
    if (raw) {
      try {
        const parsed = JSON.parse(raw) as AdminUser;
        if (parsed && typeof parsed.id === "number" && typeof parsed.username === "string") {
          setActiveUser(parsed);
        }
      } catch {
        window.localStorage.removeItem(ACTIVE_USER_KEY);
      }
    }
  }, []);

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch(`${API_URL}/api/admin/users`);
        if (!res.ok) throw new Error("Failed to load users");
        const data: AdminUser[] = await res.json();
        setUsers(data);
      } catch (err) {
        console.error("Load users error:", err);
        setError("Failed to load users");
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  const loginAs = (user: AdminUser) => {
    window.localStorage.setItem(ACTIVE_USER_KEY, JSON.stringify(user));
    setActiveUser(user);
    window.location.href = "/";
  };

  const clearActiveUser = () => {
    window.localStorage.removeItem(ACTIVE_USER_KEY);
    setActiveUser(null);
  };

  return (
    <div className="min-h-screen bg-zinc-50 px-4 py-10 dark:bg-zinc-950">
      <div className="mx-auto w-full max-w-3xl">
        <h1 className="mb-2 text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
          Admin Login
        </h1>
        <p className="mb-6 text-zinc-600 dark:text-zinc-400">
          Pick a user to view the app from their perspective. No real authentication is applied.
        </p>

        {activeUser && (
          <div className="mb-4 flex items-center justify-between rounded-xl border border-zinc-200 bg-white px-4 py-3 shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
            <p className="text-sm text-zinc-700 dark:text-zinc-300">
              Current user: <span className="font-semibold">{activeUser.username}</span> (ID:{" "}
              {activeUser.id})
            </p>
            <button
              type="button"
              onClick={clearActiveUser}
              className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm font-medium text-zinc-700 hover:bg-zinc-100 dark:border-zinc-600 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              Clear
            </button>
          </div>
        )}

        <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
          <div className="border-b border-zinc-200 px-4 py-3 text-sm font-medium text-zinc-700 dark:border-zinc-700 dark:text-zinc-300">
            Users
          </div>

          {loading ? (
            <p className="px-4 py-6 text-sm text-zinc-500 dark:text-zinc-400">Loading users...</p>
          ) : error ? (
            <p className="px-4 py-6 text-sm text-red-600 dark:text-red-300">{error}</p>
          ) : users.length === 0 ? (
            <p className="px-4 py-6 text-sm text-zinc-500 dark:text-zinc-400">
              No users yet. Create one at <code>/admin/users</code>.
            </p>
          ) : (
            <ul>
              {users.map((user) => (
                <li
                  key={user.id}
                  className="flex items-center justify-between border-b border-zinc-100 px-4 py-3 last:border-b-0 dark:border-zinc-800"
                >
                  <div>
                    <p className="font-medium text-zinc-900 dark:text-zinc-100">{user.username}</p>
                    <p className="text-sm text-zinc-500 dark:text-zinc-400">ID: {user.id}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => loginAs(user)}
                    className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300"
                  >
                    Login as
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
