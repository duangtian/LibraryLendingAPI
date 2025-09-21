FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o library-api ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/library-api /app/library-api
COPY migrations /app/migrations
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/library-api"]
