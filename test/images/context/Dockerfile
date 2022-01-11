FROM golang:1.17

ARG APP_NAME=app

WORKDIR /usr/src/$APP_NAME

COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go test -v ./...

ARG APP_MESSAGE
ARG APP_PATH="/"
ARG APP_PORT=":8000"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /usr/bin/$APP_NAME \
    -ldflags " \
        -X 'main.message=$APP_MESSAGE' \
        -X 'main.path=$APP_PATH' \
        -X 'main.port=$APP_PORT' \
    "



FROM alpine:latest

WORKDIR /root

COPY --from=0 /usr/bin/$APP_NAME .

ARG APP_NAME=app

ENV APP_NAME=$APP_NAME

CMD ["sh", "-c", "./$APP_NAME"]
