FROM golang:1.20-alpine AS builder

WORKDIR /build
COPY . .

RUN go build cmd/minecraftrouter.go \
    && mkdir /out \
    && mv minecraftrouter /out


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/minecraftrouter /usr/local/bin/

EXPOSE 25565

ENTRYPOINT ["minecraftrouter"]