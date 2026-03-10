FROM node:22-alpine
WORKDIR /app
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:2024
ARG API_BASE_URL=http://backend:2024
ARG NEXT_PUBLIC_N8N_URL=http://localhost:5678
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV API_BASE_URL=$API_BASE_URL
ENV NEXT_PUBLIC_N8N_URL=$NEXT_PUBLIC_N8N_URL
COPY webapp/package.json webapp/package-lock.json ./
RUN npm ci
COPY webapp/ ./
RUN npm run build
EXPOSE 3000
CMD ["npm", "run", "start"]
