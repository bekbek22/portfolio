import type { NextConfig } from "next";

// Server-side only: where the Go API lives (Railway in production, localhost in dev).
const apiUrl = process.env.API_URL ?? "http://localhost:4000";

const nextConfig: NextConfig = {
  cacheComponents: true,
  partialPrefetching: true,
  // Browser calls to /api/* are forwarded to the Go API, so cookies it sets are
  // first-party. Server Components call API_URL directly instead.
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiUrl}/api/:path*` }];
  },
  turbopack: {
    // Pin the root: otherwise Next scans parent dirs for lockfiles (incl. the home directory).
    root: __dirname,
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
