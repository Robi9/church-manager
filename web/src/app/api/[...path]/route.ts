import { NextRequest, NextResponse } from "next/server";

export const runtime = "nodejs";
const production = process.env.NODE_ENV === "production";
const cookieName = production ? "__Host-church_session" : "church_session";
const maxBody = 6 * 1024 * 1024;

function jsonError(status: number, message: string) {
  return NextResponse.json({ data: null, error: message }, {
    status, headers: { "Cache-Control": "no-store" },
  });
}

function expectedOrigin(request: NextRequest): string | null {
  // Serverless adapters can rewrite nextUrl to an internal HTTP address.
  // Use trusted deployment configuration, never client-supplied forwarded headers.
  const configured = process.env.APP_ORIGIN || process.env.URL;
  if (!configured) return request.nextUrl.origin;
  try {
    const url = new URL(configured);
    if (url.username || url.password || url.search || url.hash || url.pathname !== "/") return null;
    if (url.protocol !== "https:" && (production || url.protocol !== "http:")) return null;
    return url.origin;
  } catch {
    return null;
  }
}

// Only fixed API routes are forwarded; no caller-controlled host or headers.
function allowed(path: string, method: string) {
  if (/^auth\/(login|logout|logout-all)$/.test(path)) return method === "POST";
  if (path === "auth/me" || path === "dashboard/stats") return method === "GET";
  if (path === "members") return ["GET", "POST"].includes(method);
  if (/^members\/\d+$/.test(path)) return ["GET", "PUT", "DELETE"].includes(method);
  if (path === "members/check-duplicates" || path === "members/import") return method === "POST";
  if (path === "members/import/template") return method === "GET";
  if (/^members\/import\/errors\/[a-f0-9-]+$/.test(path)) return method === "GET";
  return /^members\/import\/errors\/[a-f0-9-]+\/(confirm|dismiss)$/.test(path) && method === "POST";
}

async function proxy(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path: parts } = await context.params;
  const path = parts.join("/");
  if (!allowed(path, request.method)) return jsonError(404, "Endpoint não encontrado.");

  const mutation = !["GET", "HEAD"].includes(request.method);
  // Check login too, to prevent login CSRF. Missing or invalid origins fail closed.
  const origin = expectedOrigin(request);
  if (mutation && (!origin || request.headers.get("origin") !== origin)) {
    return jsonError(403, "Origem da requisição inválida.");
  }
  if (request.headers.get("sec-fetch-site") === "cross-site") {
    return jsonError(403, "Origem da requisição inválida.");
  }
  const token = request.cookies.get(cookieName)?.value;
  if (path !== "auth/login" && !token) return jsonError(401, "Sessão expirada.");

  const configured = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!configured && production) return jsonError(503, "Serviço indisponível.");
  const base = new URL((configured || "http://localhost:8080/api").replace(/\/$/, "") + "/");
  if (production && base.protocol !== "https:") return jsonError(503, "Serviço indisponível.");
  const target = new URL(path, base);
  target.search = request.nextUrl.search;
  const headers = new Headers();
  if (token && path !== "auth/login") headers.set("Authorization", `Bearer ${token}`);
  const contentType = request.headers.get("content-type");
  if (contentType) headers.set("Content-Type", contentType);

  let body: Uint8Array | undefined;
  if (mutation && request.body) {
    const reader = request.body.getReader();
    const chunks: Uint8Array[] = [];
    let size = 0;
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      size += value.byteLength;
      if (size > maxBody) {
        await reader.cancel();
        return jsonError(413, "Arquivo ou requisição muito grande.");
      }
      chunks.push(value);
    }
    body = new Uint8Array(size);
    let offset = 0;
    for (const chunk of chunks) { body.set(chunk, offset); offset += chunk.byteLength; }
  }

  try {
    const upstream = await fetch(target, {
      method: request.method, headers, body: body as BodyInit | undefined,
      cache: "no-store", redirect: "error", signal: AbortSignal.timeout(65_000),
    });
    if (path === "auth/login" && upstream.ok) {
      const result = await upstream.json();
      if (typeof result.data?.token !== "string" || result.data.token.length !== 43) {
        return jsonError(502, "Resposta inválida do serviço.");
      }
      const response = NextResponse.json({ data: { authenticated: true }, error: null }, {
        headers: { "Cache-Control": "no-store" },
      });
      response.cookies.set(cookieName, result.data.token, {
        httpOnly: true, secure: production, sameSite: "lax", path: "/",
        maxAge: Math.min(Number(result.data.expires_in) || 7200, 7200),
      });
      return response;
    }
    const responseHeaders = new Headers({ "Cache-Control": "no-store" });
    for (const name of ["content-type", "content-disposition"]) {
      const value = upstream.headers.get(name);
      if (value) responseHeaders.set(name, value);
    }
    const response = new NextResponse(upstream.body, { status: upstream.status, headers: responseHeaders });
    if (upstream.status === 401 || (upstream.ok && ["auth/logout", "auth/logout-all"].includes(path))) {
      response.cookies.set(cookieName, "", { httpOnly: true, secure: production, sameSite: "lax", path: "/", maxAge: 0 });
    }
    return response;
  } catch {
    return jsonError(503, "Serviço temporariamente indisponível. Tente novamente.");
  }
}

export { proxy as GET, proxy as POST, proxy as PUT, proxy as DELETE, proxy as PATCH };
