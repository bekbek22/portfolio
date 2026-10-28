import { describe, expect, test } from "bun:test";
import { getClientIp } from "../src/lib/client-ip";

describe("getClientIp", () => {
  test("uses the socket address when no proxies are trusted", () => {
    expect(getClientIp("6.6.6.6", "1.1.1.1", 0)).toBe("1.1.1.1");
  });

  test("takes the entry appended by the outermost trusted proxy", () => {
    // spoofed by client, real client (appended by proxy A), proxy A (appended by proxy B)
    expect(getClientIp("6.6.6.6, 2.2.2.2, 10.0.0.1", "10.0.0.2", 2)).toBe("2.2.2.2");
    expect(getClientIp("6.6.6.6, 2.2.2.2", "10.0.0.2", 1)).toBe("2.2.2.2");
  });

  test("falls back to the socket when the header is missing or too short", () => {
    expect(getClientIp(null, "1.1.1.1", 1)).toBe("1.1.1.1");
    expect(getClientIp("2.2.2.2", "1.1.1.1", 3)).toBe("1.1.1.1");
    expect(getClientIp(null, undefined, 0)).toBe("unknown");
  });
});
