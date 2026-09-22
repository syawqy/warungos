FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/api-gateway/ /app/services/api-gateway/
RUN cd /app/services/api-gateway && go mod tidy && CGO_ENABLED=0 go build -o /api-gateway .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
CMD ["/api-gateway"]
