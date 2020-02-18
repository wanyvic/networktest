package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	timeout  time.Duration
	max      time.Duration
	min      time.Duration
	lost     int
	filename = "output.txt"
	f        *os.File
)

func main() {
	var addr string
	var err error
	min = time.Duration(int64(^uint64(0) >> 1))
	timeout = time.Duration(5) * time.Second

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err = os.Create(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	f, err = os.OpenFile(filename, os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(os.Args) > 1 {
		addr = os.Args[1]
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		timeout = time.Duration(i) * time.Second
	}
	for {
		connect(addr)
		lost++
	}
}
func connect(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logrus.Error(err)
	}
	fmt.Println(conn.RemoteAddr())
	defer conn.Close()
	t := time.Now()

	// scanner := bufio.NewScanner(conn)
	// for scanner.Scan() {

	// }

	// fmt.Println("ping: ", time.Since(t))
	subscribe := `{"id": 1, "method": "mining.subscribe", "params": ["bmminer/2.0.0/Antminer S9/13500"]}`

	conn.Write([]byte(subscribe + "\n"))

	// fmt.Println("<--", subscribe)
	scanner := bufio.NewScanner(conn)
	err = conn.SetReadDeadline(time.Now().Add(timeout)) // timeout
	if err != nil {
		fmt.Println("setReadDeadline failed:", err)
	}
	for scanner.Scan() {
		// str := scanner.Text()
		// fmt.Println("-->", str, "ping: ", time.Since(t))
		r := time.Since(t)
		if r < min {
			min = r
		}
		if r > max {
			max = r
		}

		// d1 := fmt.Sprintf("[%s] close min %s max %s timeout %s\n", time.Now().Format("2006-01-02 15:04:05"), min, max, timeout)
		// fmt.Println(d1)
		// _, err := f.Write([]byte(d1))
		// if err != nil {
		// 	fmt.Println(err)
		// }
		fmt.Println("ping: ", r)
		time.Sleep(time.Second)
		t = time.Now()

		conn.Write([]byte(subscribe + "\n"))
		// fmt.Println("<--", subscribe)
		err = conn.SetReadDeadline(time.Now().Add(timeout)) // timeout
		if err != nil {
			fmt.Println("setReadDeadline failed:", err)
		}

	}
	d1 := fmt.Sprintf("[%s] close min %s max %s timeout %s\n", time.Now().Format("2006-01-02 15:04:05"), min, max, timeout)
	fmt.Println("lost connection")
	fmt.Println(d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		fmt.Println(err)
	}
}

// #! /bin/bash
// yourdomain="xxx.xxx.xxx.xxx"
// /root/.acme.sh/acme.sh --days 30 --renew --dns -d ${yourdomain}
// cert_file="/root/.acme.sh/${yourdomain}/${yourdomain}.cer"
// key_file="/root/.acme.sh/${yourdomain}/${yourdomain}.key"
// sudo cp -f $cert_file /usr/local/etc/ipsec.d/certs/server.cert.pem
// sudo cp -f $key_file /usr/local/etc/ipsec.d/private/server.pem
// sudo cp -f $cert_file /usr/local/etc/ipsec.d/certs/client.cert.pem
// sudo cp -f $key_file /usr/local/etc/ipsec.d/private/client.pem
// sudo /usr/local/sbin/ipsec restart
