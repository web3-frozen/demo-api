# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /inventory ./cmd/inventory && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pricing ./cmd/pricing

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
COPY --from=builder /inventory /inventory
COPY --from=builder /pricing /pricing
USER nonroot:nonroot
ENTRYPOINT ["/server"]
