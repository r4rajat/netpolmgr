FROM golang:1.18.3-alpine

WORKDIR /

COPY . .

RUN go build

ENTRYPOINT ["./netpolmgr"]