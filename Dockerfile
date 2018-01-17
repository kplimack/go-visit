FROM golang:1.9.2 as builder

WORKDIR /go/src/github.com/partkyle/go-visit
COPY . /go/src/github.com/partkyle/go-visit

RUN go build -o visit

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/partkyle/go-visit/visit .
ENTRYPOINT ["./visit"]