FROM alpine as cert
COPY cert /cert

RUN apk update && apk add openssl

WORKDIR /cert
RUN ./gen.sh

FROM golang:1.20 as build
COPY . /src

WORKDIR /src/cmd/server/
RUN CGO_ENABLED=0 go build -o /build/gophkeeper-server .

FROM scratch
COPY --from=cert /cert/server-cert.pem /
COPY --from=cert /cert/server-key.pem /
COPY --from=build /build/* /
ENTRYPOINT ["/gophkeeper-server", "-c", "server-cert.pem", "-k", "server-key.pem"]
