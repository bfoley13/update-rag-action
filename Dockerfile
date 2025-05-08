FROM golang:1.23 AS builder

WORKDIR /src/app
COPY . .

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /tmp/rag-updater

FROM alpine:3.21 AS run-env

COPY --from=builder /tmp/rag-updater /rag-updater

CMD [ "/rag-updater" ]