FROM golang:1.26.10-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app
ENV APP_ENV=production GIN_MODE=release
COPY --from=build /out/api ./api
COPY migrations ./migrations
USER app
EXPOSE 8080
CMD ["./api"]
