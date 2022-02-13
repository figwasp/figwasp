FROM golang:1.17

ARG BINARY_NAME=http-status-code-server

WORKDIR /usr/src/$BINARY_NAME

COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go test -v ./...

ARG STATUS_CODE="501"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /usr/bin/$BINARY_NAME \
    -ldflags "-X 'main.statusCodeString=$STATUS_CODE'"



FROM alpine:latest

WORKDIR /root

COPY --from=0 /usr/bin/$BINARY_NAME .

ARG BINARY_NAME=http-status-code-server

ENV BINARY_NAME=$BINARY_NAME

CMD ["sh", "-c", "./$BINARY_NAME"]
