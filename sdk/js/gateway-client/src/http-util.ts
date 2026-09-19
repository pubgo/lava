/** Join a base URL and a path without leaving a doubled or missing slash. */
export function joinURL(base: string, path: string): string {
  const b = base.replace(/\/+$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${b}${p}`;
}

/** Merge header sets in priority order; later parts win, so per-call headers override client defaults. */
export function mergeHeaders(...parts: Array<HeadersInit | undefined>): Headers {
  const out = new Headers();
  for (const part of parts) {
    if (!part) continue;
    new Headers(part).forEach((v, k) => out.set(k, v));
  }
  return out;
}
