const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

interface RequestOptions {
  method?: string;
  body?: unknown;
  token?: string | null;
}

interface ApiResponse<T> {
  data: T;
  error: string | null;
}

export async function api<T>(
  path: string,
  { method = "GET", body, token }: RequestOptions = {}
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    localStorage.removeItem("church_manager_token");
    window.location.href = "/login";
    throw new Error("Session expired");
  }

  const json = await res.json();

  if (!res.ok) {
    throw new Error(json.error || "An unexpected error occurred");
  }

  return json;
}
