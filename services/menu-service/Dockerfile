FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/menu-service/ /app/services/menu-service/
RUN cd /app/services/menu-service && go mod tidy && CGO_ENABLED=0 go build -o /menu-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /menu-service /menu-service
EXPOSE 8082
CMD ["/menu-service"]
