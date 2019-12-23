package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cloudnoize/dig/dnsmsg"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	localport := flag.Int("p", 53, "local port to listen to")
	dnsserver := flag.String("d", "8.8.8.8:53", "remote dns")
	ttl := flag.Uint("t", 0, "new ttl")
	flag.Parse()

	sport := strconv.Itoa(*localport)

	add, err := net.ResolveUDPAddr("udp", ":"+sport)
	if err != nil {
		log.Fatalf(err.Error())
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", *dnsserver)
	if err != nil {
		log.Fatalf(err.Error())
	}

	go HandleUDP(add, remoteAddr, *ttl, done)

}

func HandleUDP(add, remoteAddr *net.UDPAddr, ttl uint, done chan bool) {
	conn, err := net.ListenUDP("udp", add)
	if err != nil {
		log.Fatalf(err.Error())
	}
	conn.SetReadBuffer(2046)
	defer conn.Close()

	remConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer remConn.Close()

	for {
		select {
		case <-done:
			log.Println("Exiting")
			return
		default:
			var buf [2046]byte
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, claddr, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				continue
			}
			go func() {
				dq := dnsmsg.NewDnsQuery("")
				remConn.Write(buf[:n])
				remConn.ReadFromUDP(buf[:])
				dq.Deserialize(buf[:])

				if ttl != 0 {
					dq.SetTTL(uint32(ttl))
				}
				conn.WriteTo(buf[:], claddr)
			}()
		}
	}
}
