FROM golang:1.12

WORKDIR $GOPATH/src/go.ryanbrainard.com/timesink
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
