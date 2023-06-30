# service image

FROM golang:1.20 AS builder

ARG SERVICE

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /src
COPY . .

RUN go build \
  -trimpath \
  -o /bin/service \
  ./cmd/${SERVICE}


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/service /bin/service

ENV PORT 8080

ENTRYPOINT ["/bin/service"]