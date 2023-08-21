# Build go
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -v -o Aiko-Server -trimpath -ldflags "-s -w -buildid=" ./Aiko-Server

# Release
FROM  alpine
RUN  apk --update --no-cache add tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN mkdir /etc/Aiko-Server/
COPY --from=builder /app/Aiko-Server /usr/local/bin

ENTRYPOINT [ "Aiko-Server", "--config", "/etc/Aiko-Server/aiko.yml"]