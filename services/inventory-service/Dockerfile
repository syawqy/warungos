FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/inventory-service/ /app/services/inventory-service/
RUN cd /app/services/inventory-service && go mod tidy && CGO_ENABLED=0 go build -o /inventory-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /inventory-service /inventory-service
EXPOSE 8085
CMD ["/inventory-service"]
