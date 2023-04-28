FROM --platform=linux/amd64 golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add make

WORKDIR /go/src/zerok-injector
ADD . .
RUN make build

FROM --platform=linux/amd64 alpine:3.17
WORKDIR /zerok-injector
COPY --from=build /go/src/zerok-injector/zerok-injector .
COPY internal/config/config.yaml /zerok-injector/
CMD ["/zerok-injector/zerok-injector", "-c", "/zerok-injector/config.yaml"]