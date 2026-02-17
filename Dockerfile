# Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.24-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o api ./cmd/api
RUN CGO_ENABLED=1 go build -o addrepo ./cmd/addrepo

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates git openssh-client
WORKDIR /app
COPY --from=backend /app/api .
COPY --from=backend /app/addrepo .
COPY --from=frontend /app/web/dist ./web/dist
EXPOSE 3000
CMD ["./api"]
