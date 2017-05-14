FROM golang:1.8

ENV GOPATH /go
ENV PATH "$GOPATH/bin:$PATH"

COPY . /go/src/github.com/leopoldodonnell/demo-controller
WORKDIR /go/src/github.com/leopoldodonnell/demo-controller
RUN go install

# Clean up!
WORKDIR /go
RUN rm -rf /go/src/
ENTRYPOINT [ "demo-controller" ]
