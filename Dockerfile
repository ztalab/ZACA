FROM hub.oneitfarm.com/library/golang:1.17.8-alpine AS builder

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.oneitfarm.com,https://goproxy.cn,direct

WORKDIR /build

COPY . .
RUN CGO_ENABLED=0 go build -o capitalizone .

FROM harbor.oneitfarm.com/bifrost/ubuntu:20.04

WORKDIR /capitalizone

COPY --from=builder /build/capitalizone .
COPY --from=builder /build/database/mysql/migrations ./database/mysql/migrations
COPY --from=builder /build/conf.default.yml .
COPY --from=builder /build/conf.prod.yml .
COPY --from=builder /build/conf.test.yml .
RUN chmod +x capitalizone

CMD ["./capitalizone", "http"]