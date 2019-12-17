package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"net"
	"strings"
)

type DnsHeader struct {
	id         int16
	flags      *HeaderFlags
	qcount     int16
	emptyCunts [3]int16
}

func (dh *DnsHeader) Serialize() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, dh.id)
	buf.Write(dh.flags.Serialize())
	binary.Write(&buf, binary.BigEndian, dh.qcount)
	for _, v := range dh.emptyCunts {
		binary.Write(&buf, binary.BigEndian, v)
	}
	return buf.Bytes()
}

func (dh *DnsHeader) Deserialize(buf []byte) {
	id, _ := binary.Varint(buf[:2])
	dh.id = int16(id)
	fl, _ := binary.Varint(buf[2:4])
	dh.flags.flags = int16(fl)
	c, _ := binary.Varint(buf[4:6])
	dh.qcount = int16(c)
}

func NewDnsHeader() *DnsHeader {
	dh := &DnsHeader{flags: NewHeaderFlags()}
	dh.qcount = 1
	dh.id = 8
	return dh
}

type HeaderFlags struct {
	flags int16
}

func (hf *HeaderFlags) Serialize() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, hf.flags)
	return buf.Bytes()
}

type Query struct {
	domain []string
	qtype  uint16
	qclass uint16
}

//lscpu | grep "Byte Order"
func (q *Query) Serialize() []byte {
	buf := bytes.Buffer{}
	for _, v := range q.domain {
		binary.Write(&buf, binary.BigEndian, byte(len(v)))
		buf.Write([]byte(v))
	}
	binary.Write(&buf, binary.BigEndian, byte(0))
	binary.Write(&buf, binary.BigEndian, q.qtype)
	binary.Write(&buf, binary.BigEndian, q.qclass)
	return buf.Bytes()
}

func (q *Query) Deserialize(buf []byte) int {
	n := 0
	for i := 0; i < len(buf); {
		if buf[i] == 0 {
			n = i + 1
			break
		}
		ln := int(buf[i])
		i++
		q.domain = append(q.domain, string(buf[i:i+ln]))
		i += ln
	}
	t, _ := binary.Varint(buf[n : n+2])
	n += 2
	q.qtype = uint16(t)

	c, _ := binary.Varint(buf[n : n+2])
	n += 2
	q.qclass = uint16(c)
	return n
}

func NewQuery(dom string) *Query {
	q := &Query{domain: strings.Split(dom, "."), qtype: 1, qclass: 1}
	return q
}

type Response struct {
	res []byte
}

func (r *Response) String() string {
	ipBuf := r.res[len(r.res)-4:]
	ip := net.IP(ipBuf)
	return ip.String()
}

func NewDnsRes(res []byte) *DnsQuery {
	dq := &DnsQuery{h: NewDnsHeader(), q: NewQuery("")}
	dq.Deserialize(res)
	return dq
}

type DnsQuery struct {
	h *DnsHeader
	q *Query
	r *Response
}

func (dq *DnsQuery) String() string {
	res := ""
	for _, v := range dq.q.domain {
		res += v + "."
	}
	res += "	" + dq.r.String()
	return res
}

func NewDnsQuery(dom string) *DnsQuery {
	return &DnsQuery{h: NewDnsHeader(), q: NewQuery(dom)}
}

func (dq *DnsQuery) Serialize() []byte {
	buf := bytes.Buffer{}
	buf.Write(dq.h.Serialize())
	buf.Write(dq.q.Serialize())
	return buf.Bytes()
}

func (dq *DnsQuery) Deserialize(buf []byte) {
	log.Printf("Len res %d\n", len(buf))
	dq.h.Deserialize(buf[:12])
	buf = buf[12:]
	n := dq.q.Deserialize(buf[:])
	buf = buf[n:]
	dq.r = &Response{res: buf}
}

func NewHeaderFlags() *HeaderFlags {
	hf := &HeaderFlags{}
	hf.flags |= 0x01
	hf.flags <<= 8
	return hf
}

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
	dq := NewDnsQuery(*doamin)
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

	r := NewDnsRes(res[:n])

	//https://osqa-ask.wireshark.org/questions/50806/help-understanding-dns-packet-data
	log.Printf("res part % x\n", r.r.res)

	log.Println(r)

}
