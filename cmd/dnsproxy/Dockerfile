FROM golang:1.12.1 AS basebuilder

WORKDIR $GOPATH/src/dig
COPY ./ ./

RUN env GIT_TERMINAL_PROMPT=1 go get ./...
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -o /dnsproxy ./ 


#############################################
FROM ubuntu AS service

COPY --from=basebuilder /dnsproxy /

RUN apt-get -y update
RUN apt-get install -y strace 
RUN apt-get install -y dnsutils 
RUN apt-get install -y netcat

ENTRYPOINT [ "/bin/bash" ]
