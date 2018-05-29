# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/halink0803/telegram-task-manager

WORKDIR /go/src/github.com/halink0803/telegram-task-manager
RUN go install -v github.com/halink0803/telegram-task-manager

EXPOSE 8888