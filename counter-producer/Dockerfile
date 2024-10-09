FROM golang:1.22.1-alpine3.19 AS builder
WORKDIR /go/producer
COPY ./counter-producer ./
COPY ./shared ../shared
RUN go mod download
RUN mkdir bin
RUN cd cmd/producer/ && go build -o ../../bin/producer

FROM golang:1.22.1-alpine3.19
WORKDIR /go
COPY --from=builder /go/producer/bin/producer .
CMD ["./producer"]