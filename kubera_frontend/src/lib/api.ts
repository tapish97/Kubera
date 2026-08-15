import "server-only";

import { auth } from "@/lib/auth/server";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
  }
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  if (!apiURL) {
    throw new Error("NEXT_PUBLIC_API_URL is missing");
  }

  const { data: session } = await auth.getSession();
  const token = session?.session?.token;
  if (!token) {
    throw new ApiError("Authentication required", 401);
  }

  const headers = new Headers(init.headers);
  headers.set("Authorization", `Bearer ${token}`);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${apiURL.replace(/\/$/, "")}${path}`, {
    ...init,
    headers,
    cache: init.cache ?? "no-store",
  });

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as
      | { error?: string }
      | null;
    throw new ApiError(body?.error ?? "Kubera API request failed", response.status);
  }

  return (await response.json()) as T;
}
