import "server-only";

import { auth } from "@/lib/auth/server";
import { cache } from "react";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
  }
}

const getAccessToken = cache(async () => {
  let message = "Authentication required";
  for (let attempt = 0; attempt < 2; attempt++) {
    const { data, error } = await auth.token();
    if (data?.token) return data.token;
    message = error?.message ?? message;
    if (attempt === 0 && /timeout|network|connect/i.test(message)) {
      await new Promise((resolve) => setTimeout(resolve, 250));
      continue;
    }
    break;
  }
  throw new ApiError(message, 401);
});

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  if (!apiURL) {
    throw new Error("NEXT_PUBLIC_API_URL is missing");
  }

  const token = await getAccessToken();

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
