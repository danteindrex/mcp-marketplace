FROM node:22-alpine
WORKDIR /app
COPY webapp/package.json webapp/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY webapp/ ./
RUN pnpm build
EXPOSE 3000
CMD ["pnpm", "start"]