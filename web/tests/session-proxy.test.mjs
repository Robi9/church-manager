import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import { createRequire } from 'node:module';
import ts from 'typescript';
import { NextRequest } from 'next/server.js';

const require = createRequire(import.meta.url);

// Compile the actual route with the project's TypeScript compiler, then supply
// a fake upstream. No live credentials or production services are used.
function route(upstream, environment = {}) {
  const source = fs.readFileSync('src/app/api/[...path]/route.ts', 'utf8');
  const compiled = ts.transpileModule(source, { compilerOptions: {
    module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022,
  } }).outputText;
  const exports = {};
  vm.runInNewContext(compiled, {
    exports, require, URL, Headers, Uint8Array, AbortSignal,
    process: { env: { NODE_ENV: 'production', API_URL: 'https://api.example.org/api', ...environment } },
    fetch: upstream,
  });
  return exports;
}
function request(path, options = {}) {
  return new NextRequest(`https://church.example.org/api/${path}`, options);
}
const context = (path) => ({ params: Promise.resolve({ path: path.split('/') }) });

test('login sets a secure HttpOnly cookie and never returns the session token', async () => {
  const token = 'a'.repeat(43);
  const handler = route(async (url, options) => {
    assert.equal(String(url), 'https://api.example.org/api/auth/login');
    assert.equal(options.headers.get('authorization'), null);
    return Response.json({ data: { token, expires_in: 7200 }, error: null });
  });
  const result = await handler.POST(request('auth/login', {
    method: 'POST', headers: { origin: 'https://church.example.org', 'content-type': 'application/json' },
    body: JSON.stringify({ email: 'test@example.org', password: 'example-only' }),
  }), context('auth/login'));
  const cookie = result.headers.get('set-cookie');
  assert.match(cookie, /__Host-church_session=/);
  assert.match(cookie, /HttpOnly/i);
  assert.match(cookie, /Secure/);
  assert.match(cookie, /SameSite=lax/i);
  assert.match(cookie, /Max-Age=7200/);
  assert.equal(result.headers.get('cache-control'), 'no-store');
  assert.equal((await result.text()).includes(token), false);
});

test('cross-site and missing-origin mutations, including login, are blocked', async () => {
  const handler = route(() => { throw new Error('must not contact API'); });
  for (const origin of [undefined, 'https://attacker.example']) {
    const headers = origin ? { origin } : {};
    const result = await handler.POST(request('auth/login', { method: 'POST', headers }), context('auth/login'));
    assert.equal(result.status, 403);
  }
});

test('browser Authorization headers are ignored; only the HttpOnly cookie is forwarded', async () => {
  const token = 'b'.repeat(43);
  const handler = route(async (_, options) => {
    assert.equal(options.headers.get('authorization'), `Bearer ${token}`);
    assert.equal(options.headers.get('cookie'), null);
    return Response.json({ data: { user_id: 9 }, error: null });
  });
  const result = await handler.GET(request('auth/me', { headers: {
    cookie: `__Host-church_session=${token}; unrelated=secret`, authorization: 'Bearer forged',
  } }), context('auth/me'));
  assert.equal(result.status, 200);
  const missing = await handler.GET(request('auth/me', { headers: { authorization: `Bearer ${token}` } }), context('auth/me'));
  assert.equal(missing.status, 401);
});

test('successful revocation clears the cookie; failed revocation keeps it for retry', async () => {
  for (const status of [200, 503]) {
    const handler = route(async () => Response.json({ data: null }, { status }));
    const result = await handler.POST(request('auth/logout-all', {
      method: 'POST', headers: { origin: 'https://church.example.org', cookie: `__Host-church_session=${'c'.repeat(43)}` },
    }), context('auth/logout-all'));
    assert.equal(result.status, status);
    if (status === 200) assert.match(result.headers.get('set-cookie'), /Max-Age=0/);
    else assert.equal(result.headers.get('set-cookie'), null);
  }
});

test('unlisted routes cannot turn the frontend into an open proxy', async () => {
  const handler = route(() => { throw new Error('must not contact API'); });
  for (const path of ['auth/register', 'https://attacker.example', '../health', 'members/../auth/login']) {
    const result = await handler.GET(request(path), context(path));
    assert.equal(result.status, 404);
  }
});

test('oversized uploads are rejected before contacting the API', async () => {
  const handler = route(() => { throw new Error('must not contact API'); });
  const result = await handler.POST(request('members/import', {
    method: 'POST', headers: { origin: 'https://church.example.org', cookie: `__Host-church_session=${'d'.repeat(43)}` },
    body: 'x'.repeat(6 * 1024 * 1024 + 1),
  }), context('members/import'));
  assert.equal(result.status, 413);
});


test('Netlify public URL validates login when the adapter rewrites the request URL', async () => {
  const handler = route(async () => Response.json({ data: { token: 'e'.repeat(43), expires_in: 7200 } }), {
    URL: 'https://church.example.org',
  });
  const result = await handler.POST(new NextRequest('http://localhost:3000/api/auth/login', {
    method: 'POST', headers: { origin: 'https://church.example.org', 'sec-fetch-site': 'same-origin' },
  }), context('auth/login'));
  assert.equal(result.status, 200);
  assert.match(result.headers.get('set-cookie'), /Secure/);
});

test('explicit public origin overrides the Netlify URL for custom domains', async () => {
  const handler = route(async () => Response.json({ data: { token: 'f'.repeat(43) } }), {
    APP_ORIGIN: 'https://custom.example.org/', URL: 'https://church.example.org',
  });
  const result = await handler.POST(request('auth/login', {
    method: 'POST', headers: { origin: 'https://custom.example.org' },
  }), context('auth/login'));
  assert.equal(result.status, 200);
});

test('public origin configuration still rejects foreign, missing, and spoofed origins', async () => {
  const handler = route(() => { throw new Error('must not contact API'); }, {
    URL: 'https://church.example.org',
  });
  for (const headers of [
    { origin: 'https://attacker.example', 'x-forwarded-host': 'attacker.example', 'x-forwarded-proto': 'https' },
    { origin: 'https://church.example.org.attacker.example' },
    { origin: 'null' },
    { origin: 'http://localhost:3000' },
    { 'x-forwarded-origin': 'https://church.example.org' },
    { origin: 'https://church.example.org', 'sec-fetch-site': 'cross-site' },
  ]) {
    const result = await handler.POST(new NextRequest('http://localhost:3000/api/auth/login', {
      method: 'POST', headers,
    }), context('auth/login'));
    assert.equal(result.status, 403);
  }
});

test('invalid configured public origins fail closed', async () => {
  for (const APP_ORIGIN of ['not-a-url', 'http://church.example.org', 'https://church.example.org/path',
    'https://user:password@church.example.org', 'https://church.example.org?query=value']) {
    const handler = route(() => { throw new Error('must not contact API'); }, { APP_ORIGIN });
    const result = await handler.POST(request('auth/login', {
      method: 'POST', headers: { origin: 'https://church.example.org' },
    }), context('auth/login'));
    assert.equal(result.status, 403);
  }
});
