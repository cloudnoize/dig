FROM golang:1.12.1 AS basebuilder

WORKDIR $GOPATH/src/dig
COPY ./ ./

RUN go get ./...
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -o /mdig ./ 


#############################################
FROM ubuntu AS service

COPY --from=basebuilder /mdig /

RUN apt-get -y update
RUN apt-get install -y strace 
RUN apt-get install -y dnsutils 

ENTRYPOINT [ "/bin/bash" ]
