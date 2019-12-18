package dnsmsg

import (
	"bytes"
	"encoding/binary"
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
