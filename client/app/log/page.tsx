"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

interface ActiveUser {
  id: number;
  username: string;
}

interface MovieLogEntry {
  log_id: number;
  user_id: number;
  movie_id: number;
  original_title: string;
  watched_on: string;
  note: string | null;
  rank_position: number | null;
  created_at: string;
  updated_at: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const ACTIVE_USER_KEY = "moviestack_active_user";

export default function MovieLogPage() {
  const [activeUser, setActiveUser] = useState<ActiveUser | null>(null);
  const [entries, setEntries] = useState<MovieLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deletingLogId, setDeletingLogId] = useState<number | null>(null);

  useEffect(() => {
    const raw = window.localStorage.getItem(ACTIVE_USER_KEY);
    if (!raw) {
      setLoading(false);
      return;
    }

    try {
      const parsed = JSON.parse(raw) as ActiveUser;
      if (parsed && typeof parsed.id === "number" && typeof parsed.username === "string") {
        setActiveUser(parsed);
      } else {
        window.localStorage.removeItem(ACTIVE_USER_KEY);
      }
    } catch {
      window.localStorage.removeItem(ACTIVE_USER_KEY);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchLog = async (userId: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_URL}/api/users/${userId}/log`);
      if (!res.ok) {
        const payload = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error || "Failed to load movie log");
      }

      const data: MovieLogEntry[] = await res.json();
      setEntries(data);
    } catch (err) {
      console.error("Load movie log error:", err);
      setError(err instanceof Error ? err.message : "Failed to load movie log");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!activeUser) return;
    fetchLog(activeUser.id);
  }, [activeUser]);

  const rankedEntries = useMemo(
    () => entries.filter((entry) => entry.rank_position !== null),
    [entries]
  );
  const unrankedEntries = useMemo(
    () => entries.filter((entry) => entry.rank_position === null),
    [entries]
  );

  const onDelete = async (entry: MovieLogEntry) => {
    if (!activeUser) return;
    if (!window.confirm(`Delete "${entry.original_title}" from your log?`)) return;

    setDeletingLogId(entry.log_id);
    setError(null);
    try {
      const res = await fetch(`${API_URL}/api/users/${activeUser.id}/log/${entry.log_id}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        const payload = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error || "Failed to delete log entry");
      }

      setEntries((prev) => prev.filter((row) => row.log_id !== entry.log_id));
    } catch (err) {
      console.error("Delete movie log error:", err);
      setError(err instanceof Error ? err.message : "Failed to delete log entry");
    } finally {
      setDeletingLogId(null);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-50 px-4 py-10 dark:bg-zinc-950">
      <div className="mx-auto w-full max-w-3xl">
        <div className="mb-4 flex flex-wrap gap-2 text-sm">
          <Link
            href="/"
            className="rounded-md border border-zinc-300 bg-white px-3 py-1.5 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            Movie Search
          </Link>
          <Link
            href="/log"
            className="rounded-md border border-zinc-300 bg-white px-3 py-1.5 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            My Log
          </Link>
          <Link
            href="/admin/login"
            className="rounded-md border border-zinc-300 bg-white px-3 py-1.5 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            Admin Login
          </Link>
          <Link
            href="/admin/users"
            className="rounded-md border border-zinc-300 bg-white px-3 py-1.5 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            Admin Users
          </Link>
        </div>

        <h1 className="mb-2 text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
          My Movie Log
        </h1>

        {!activeUser && !loading ? (
          <p className="rounded-xl border border-zinc-200 bg-white px-4 py-3 text-sm text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-400">
            No active user selected. Go to <Link href="/admin/login" className="underline">/admin/login</Link>{" "}
            to pick one.
          </p>
        ) : (
          <p className="mb-4 text-sm text-zinc-600 dark:text-zinc-400">
            Viewing as <span className="font-semibold">{activeUser?.username}</span>
          </p>
        )}

        {error && (
          <p className="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/40 dark:text-red-300">
            {error}
          </p>
        )}

        {loading ? (
          <p className="text-sm text-zinc-500 dark:text-zinc-400">Loading movie log...</p>
        ) : activeUser ? (
          <div className="space-y-6">
            <section className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
              <div className="border-b border-zinc-200 px-4 py-3 text-sm font-medium text-zinc-700 dark:border-zinc-700 dark:text-zinc-300">
                Ranked
              </div>
              {rankedEntries.length === 0 ? (
                <p className="px-4 py-5 text-sm text-zinc-500 dark:text-zinc-400">
                  No ranked movies yet.
                </p>
              ) : (
                <ul>
                  {rankedEntries.map((entry) => (
                    <li
                      key={entry.log_id}
                      className="flex items-center justify-between border-b border-zinc-100 px-4 py-3 last:border-b-0 dark:border-zinc-800"
                    >
                      <div>
                        <p className="font-medium text-zinc-900 dark:text-zinc-100">
                          #{entry.rank_position} {entry.original_title}
                        </p>
                        <p className="text-sm text-zinc-500 dark:text-zinc-400">
                          Watched: {entry.watched_on}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => onDelete(entry)}
                        disabled={deletingLogId === entry.log_id}
                        className="rounded-md border border-red-200 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-red-900/60 dark:text-red-300 dark:hover:bg-red-950/40"
                      >
                        {deletingLogId === entry.log_id ? "Deleting..." : "Delete"}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
              <div className="border-b border-zinc-200 px-4 py-3 text-sm font-medium text-zinc-700 dark:border-zinc-700 dark:text-zinc-300">
                Unranked
              </div>
              {unrankedEntries.length === 0 ? (
                <p className="px-4 py-5 text-sm text-zinc-500 dark:text-zinc-400">
                  No unranked movies yet. Add one from search.
                </p>
              ) : (
                <ul>
                  {unrankedEntries.map((entry) => (
                    <li
                      key={entry.log_id}
                      className="flex items-center justify-between border-b border-zinc-100 px-4 py-3 last:border-b-0 dark:border-zinc-800"
                    >
                      <div>
                        <p className="font-medium text-zinc-900 dark:text-zinc-100">
                          {entry.original_title}
                        </p>
                        <p className="text-sm text-zinc-500 dark:text-zinc-400">
                          Watched: {entry.watched_on}
                          {entry.note ? ` | Note: ${entry.note}` : ""}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => onDelete(entry)}
                        disabled={deletingLogId === entry.log_id}
                        className="rounded-md border border-red-200 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-red-900/60 dark:text-red-300 dark:hover:bg-red-950/40"
                      >
                        {deletingLogId === entry.log_id ? "Deleting..." : "Delete"}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>
        ) : null}
      </div>
    </div>
  );
}
