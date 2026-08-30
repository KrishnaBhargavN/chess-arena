# 1) build the frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 2) build the backend as a static binary
FROM golang:1.25-alpine AS backend
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /chess-server ./cmd/server

# 3) minimal runtime image serving both the API and the built frontend
FROM alpine:latest
WORKDIR /app
COPY --from=backend /chess-server /app/chess-server
COPY --from=frontend /app/frontend/dist /app/frontend/dist
EXPOSE 8080
CMD ["/app/chess-server"]
