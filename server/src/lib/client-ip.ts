/**
 * Resolve the caller's IP.
 *
 * Each proxy appends the address it received the request from to X-Forwarded-For, so with
 * `trustedHops` proxies in front of us the client is `trustedHops` entries from the right.
 * Entries further left are client-controlled and can be spoofed, so they are never used.
 */
export function getClientIp(
  forwardedFor: string | null,
  socketIp: string | undefined,
  trustedHops: number,
): string {
  if (trustedHops > 0 && forwardedFor) {
    const hops = forwardedFor.split(",").map((s) => s.trim());
    const ip = hops[hops.length - trustedHops];
    if (ip) return ip;
  }
  return socketIp ?? "unknown";
}
