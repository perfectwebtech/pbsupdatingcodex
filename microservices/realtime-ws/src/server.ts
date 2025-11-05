import { logger } from "./logger.ts";
import type { RedisClient } from "./redis.ts";
import type { PostgresClient } from "./postgres.ts";
import { verifyJWT } from "./auth.ts";

export interface WSMessage {
  type: string;
  payload?: unknown;
  timestamp?: number;
}

export interface Client {
  id: string;
  socket: WebSocket;
  userId: number;
  username: string;
  channels: Set<string>;
  metadata: {
    ipAddress: string;
    userAgent: string;
    connectedAt: Date;
    lastActivity: Date;
  };
}

export class WebSocketServer {
  private clients: Map<string, Client>;
  private userIdToClients: Map<number, Set<string>>;
  private channelToClients: Map<string, Set<string>>;
  private redis: RedisClient;
  private postgres: PostgresClient;
  private metrics: {
    totalConnections: number;
    activeConnections: number;
    messagesSent: number;
    messagesReceived: number;
  };

  constructor(redis: RedisClient, postgres: PostgresClient) {
    this.clients = new Map();
    this.userIdToClients = new Map();
    this.channelToClients = new Map();
    this.redis = redis;
    this.postgres = postgres;
    this.metrics = {
      totalConnections: 0,
      activeConnections: 0,
      messagesSent: 0,
      messagesReceived: 0,
    };

    // Subscribe to Redis pub/sub for inter-server communication
    this.subscribeToRedis();
  }

  async handleConnection(socket: WebSocket, req: Request): Promise<void> {
    const url = new URL(req.url);
    const token = url.searchParams.get("token");

    if (!token) {
      socket.close(1008, "Missing authentication token");
      return;
    }

    // Verify JWT token
    const payload = await verifyJWT(token);
    if (!payload) {
      socket.close(1008, "Invalid authentication token");
      return;
    }

    // Create client
    const clientId = crypto.randomUUID();
    const client: Client = {
      id: clientId,
      socket,
      userId: payload.user_id,
      username: payload.username,
      channels: new Set(),
      metadata: {
        ipAddress: req.headers.get("x-forwarded-for") || "unknown",
        userAgent: req.headers.get("user-agent") || "unknown",
        connectedAt: new Date(),
        lastActivity: new Date(),
      },
    };

    this.clients.set(clientId, client);
    this.metrics.totalConnections++;
    this.metrics.activeConnections++;

    // Track user connections
    if (!this.userIdToClients.has(client.userId)) {
      this.userIdToClients.set(client.userId, new Set());
    }
    this.userIdToClients.get(client.userId)!.add(clientId);

    logger.info(`Client connected: ${clientId} (User: ${client.username})`);

    // Send welcome message
    this.sendToClient(client, {
      type: "connected",
      payload: {
        clientId,
        userId: client.userId,
        timestamp: Date.now(),
      },
    });

    // Set up message handler
    socket.onmessage = (event) => {
      this.handleMessage(client, event.data);
    };

    // Set up close handler
    socket.onclose = () => {
      this.handleDisconnect(client);
    };

    // Set up error handler
    socket.onerror = (error) => {
      logger.error(`WebSocket error for client ${clientId}:`, error);
    };

    // Publish connection event to Redis
    await this.redis.publish("user:connected", {
      userId: client.userId,
      clientId,
      timestamp: Date.now(),
    });
  }

  private async handleMessage(client: Client, data: string): Promise<void> {
    try {
      const message: WSMessage = JSON.parse(data);
      this.metrics.messagesReceived++;
      client.metadata.lastActivity = new Date();

      logger.debug(`Message from ${client.id}:`, message);

      switch (message.type) {
        case "ping":
          this.sendToClient(client, { type: "pong", timestamp: Date.now() });
          break;

        case "subscribe":
          await this.handleSubscribe(client, message.payload as { channel: string });
          break;

        case "unsubscribe":
          await this.handleUnsubscribe(client, message.payload as { channel: string });
          break;

        case "presence":
          await this.handlePresence(client, message.payload as { streamId: number });
          break;

        case "message":
          await this.handleChatMessage(client, message.payload as { channel: string; text: string });
          break;

        default:
          logger.warn(`Unknown message type: ${message.type}`);
      }
    } catch (error) {
      logger.error(`Error handling message from ${client.id}:`, error);
    }
  }

  private async handleSubscribe(client: Client, payload: { channel: string }): Promise<void> {
    const { channel } = payload;

    client.channels.add(channel);

    if (!this.channelToClients.has(channel)) {
      this.channelToClients.set(channel, new Set());
    }
    this.channelToClients.get(channel)!.add(client.id);

    logger.info(`Client ${client.id} subscribed to channel: ${channel}`);

    this.sendToClient(client, {
      type: "subscribed",
      payload: { channel },
    });

    // Publish subscription event
    await this.redis.publish(`channel:${channel}:join`, {
      userId: client.userId,
      username: client.username,
      timestamp: Date.now(),
    });
  }

  private async handleUnsubscribe(client: Client, payload: { channel: string }): Promise<void> {
    const { channel } = payload;

    client.channels.delete(channel);

    const channelClients = this.channelToClients.get(channel);
    if (channelClients) {
      channelClients.delete(client.id);
      if (channelClients.size === 0) {
        this.channelToClients.delete(channel);
      }
    }

    logger.info(`Client ${client.id} unsubscribed from channel: ${channel}`);

    this.sendToClient(client, {
      type: "unsubscribed",
      payload: { channel },
    });
  }

  private async handlePresence(client: Client, payload: { streamId: number }): Promise<void> {
    const { streamId } = payload;
    const channel = `stream:${streamId}`;

    // Get current viewers count
    const viewersCount = this.channelToClients.get(channel)?.size || 0;

    // Store presence in Redis (expires in 30 seconds)
    await this.redis.setex(
      `presence:stream:${streamId}:${client.userId}`,
      30,
      JSON.stringify({
        userId: client.userId,
        username: client.username,
        timestamp: Date.now(),
      })
    );

    this.sendToClient(client, {
      type: "presence",
      payload: {
        streamId,
        viewersCount: viewersCount + 1,
      },
    });

    // Broadcast to channel
    await this.broadcastToChannel(channel, {
      type: "viewer_joined",
      payload: {
        streamId,
        userId: client.userId,
        username: client.username,
        viewersCount: viewersCount + 1,
      },
    }, client.id);
  }

  private async handleChatMessage(
    client: Client,
    payload: { channel: string; text: string }
  ): Promise<void> {
    const { channel, text } = payload;

    if (!client.channels.has(channel)) {
      return;
    }

    // Store message in database
    await this.postgres.execute(
      `INSERT INTO chat_messages (user_id, channel, message, created_at)
       VALUES ($1, $2, $3, NOW())`,
      [client.userId, channel, text]
    );

    // Broadcast to channel
    await this.broadcastToChannel(channel, {
      type: "chat_message",
      payload: {
        userId: client.userId,
        username: client.username,
        channel,
        text,
        timestamp: Date.now(),
      },
    });
  }

  private handleDisconnect(client: Client): void {
    logger.info(`Client disconnected: ${client.id}`);

    // Remove from clients map
    this.clients.delete(client.id);
    this.metrics.activeConnections--;

    // Remove from user mapping
    const userClients = this.userIdToClients.get(client.userId);
    if (userClients) {
      userClients.delete(client.id);
      if (userClients.size === 0) {
        this.userIdToClients.delete(client.userId);
      }
    }

    // Remove from all channels
    for (const channel of client.channels) {
      const channelClients = this.channelToClients.get(channel);
      if (channelClients) {
        channelClients.delete(client.id);
        if (channelClients.size === 0) {
          this.channelToClients.delete(channel);
        }
      }
    }

    // Publish disconnection event
    this.redis.publish("user:disconnected", {
      userId: client.userId,
      clientId: client.id,
      timestamp: Date.now(),
    });
  }

  private sendToClient(client: Client, message: WSMessage): void {
    try {
      if (client.socket.readyState === WebSocket.OPEN) {
        client.socket.send(JSON.stringify(message));
        this.metrics.messagesSent++;
      }
    } catch (error) {
      logger.error(`Error sending message to client ${client.id}:`, error);
    }
  }

  private async broadcastToChannel(
    channel: string,
    message: WSMessage,
    excludeClientId?: string
  ): Promise<void> {
    const channelClients = this.channelToClients.get(channel);
    if (!channelClients) return;

    for (const clientId of channelClients) {
      if (clientId === excludeClientId) continue;

      const client = this.clients.get(clientId);
      if (client) {
        this.sendToClient(client, message);
      }
    }
  }

  public async broadcastToUser(userId: number, message: WSMessage): Promise<void> {
    const userClients = this.userIdToClients.get(userId);
    if (!userClients) return;

    for (const clientId of userClients) {
      const client = this.clients.get(clientId);
      if (client) {
        this.sendToClient(client, message);
      }
    }
  }

  private async subscribeToRedis(): Promise<void> {
    // Subscribe to broadcast events
    await this.redis.subscribe("broadcast:all", (message) => {
      for (const client of this.clients.values()) {
        this.sendToClient(client, message);
      }
    });

    // Subscribe to user-specific events
    await this.redis.subscribe("broadcast:user:*", (message, channel) => {
      const userId = parseInt(channel.split(":")[2]);
      this.broadcastToUser(userId, message);
    });

    logger.info("Subscribed to Redis pub/sub channels");
  }

  public getMetrics() {
    return {
      ...this.metrics,
      channels: this.channelToClients.size,
      uniqueUsers: this.userIdToClients.size,
    };
  }

  public shutdown(): void {
    logger.info("Shutting down WebSocket server...");

    // Close all connections
    for (const client of this.clients.values()) {
      client.socket.close(1001, "Server shutting down");
    }

    this.clients.clear();
    this.userIdToClients.clear();
    this.channelToClients.clear();

    logger.info("WebSocket server shut down");
  }
}
