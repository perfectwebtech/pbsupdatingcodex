import { connect, type Redis } from "redis";

export interface RedisConfig {
  hostname: string;
  port: number;
  password?: string;
}

export class RedisClient {
  private client: Redis | null = null;
  private config: RedisConfig;
  private subscribers: Map<string, Set<(message: unknown, channel: string) => void>>;

  constructor(config: RedisConfig) {
    this.config = config;
    this.subscribers = new Map();
  }

  async connect(): Promise<void> {
    this.client = await connect({
      hostname: this.config.hostname,
      port: this.config.port,
      password: this.config.password,
    });
  }

  async publish(channel: string, message: unknown): Promise<void> {
    if (!this.client) throw new Error("Redis client not connected");
    await this.client.publish(channel, JSON.stringify(message));
  }

  async subscribe(
    pattern: string,
    callback: (message: unknown, channel: string) => void
  ): Promise<void> {
    if (!this.client) throw new Error("Redis client not connected");

    if (!this.subscribers.has(pattern)) {
      this.subscribers.set(pattern, new Set());

      // Subscribe to Redis pattern
      const sub = await connect({
        hostname: this.config.hostname,
        port: this.config.port,
        password: this.config.password,
      });

      if (pattern.includes("*")) {
        await sub.psubscribe(pattern);
      } else {
        await sub.subscribe(pattern);
      }

      // Listen for messages
      (async () => {
        for await (const message of sub.receive()) {
          const channel = message.channel;
          const data = JSON.parse(message.message as string);

          const callbacks = this.subscribers.get(pattern);
          if (callbacks) {
            for (const cb of callbacks) {
              cb(data, channel);
            }
          }
        }
      })();
    }

    this.subscribers.get(pattern)!.add(callback);
  }

  async get(key: string): Promise<string | null> {
    if (!this.client) throw new Error("Redis client not connected");
    return await this.client.get(key);
  }

  async set(key: string, value: string): Promise<void> {
    if (!this.client) throw new Error("Redis client not connected");
    await this.client.set(key, value);
  }

  async setex(key: string, seconds: number, value: string): Promise<void> {
    if (!this.client) throw new Error("Redis client not connected");
    await this.client.setex(key, seconds, value);
  }

  async del(key: string): Promise<void> {
    if (!this.client) throw new Error("Redis client not connected");
    await this.client.del(key);
  }

  close(): void {
    if (this.client) {
      this.client.quit();
      this.client = null;
    }
  }
}
