FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/shared/ /app/services/shared/
COPY services/auth-service/ /app/services/auth-service/
RUN cd /app/services/auth-service && go mod tidy && CGO_ENABLED=0 go build -o /auth-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /auth-service /auth-service
EXPOSE 8081
CMD ["/auth-service"]
