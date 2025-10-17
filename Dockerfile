# Build frontend
FROM node:18 AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build Go backend
FROM golang:1.21 AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-build /app/frontend/build/index.html ./static/
COPY --from=frontend-build /app/frontend/build/static/ ./static/
RUN go build -o turingroulette ./cmd/server

# Final image
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend-build /app/turingroulette .
COPY --from=backend-build /app/static ./static
COPY --from=backend-build /app/data ./data
EXPOSE 8080
CMD ["./turingroulette"]
