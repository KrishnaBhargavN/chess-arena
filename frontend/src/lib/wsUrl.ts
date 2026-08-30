// Builds an absolute WebSocket URL on the current origin, e.g. wss://host/api/ws.
// Works the same in dev (via the Vite proxy) and prod (via Caddy) — no env var needed.
export function wsUrl(path: string): string {
  const proto = location.protocol === "https:" ? "wss" : "ws";
  return `${proto}://${location.host}${path}`;
}
