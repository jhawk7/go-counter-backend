FROM golang:1.22.1-alpine3.19 AS builder
WORKDIR /go/app
COPY . ./
RUN go mod download
RUN mkdir bin
RUN cd cmd/api/ && go build -o ../../bin/api

FROM golang:1.22.1-alpine3.19
WORKDIR /go
COPY --from=builder /go/app/bin/api .
CMD ["./api"]