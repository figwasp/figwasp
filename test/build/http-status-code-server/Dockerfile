# build context assumed to be at the root of the Go module

FROM golang:1.17

ARG BINARY_NAME=http-status-code-server
ARG SOURCE_PATH=./test/cmd/$BINARY_NAME

WORKDIR /usr/src/$BINARY_NAME

COPY . .

#RUN CGO_ENABLED=0 go test -v $SOURCE_PATH

ARG SERVER_PORT="3000"
ARG STATUS_CODE="501"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /usr/bin/$BINARY_NAME \
    -ldflags " \
        -X 'main.serverPortString=$SERVER_PORT' \
        -X 'main.statusCodeString=$STATUS_CODE' " \
    $SOURCE_PATH



FROM alpine:latest

WORKDIR /root

COPY --from=0 /usr/bin/$BINARY_NAME .

ARG BINARY_NAME=http-status-code-server

ENV BINARY_NAME=$BINARY_NAME

CMD ["sh", "-c", "./$BINARY_NAME"]
