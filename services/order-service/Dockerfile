FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/order-service/ /app/services/order-service/
RUN cd /app/services/order-service && go mod tidy && CGO_ENABLED=0 go build -o /order-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /order-service /order-service
EXPOSE 8083
CMD ["/order-service"]
