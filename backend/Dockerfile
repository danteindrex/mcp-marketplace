# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /out/mcp-local-bridge.exe ./cmd/local-bridge
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /out/mcp-local-bridge-macos ./cmd/local-bridge

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/mcp-local-bridge.exe /app/bin/mcp-local-bridge.exe
COPY --from=build /out/mcp-local-bridge-macos /app/bin/mcp-local-bridge-macos
EXPOSE 2024
ENV PORT=2024
ENTRYPOINT ["/app/server"]
