FROM golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add make

WORKDIR /go/src/zerok-injector
ADD . .
RUN make build

FROM alpine:3.17
WORKDIR /zerok-injector
COPY --from=build /go/src/zerok-injector/zerok-injector .
CMD ["/zerok-injector/zerok-injector"]