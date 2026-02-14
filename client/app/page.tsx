"use client";

import { useState, useEffect, useRef } from "react";

interface Movie {
  id: number;
  original_title: string;
  adult: boolean;
  video: boolean;
  popularity: number;
  score: number;
}

interface ActiveUser {
  id: number;
  username: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const ACTIVE_USER_KEY = "moviestack_active_user";

export default function Home() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Movie[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [activeUser, setActiveUser] = useState<ActiveUser | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const raw = window.localStorage.getItem(ACTIVE_USER_KEY);
    if (!raw) return;
    try {
      const parsed = JSON.parse(raw) as ActiveUser;
      if (parsed && typeof parsed.id === "number" && typeof parsed.username === "string") {
        setActiveUser(parsed);
      }
    } catch {
      window.localStorage.removeItem(ACTIVE_USER_KEY);
    }
  }, []);

  useEffect(() => {
    if (query.trim() === "") {
      setResults([]);
      setHasSearched(false);
      return;
    }

    const timeout = setTimeout(async () => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setLoading(true);
      try {
        const res = await fetch(
          `${API_URL}/api/movies/search?q=${encodeURIComponent(query.trim())}`,
          { signal: controller.signal }
        );
        if (!res.ok) throw new Error("Search failed");
        const data: Movie[] = await res.json();
        setResults(data);
        setHasSearched(true);
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") return;
        console.error("Search error:", err);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timeout);
  }, [query]);

  const onClearUser = () => {
    window.localStorage.removeItem(ACTIVE_USER_KEY);
    setActiveUser(null);
  };

  return (
    <div className="flex min-h-screen flex-col items-center bg-zinc-50 px-4 pt-16 font-sans dark:bg-zinc-950">
      <div className="mb-8 w-full max-w-xl rounded-xl border border-zinc-200 bg-white p-3 text-sm shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
        {activeUser ? (
          <div className="flex items-center justify-between gap-3">
            <p className="text-zinc-700 dark:text-zinc-300">
              Viewing as <span className="font-semibold">{activeUser.username}</span> (ID:{" "}
              {activeUser.id})
            </p>
            <button
              type="button"
              onClick={onClearUser}
              className="rounded-md border border-zinc-300 px-2.5 py-1 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-600 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              Clear
            </button>
          </div>
        ) : (
          <p className="text-zinc-600 dark:text-zinc-400">
            No active user selected. Go to <code>/admin/login</code> to pick one.
          </p>
        )}
      </div>

      <h1 className="mb-8 text-4xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
        MovieStack
      </h1>

      <div className="w-full max-w-xl">
        <div className="relative">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search for a movie..."
            autoFocus
            className="w-full rounded-xl border border-zinc-200 bg-white px-5 py-3.5 text-lg text-zinc-900 shadow-sm outline-none transition-all placeholder:text-zinc-400 focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100 dark:placeholder:text-zinc-500 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
          />
          {loading && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600 dark:border-zinc-600 dark:border-t-zinc-300" />
            </div>
          )}
        </div>

        {hasSearched && results.length === 0 && !loading && (
          <p className="mt-6 text-center text-zinc-500 dark:text-zinc-400">
            No movies found for &ldquo;{query}&rdquo;
          </p>
        )}

        {results.length > 0 && (
          <ul className="mt-3 overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900">
            {results.map((movie) => (
              <li
                key={movie.id}
                className="flex items-center justify-between border-b border-zinc-100 px-5 py-3.5 last:border-b-0 hover:bg-zinc-50 dark:border-zinc-800 dark:hover:bg-zinc-800/50"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate text-base font-medium text-zinc-900 dark:text-zinc-100">
                    {movie.original_title}
                  </p>
                  <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                    Popularity: {movie.popularity.toLocaleString()}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
