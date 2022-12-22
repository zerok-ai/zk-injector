FROM golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add make

WORKDIR /go/src/zk-injector
ADD . .
RUN make build

FROM alpine:3.17
WORKDIR /zk-injector
COPY --from=build /go/src/zk-injector/zk-injector .
CMD ["/zk-injector/zk-injector"]