FROM golang:1.22.1-alpine3.19 AS builder
WORKDIR /go/consumer
COPY ./counter-consumer ./
COPY ./shared ../shared
RUN go mod download
RUN mkdir bin
RUN cd cmd/consumer/ && go build -o ../../bin/consumer

FROM golang:1.22.1-alpine3.19
WORKDIR /go
COPY --from=builder /go/consumer/bin/consumer .
CMD ["./consumer"]