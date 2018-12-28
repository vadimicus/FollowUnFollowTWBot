FROM golang:1.9.2


RUN go get github.com/vadimicus/FollowUnFollowTWBot && \
    cd $GOPATH/src/github.com/vadimicus/FollowUnFollowTWBot && \
    git pull && \
    go build main.go



WORKDIR $GOPATH/src/github.com/vadimicus/FollowUnFollowTWBot

RUN echo "VERSION 0.1.1"



ENTRYPOINT ./main
