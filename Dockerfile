FROM golang:alpine3.7 as builder

ENV GOPATH /go
ENV PATH "$GOPATH/bin:$PATH"

RUN apk --no-cache add git && \
  go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/leopoldodonnell/demo-kubernetes-controller
COPY . .
RUN dep init && dep ensure
RUN go install

FROM alpine:3.7

WORKDIR /share
COPY --from=builder /go/bin/demo-kubernetes-controller /usr/local/bin/demo-controller

ENTRYPOINT [ "/usr/local/bin/demo-controller" ]
