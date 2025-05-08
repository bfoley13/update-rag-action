FROM golang:1.23 AS builder

WORKDIR /src
COPY . .

ENV GO111MODULE=on

RUN go build -o rag-updater

FROM scratch

COPY --from=builder /src/rag-updater /bin/rag-updater

ENTRYPOINT ["/bin/rag-updater"]