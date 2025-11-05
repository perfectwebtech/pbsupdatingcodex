import { Client } from "postgres";

export interface PostgresConfig {
  hostname: string;
  port: number;
  username: string;
  password: string;
  database: string;
}

export class PostgresClient {
  private client: Client | null = null;
  private config: PostgresConfig;

  constructor(config: PostgresConfig) {
    this.config = config;
  }

  async connect(): Promise<void> {
    this.client = new Client({
      hostname: this.config.hostname,
      port: this.config.port,
      user: this.config.username,
      password: this.config.password,
      database: this.config.database,
    });

    await this.client.connect();
  }

  async execute(query: string, params?: unknown[]): Promise<unknown> {
    if (!this.client) throw new Error("PostgreSQL client not connected");

    const result = await this.client.queryObject({
      text: query,
      args: params,
    });

    return result.rows;
  }

  async query<T>(query: string, params?: unknown[]): Promise<T[]> {
    if (!this.client) throw new Error("PostgreSQL client not connected");

    const result = await this.client.queryObject<T>({
      text: query,
      args: params,
    });

    return result.rows;
  }

  close(): void {
    if (this.client) {
      this.client.end();
      this.client = null;
    }
  }
}
