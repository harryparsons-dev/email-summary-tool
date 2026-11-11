FROM golang:1.26.5-alpine AS development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "tool", "CompileDaemon", "-build=go build -buildvcs=false -o /tmp/mailai ./main", "-command=/tmp/mailai", "-pattern=(.+\\.go|.+\\.sql)$", "-exclude-dir=.git", "-exclude-dir=frontend", "-exclude-dir=tmp", "-polling", "-polling-interval=200", "-graceful-kill"]

FROM development AS build

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /mailai ./main

FROM alpine:3.23

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=build /mailai /usr/local/bin/mailai

USER app
EXPOSE 8080 2525

ENTRYPOINT ["mailai"]
