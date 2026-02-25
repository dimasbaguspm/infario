
FROM golang:1.25.5-alpine AS dev

RUN go install github.com/air-verse/air@v1.64.5
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]


FROM golang:1.25.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/infario ./cmd/app

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/infario /app/infario
EXPOSE 8080
CMD ["/app/infario"]
