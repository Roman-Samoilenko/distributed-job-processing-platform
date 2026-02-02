FROM golang:1.25-alpine AS modules
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


FROM modules AS builder
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o worker ./cmd/worker/main.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/worker /worker

COPY .env /.env 

USER 1000:1000

ENTRYPOINT ["/worker"]
