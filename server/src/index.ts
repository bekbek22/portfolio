import { app } from "./app";
import { env } from "./config/env";
import { prisma } from "./lib/prisma";

app.listen(env.PORT);
console.log(`API listening on http://localhost:${env.PORT}/api/v1`);

async function shutdown(signal: string) {
  console.log(`${signal} received, shutting down`);
  await app.stop();
  await prisma.$disconnect();
  process.exit(0);
}

process.on("SIGINT", () => void shutdown("SIGINT"));
process.on("SIGTERM", () => void shutdown("SIGTERM"));
