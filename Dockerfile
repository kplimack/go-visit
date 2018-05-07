FROM golang:1.9.2 as builder

WORKDIR /go/src/github.com/partkyle/go-visit
COPY . .

RUN make

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/partkyle/go-visit/build/visit .
ENTRYPOINT ["./visit"]
