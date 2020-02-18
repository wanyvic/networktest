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
	d1 := fmt.Sprintf("\n***************\n[%s] net test started\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		fmt.Println(err)
	}
	for {
		lost++
		connect(addr)
	}
}
func connect(addr string) {

	min = time.Duration(int64(^uint64(0) >> 1))
	max = 0
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logrus.Error(err)
	}
	d1 := fmt.Sprintf("[%s] new connection %s \n", time.Now().Format("2006-01-02 15:04:05"), conn.RemoteAddr().String())
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		fmt.Println(err)
	}
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
			d1 := fmt.Sprintf("[%s] new min ping: %s\n", time.Now().Format("2006-01-02 15:04:05"), min)
			fmt.Printf("%s", d1)
			_, err = f.Write([]byte(d1))
			if err != nil {
				fmt.Println(err)
			}
		}
		if r > max {
			max = r
			d1 := fmt.Sprintf("[%s] new max ping: %s\n", time.Now().Format("2006-01-02 15:04:05"), max)
			fmt.Printf("%s", d1)
			_, err = f.Write([]byte(d1))
			if err != nil {
				fmt.Println(err)
			}
		}

		// d1 := fmt.Sprintf("[%s] close min %s max %s timeout %s\n", time.Now().Format("2006-01-02 15:04:05"), min, max, timeout)
		// fmt.Printf("%s",d1)
		// _, err := f.Write([]byte(d1))
		// if err != nil {
		// 	fmt.Println(err)
		// }
		fmt.Printf("[%s] ping: %s\n", time.Now().Format("2006-01-02 15:04:05"), r.String())
		time.Sleep(time.Second)
		t = time.Now()

		conn.Write([]byte(subscribe + "\n"))
		// fmt.Println("<--", subscribe)
		err = conn.SetReadDeadline(time.Now().Add(timeout)) // timeout
		if err != nil {
			fmt.Println("setReadDeadline failed:", err)
		}

	}

	r := time.Since(t)
	d1 = fmt.Sprintf("[%s] close index %d min %s max %s ReadDeadline set timeout %s time use %s\n", time.Now().Format("2006-01-02 15:04:05"), lost, min, max, timeout, r)
	fmt.Println("lost connection")
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		fmt.Println(err)
	}
}
