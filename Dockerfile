FROM --platform=$BUILDPLATFORM golang:alpine as builder

ARG BUILDPLATFORM
ARG TARGETARCH
ARG TARGETOS
ENV GOARCH=${TARGETARCH} GOOS=${TARGETOS}

WORKDIR /go/src/github.com/wish/mongoproxy

COPY . /go/src/github.com/wish/mongoproxy
RUN cd /go/src/github.com/wish/mongoproxy/cmd/mongoproxy && CGO_ENABLED=0 go build -mod=vendor

FROM golang:alpine

COPY --from=builder /go/src/github.com/wish/mongoproxy/cmd/mongoproxy/mongoproxy /bin/mongoproxy

USER nobody

ENTRYPOINT [ "/bin/mongoproxy" ]

