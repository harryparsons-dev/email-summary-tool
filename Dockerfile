FROM golang:1.26.5-alpine AS development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "tool", "air", "-c", ".air.toml"]

FROM development AS build

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /email-summary-tool ./main

FROM alpine:3.23

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=build /email-summary-tool /usr/local/bin/email-summary-tool

USER app
EXPOSE 8080

ENTRYPOINT ["email-summary-tool"]
