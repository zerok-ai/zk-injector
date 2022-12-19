FROM golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

WORKDIR /go/src/github.com/zerok-ai/zk-injector
ADD . .
RUN make app

FROM alpine
WORKDIR /app
COPY --from=build /go/src/github.com/zerok-ai/zk-injector/zk-injector .
CMD ["/app/zk-injector"]