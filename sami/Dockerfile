FROM golang:1.18-alpine

RUN export GO111MODULE=on
ENV APP_NAME sami
ENV CMD_PATH main.go
COPY ./sami $GOPATH/src/$APP_NAME
WORKDIR $GOPATH/src/$APP_NAME

RUN go build

CMD ./$APP_NAME