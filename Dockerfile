FROM debian:stable-slim

# COPY source destination
COPY go-server /go-server

CMD ["./go-server"]

