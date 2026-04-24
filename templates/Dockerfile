FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ARG GAME_ID
RUN go build -trimpath -ldflags="-s -w" -o /out/game-${GAME_ID} ./cmd/game-${GAME_ID}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/game-* /usr/local/bin/
ENV GRPC_PORT=9201
EXPOSE 9201
ARG GAME_ID
ENV GAME_BINARY=/usr/local/bin/game-${GAME_ID}
ENTRYPOINT ["/bin/sh", "-c", "exec $GAME_BINARY"]
