/*
  Thin fetch wrapper around the HydraNet core HTTP API.
  In dev, Vite proxies /api → http://localhost:8080 (see vite.config.js).
  Override with VITE_API_BASE for other environments.
*/
const BASE = import.meta.env.VITE_API_BASE ?? '';

export class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request(path, { timeout, headers, ...options } = {}) {
  const controller = new AbortController();
  const timer = timeout ? setTimeout(() => controller.abort(), timeout) : null;

  let res;
  try {
    res = await fetch(`${BASE}${path}`, {
      headers: { 'Content-Type': 'application/json', ...headers },
      signal: controller.signal,
      ...options,
    });
  } catch (err) {
    if (err.name === 'AbortError') {
      throw new ApiError('Request timed out.', 0);
    }
    throw new ApiError('Cannot reach the HydraNet core service.', 0);
  } finally {
    if (timer) clearTimeout(timer);
  }

  const isJson = res.headers.get('content-type')?.includes('application/json');
  const payload = isJson ? await res.json().catch(() => null) : null;

  if (!res.ok) {
    const msg = payload?.error || `Request failed (${res.status})`;
    throw new ApiError(msg, res.status);
  }
  return payload;
}

export const api = {
  get: (path, opts) => request(path, opts),
  post: (path, body, opts) => request(path, { method: 'POST', body: JSON.stringify(body), ...opts }),
  del: (path) => request(path, { method: 'DELETE' }),
};
