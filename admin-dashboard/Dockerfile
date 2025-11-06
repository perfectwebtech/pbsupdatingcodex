# ============================================================================
# Multi-stage Dockerfile for IPTV Admin Dashboard
# ============================================================================

# ============================================================================
# Stage 1: Builder
# ============================================================================
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /build

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# ============================================================================
# Stage 2: Runtime (Nginx)
# ============================================================================
FROM nginx:alpine

# Copy built files from builder
COPY --from=builder /build/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:80 || exit 1

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
