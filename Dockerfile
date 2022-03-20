FROM golang:1.17

ARG BINARY_NAME=alduin

ARG SOURCE_PATH=./cmd/$BINARY_NAME

WORKDIR /usr/src/$BINARY_NAME

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

#RUN CGO_ENABLED=0 go test -v $SOURCE_PATH

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /usr/bin/$BINARY_NAME \
    -v \
    $SOURCE_PATH



FROM alpine:latest

WORKDIR /root

COPY --from=0 /usr/bin/$BINARY_NAME .

ARG BINARY_NAME=alduin

ENV BINARY_NAME=$BINARY_NAME

CMD ["sh", "-c", "./$BINARY_NAME"]
