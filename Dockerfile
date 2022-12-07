FROM golang:1.14-buster AS builder
COPY ./ /go/src/github.com/nighttsarina/homeplug_exporter/
WORKDIR /go/src/github.com/nighttsarina/homeplug_exporter/
RUN make

FROM scratch
COPY --from=builder /go/src/github.com/nighttsarina/homeplug_exporter /go/bin/
ENTRYPOINT ["/go/bin/homeplug_exporter"]
