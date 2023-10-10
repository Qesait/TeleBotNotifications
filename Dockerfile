FROM golang:1.20-alpine AS builder
COPY . /TeleBotNotifications/
WORKDIR /TeleBotNotifications/
RUN go build -o ./.bin/bot cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /TeleBotNotifications/.bin/bot .
COPY --from=0 /TeleBotNotifications/entrypoint.sh .
COPY --from=0 /TeleBotNotifications/configs configs/

ENV WORKING_DIRECTORY=/var/lib/spotify_notifications_bot
EXPOSE 8888

ENTRYPOINT ["/root/entrypoint.sh"]