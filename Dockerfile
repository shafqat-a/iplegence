# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/build ./cmd/build \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/lookup ./cmd/lookup \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/validate ./cmd/validate \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/serve ./cmd/serve

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 65532 iplegence
COPY --from=build /out/build /out/lookup /out/validate /out/serve /usr/local/bin/
COPY configs /configs
COPY testdata/golden.json /testdata/golden.json
COPY dist/Superior-IP.mmdb /data/Superior-IP.mmdb
ENV MMDB_PATH=/data/Superior-IP.mmdb \
    LISTEN_ADDR=:8080
WORKDIR /
USER iplegence
EXPOSE 8080
ENTRYPOINT ["serve"]
CMD ["-addr", ":8080"]
