FROM golang:1.15 AS build
WORKDIR /capitalizone
ADD ../../../../ .
ARG TAG
ARG REPO_INFO
ARG COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build pkg/caclient/examples/server/server.go -mod vendor -o server -ldflags "-s -w -X main.RELEASE=$TAG -X main.COMMIT=$COMMIT -X main.REPO=$REPO_INFO"

# Final stage: the running container.
FROM alpine:3.11
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

# Create the user (ID 1000) and group that will be used in the
# running container to run the process as an unprivileged user.
RUN addgroup -S nerv && \
    adduser -S nerv -G nerv -u 1000

WORKDIR /capitalizone

# Import the compiled executable from the second stage.
COPY --from=build /capitalizone/server /bin
# Only for backwards compatibility
COPY --from=build /capitalizone/ .

RUN ls -al

# Perform any further action as an unprivileged user.
USER 1000

# Run the compiled binary.
CMD ["server"]