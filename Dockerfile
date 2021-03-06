FROM golang:1-alpine as builder

ADD . /src

WORKDIR /src

ENV GO111MODULE=on

RUN cd cmd/screenshot && go build

FROM alpine:latest

WORKDIR /bin/

COPY --from=builder /src/cmd/screenshot/screenshot .

ENTRYPOINT ["/bin/screenshot"]
