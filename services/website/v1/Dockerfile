FROM golang:1.13-alpine AS builder

RUN apk --update --no-cache add git

WORKDIR /build

COPY . .

RUN go build -o app

# Next stage
FROM alpine

RUN apk --update --no-cache add ca-certificates openssl

WORKDIR /run

COPY --from=builder /build/app .

# Include website static files
COPY files files/

CMD ["/run/app"]
