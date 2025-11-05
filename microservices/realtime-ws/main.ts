import { serve } from "@std/http/server";
import { load } from "@std/dotenv";
import { WebSocketServer } from "./src/server.ts";
import { RedisClient } from "./src/redis.ts";
import { PostgresClient } from "./src/postgres.ts";
import { logger } from "./src/logger.ts";

// Load environment variables
await load({ export: true });

const PORT = parseInt(Deno.env.get("PORT") || "8001");
const HOST = Deno.env.get("HOST") || "0.0.0.0";

// Initialize Redis
const redisClient = new RedisClient({
  hostname: Deno.env.get("REDIS_HOST") || "localhost",
  port: parseInt(Deno.env.get("REDIS_PORT") || "6379"),
});

await redisClient.connect();
logger.info("Redis connected");

// Initialize PostgreSQL
const postgresClient = new PostgresClient({
  hostname: Deno.env.get("DB_HOST") || "localhost",
  port: parseInt(Deno.env.get("DB_PORT") || "5432"),
  username: Deno.env.get("DB_USER") || "iptv_user",
  password: Deno.env.get("DB_PASSWORD") || "",
  database: Deno.env.get("DB_NAME") || "iptv_realtime",
});

await postgresClient.connect();
logger.info("PostgreSQL connected");

// Initialize WebSocket server
const wsServer = new WebSocketServer(redisClient, postgresClient);

// HTTP handler
const handler = async (req: Request): Promise<Response> => {
  const url = new URL(req.url);

  // Health check
  if (url.pathname === "/health") {
    return new Response(
      JSON.stringify({ status: "ok", service: "realtime-ws" }),
      {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }
    );
  }

  // Metrics endpoint
  if (url.pathname === "/metrics") {
    const metrics = wsServer.getMetrics();
    return new Response(JSON.stringify(metrics), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }

  // WebSocket upgrade
  if (url.pathname === "/ws") {
    if (req.headers.get("upgrade") !== "websocket") {
      return new Response("Expected WebSocket connection", { status: 426 });
    }

    const { socket, response } = Deno.upgradeWebSocket(req);
    await wsServer.handleConnection(socket, req);

    return response;
  }

  return new Response("Not Found", { status: 404 });
};

// Graceful shutdown
const abortController = new AbortController();

Deno.addSignalListener("SIGINT", () => {
  logger.info("Shutting down server...");
  abortController.abort();
  wsServer.shutdown();
  redisClient.close();
  postgresClient.close();
  Deno.exit(0);
});

// Start server
logger.info(`WebSocket server starting on ${HOST}:${PORT}`);

await serve(handler, {
  port: PORT,
  hostname: HOST,
  signal: abortController.signal,
  onListen: ({ hostname, port }) => {
    logger.info(`Server listening on ${hostname}:${port}`);
  },
});
