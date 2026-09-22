FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/payment-service/ /app/services/payment-service/
RUN cd /app/services/payment-service && go mod tidy && CGO_ENABLED=0 go build -o /payment-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /payment-service /payment-service
EXPOSE 8084
CMD ["/payment-service"]
