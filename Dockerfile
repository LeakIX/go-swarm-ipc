FROM golang
ADD . /go/src/github.com/LeakIX/go-swarm-ipc
WORKDIR /go/src/github.com/LeakIX/go-swarm-ipc
RUN go build -o /test ./cmd/test/
CMD /test
