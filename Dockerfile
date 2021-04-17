FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go test -c crawler/*.go -o crawler_lib_test
RUN cd cmd && go build -o crawler 

CMD ["sh", "-c", "./crawl.sh"]