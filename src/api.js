// API helper — uses Vite proxy in dev, direct URLs in production
const URLS = {
  koola10: import.meta.env.VITE_KOOLA10_API_URL || '',
  spiral: import.meta.env.VITE_SPIRAL_API_URL || '',
  apex: import.meta.env.VITE_APEX_API_URL || '',
};

export const ADMIN_KEY = import.meta.env.VITE_ADMIN_API_KEY || '';

/**
 * Build a fetch URL for a swarm service.
 * In dev, routes through Vite proxy (/api/<service>/...) to avoid CORS.
 * In production, uses the direct URL from env vars.
 */
export function apiUrl(service, path) {
  if (import.meta.env.DEV) {
    return `/api/${service}${path}`;
  }
  return `${URLS[service]}${path}`;
}

/**
 * Get the direct base URL for a service (used by SystemHealth for /health checks).
 */
export function serviceUrl(service) {
  if (import.meta.env.DEV) {
    return `/api/${service}`;
  }
  return URLS[service];
}

/**
 * Default headers for API calls.
 */
export function authHeaders() {
  return ADMIN_KEY ? { 'Authorization': `Bearer ${ADMIN_KEY}`, 'Content-Type': 'application/json' } : { 'Content-Type': 'application/json' };
}
