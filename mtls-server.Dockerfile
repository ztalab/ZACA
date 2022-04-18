FROM golang:1.15 AS build
WORKDIR /capitalizone
ADD . .
CMD ["go", "run", "pkg/caclient/examples/server/server.go"]