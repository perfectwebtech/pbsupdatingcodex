import * as jose from "https://deno.land/x/jose@v5.2.0/index.ts";

const JWT_SECRET = Deno.env.get("JWT_SECRET") || "your-secret-key";

export interface JWTPayload {
  user_id: number;
  username: string;
  iat: number;
  exp: number;
  iss: string;
}

export async function verifyJWT(token: string): Promise<JWTPayload | null> {
  try {
    const secret = new TextEncoder().encode(JWT_SECRET);

    const { payload } = await jose.jwtVerify(token, secret, {
      issuer: "iptv-platform",
    });

    return payload as unknown as JWTPayload;
  } catch (error) {
    console.error("JWT verification failed:", error);
    return null;
  }
}
