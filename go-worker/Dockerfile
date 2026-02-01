FROM golang:1.23-alpine AS modules
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


FROM modules AS builder
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o worker ./cmd/worker/main.go


FROM scratch

COPY --from=builder /app/worker /worker

USER 1000:1000

ENTRYPOINT ["/worker"]
