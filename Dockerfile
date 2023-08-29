FROM golang:1.20-alpine AS builder
COPY . /TeleBotNotifications/
WORKDIR /TeleBotNotifications/
RUN go build -o ./bin/bot main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /TeleBotNotifications/bin/bot .
EXPOSE 8888
CMD ["./bot"]