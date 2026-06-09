# syntax=docker/dockerfile:1.7

# ── Build stage ──
FROM node:20-alpine AS build

WORKDIR /app

# Cache npm deps separately from source so source-only edits skip the install.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

# ── Runtime stage ── tiny nginx serving the built SPA.
FROM nginx:1.27-alpine AS runtime

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
