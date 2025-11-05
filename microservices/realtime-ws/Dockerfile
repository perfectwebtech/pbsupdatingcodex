FROM denoland/deno:1.40.0

WORKDIR /app

# Copy dependency files
COPY deno.json .
COPY .env.example .env

# Copy source code
COPY main.ts .
COPY src/ ./src/

# Cache dependencies
RUN deno cache main.ts

# Expose port
EXPOSE 8001

# Run
CMD ["deno", "run", "--allow-net", "--allow-env", "--allow-read", "main.ts"]
