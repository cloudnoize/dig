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
	flag.Parse()
	sport := strconv.Itoa(*localport)

	add, err := net.ResolveUDPAddr("udp", ":"+sport)
	if err != nil {
		log.Fatalf(err.Error())
	}
	conn, err := net.ListenUDP("udp", add)
	if err != nil {
		log.Fatalf(err.Error())
	}
	conn.SetReadBuffer(2046)
	defer conn.Close()
	for {
		select {
		case <-done:
			log.Println("Exiting")
			return
		default:
			var buf [2046]byte
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			_, remaddr, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				continue
			}
			go func() {
				log.Println("Writing to ", remaddr.String())
				conn.WriteTo(buf[:], remaddr)
			}()
		}
	}

}
