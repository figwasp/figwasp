FROM alpine:latest

WORKDIR /root

ARG BINARY_NAME=alduin

COPY test/bin/$BINARY_NAME .

ENV BINARY_NAME=$BINARY_NAME

CMD ["sh", "-c", "./$BINARY_NAME"]
