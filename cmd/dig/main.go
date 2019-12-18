package main

import (
	"flag"
	"log"
	"net"

	"github.com/cloudnoize/udemy/dig/dnsmsg"
)

func main() {
	doamin := flag.String("d", "google.com", "domain")
	flag.Parse()

	udpaddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")

	if err != nil {
		log.Fatal(err)
	}

	/*
		socket(AF_INET, SOCK_DGRAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3 - create sokcet datagram return file desc
		setsockopt(3, SOL_SOCKET, SO_BROADCAST, [1], 4) = 0 - set options on socket
		connect(3, {sa_family=AF_INET, sin_port=htons(53), sin_addr=inet_addr("8.8.8.8")}, 16) = 0 - only binds the addr, does not make any conneciton.
		epoll_create1(EPOLL_CLOEXEC)            = 4
		epoll_ctl(4, EPOLL_CTL_ADD, 3, {EPOLLIN|EPOLLOUT|EPOLLRDHUP|EPOLLET, {u32=4208557832, u64=140423869333256}}) = 0
		getsockname(3, {sa_family=AF_INET, sin_port=htons(40335), sin_addr=inet_addr("172.17.0.3")}, [112->16]) = 0
		getpeername(3, {sa_family=AF_INET, sin_port=htons(53), sin_addr=inet_addr("8.8.8.8")}, [112->16]) = 0
	*/
	c, err := net.DialUDP("udp", nil, udpaddr)

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	dq := dnsmsg.NewDnsQuery(*doamin)
	buf := dq.Serialize()

	log.Printf("% x\n", buf)

	n, err := c.Write(buf)

	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("Wrote %d bytes\n", n)

	var res [2056]byte
	n, err = c.Read(res[:])

	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("% x\n", res[:n])

	r := dnsmsg.NewDnsRes(res[:n])

	//https://osqa-ask.wireshark.org/questions/50806/help-understanding-dns-packet-data

	log.Println(r)

}
