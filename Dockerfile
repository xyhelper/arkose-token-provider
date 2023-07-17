FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o arkose-token-provider main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/arkose-token-provider .
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
EXPOSE 8080
CMD ["/app/arkose-token-provider"]
