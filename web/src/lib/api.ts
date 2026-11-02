export const API_URL = "/api";

interface RequestOptions {
  method?: string;
  body?: unknown;
}

interface ApiResponse<T> {
  data: T;
  error: string | null;
}

export class ApiError<T = unknown> extends Error {
  status: number;
  data: T | null;

  constructor(message: string, status: number, data: T | null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

export async function sessionFetch(url: string, options: RequestInit = {}) {
  const response = await fetch(url, { ...options, cache: "no-store", credentials: "same-origin" });
  if (response.status === 401 && url !== `${API_URL}/auth/login`) {
    // Full navigation discards member data still held in client state.
    // eslint-disable-next-line @next/next/no-location-assign-relative-destination
    window.location.href = "/login";
    throw new Error("Sessão expirada. Entre novamente.");
  }
  return response;
}

export async function api<T>(
  path: string,
  { method = "GET", body }: RequestOptions = {}
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const res = await sessionFetch(`${API_URL}${path}`, {
    method,
    cache: "no-store",
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const json = await res.json();

  if (!res.ok) {
    throw new ApiError(
      json.error || "An unexpected error occurred",
      res.status,
      json.data ?? null,
    );
  }

  return json;
}
