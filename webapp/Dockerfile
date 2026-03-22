FROM node:22-alpine
WORKDIR /app
ARG NEXT_PUBLIC_API_BASE_URL
ARG API_BASE_URL
ARG NEXT_PUBLIC_N8N_URL
ARG SITE_URL
ARG NEXT_PUBLIC_SITE_URL
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV API_BASE_URL=$API_BASE_URL
ENV NEXT_PUBLIC_N8N_URL=$NEXT_PUBLIC_N8N_URL
ENV SITE_URL=$SITE_URL
ENV NEXT_PUBLIC_SITE_URL=$NEXT_PUBLIC_SITE_URL
COPY webapp/package.json webapp/package-lock.json ./
RUN npm ci
COPY webapp/app ./app
COPY webapp/components ./components
COPY webapp/hooks ./hooks
COPY webapp/lib ./lib
COPY webapp/public ./public
COPY webapp/styles ./styles
COPY webapp/components.json ./components.json
COPY webapp/eslint.config.mjs ./eslint.config.mjs
COPY webapp/next-env.d.ts ./next-env.d.ts
COPY webapp/next.config.mjs ./next.config.mjs
COPY webapp/postcss.config.mjs ./postcss.config.mjs
COPY webapp/proxy.ts ./proxy.ts
COPY webapp/tsconfig.json ./tsconfig.json
RUN npm run build
EXPOSE 3000
CMD ["npm", "run", "start"]
