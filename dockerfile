# Build go
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -v -o build_assets/Aiko-Server -trimpath -ldflags "-X 'github.com/AikoPanel/Aiko-Server/cmd.version=$version' -s -w -buildid="

# Release
FROM alpine
RUN apk --update --no-cache add tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime
RUN mkdir /etc/Aiko-Server/
COPY --from=builder /app/Aiko-Server /usr/local/bin

# Thiết lập CMD để chạy script và sử dụng biến môi trường
CMD ["Aiko-Server", "certificate", "--domain", "$DOMAIN", "--expire", "$EXPIRE", "&&", "Aiko-Server", "server", "--format", "$FORMAT"]
