
FROM golang:1.25.5-alpine AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run air for hot reload
CMD ["air", "-c", ".air.toml"]


# Production stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/infario ./cmd/app


FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/infario /app/infario

EXPOSE 8080

CMD ["/app/infario"]
