FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/yatori-web .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/yatori-web /app/yatori-web
EXPOSE 8080
ENTRYPOINT ["/app/yatori-web"]
