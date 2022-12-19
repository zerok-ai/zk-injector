FROM golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add make

WORKDIR /go/src/zk-injector
ADD . .
RUN make app

FROM alpine
WORKDIR /app
COPY --from=build /go/src/zk-injector/zk-injector .
CMD ["/app/zk-injector"]