FROM golang:1.16-alpine

LABEL org.opencontainers.image.source=https://github.com/williamlsh/orchid

RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

WORKDIR /src

COPY go.mod .

RUN go env -w GOPROXY="https://goproxy.io,direct"; \
    go mod download; \
    go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o orchid ./cmd/orchid

FROM scratch

COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /src/orchid /

ENTRYPOINT [ "/orchid" ]
