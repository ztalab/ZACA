.PHONY: all build clean

GOPROXY=https://goproxy.oneitfarm.com,https://goproxy.cn,direct
PROG=bin/capitalizone
SRCS=.

# git commit hash
COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
# 编译日期
BUILD_DATE=$(shell date '+%Y-%m-%d %H:%M:%S')
# 编译条件
CFLAGS = -ldflags "-s -w -X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=$(BUILD_DATE)\""

all:
	if [ ! -d "./bin/" ]; then \
	mkdir bin; \
	fi
	GOPROXY=$(GOPROXY) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  $(CFLAGS) -o $(PROG) $(SRCS)

build:
	go build -race -tags=jsoniter

swagger:
	swag init

compose:
	sudo docker-compose up -d

run:
	go run main.go

test:
	go run main.go -env test

rootca:
	go run main.go -env test -envfile ".env.rootca" -rootca

fake:
	go run test/fake/fake_server.go -env test -ca https://192.168.2.80:8381

cfssl-model:
	gen --sqltype=mysql -c "root:123456@tcp(192.168.2.80:3306)/cap?charset=utf8mb4&parseTime=True&loc=Local" -d cap --json --generate-dao --overwrite --gorm --db --module "gitlab.oneitfarm.com/bifrost/capitalizone/examples/cfssl-model" --out ./examples/cfssl-model

telegraf:
	sudo docker run --network=host -v `pwd`/telegraf.conf:/telegraf.conf --rm -it telegraf:1.19.0 telegraf --config /telegraf.conf

migration:
	go run main.go -envfile ".env.prod"

clean:
	rm -rf ./bin