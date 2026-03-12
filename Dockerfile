FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gvb-server .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/gvb-server /app/gvb-server
COPY settings.docker.yaml /app/settings.yaml
EXPOSE 8080
CMD ["/app/gvb-server"]
