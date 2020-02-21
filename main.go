package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	timeout                 time.Duration
	max                     time.Duration
	min                     time.Duration
	lost                    int
	filename                = "output.txt"
	f                       *os.File
	sendExtranonceSubscribe bool
)

func main() {
	var addr string
	var err error
	timeout = time.Duration(5) * time.Second

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err = os.Create(filename)
		if err != nil {
			logrus.Error(err)
			return
		}
	}

	f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		logrus.Error(err)
		return
	}
	if len(os.Args) > 2 {
		addr = os.Args[1]
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			logrus.Error(err)
			return
		}
		timeout = time.Duration(i) * time.Second
		b, err := strconv.ParseBool(os.Args[3])
		if err != nil {
			logrus.Error(err)
			return
		}
		sendExtranonceSubscribe = b
	}
	d1 := fmt.Sprintf("\r\n***************\r\n[%s] net test started\r\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		logrus.Error(err)
	}
	for {
		lost++
		connect(addr)
		time.Sleep(time.Millisecond * time.Duration(100))
	}
}
func connect(addr string) {

	min = time.Duration(int64(^uint64(0) >> 1))
	max = 0
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		_, err = f.Write([]byte(err.Error() + "\r\n"))
		if err != nil {
			logrus.Error(err)
		}
		logrus.Error(err)
		return
	}
	d1 := fmt.Sprintf("[%s] new connection %s \r\n", time.Now().Format("2006-01-02 15:04:05"), conn.RemoteAddr().String())
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		logrus.Error(err)
	}
	defer conn.Close()
	t := time.Now()

	subscribe := `{"id": 7, "method": "mining.subscribe", "params": ["bmminer/2.0.0/Antminer S9/13500"]}`
	authorize := `{"id": 2, "method": "mining.authorize", "params": ["myAzQj4bH4mMF2GpoLSY2v4qVquASTpzR4", "x"]}`
	conn.Write([]byte(subscribe + "\n"))
	conn.Write([]byte(authorize + "\n"))

	scanner := bufio.NewScanner(conn)
	err = conn.SetReadDeadline(time.Now().Add(timeout)) // timeout
	if err != nil {
		logrus.Error("setReadDeadline failed:", err)
	}
	for scanner.Scan() {
		str := scanner.Text()
		// fmt.Println("-->", str)
		if !strings.Contains(str, `"id":7`) {
			continue
		}
		r := time.Since(t)
		if r < min {
			min = r
			d1 := fmt.Sprintf("[%s] new min ping: %s\r\n", time.Now().Format("2006-01-02 15:04:05"), min)
			fmt.Printf("%s", d1)
			_, err = f.Write([]byte(d1))
			if err != nil {
				logrus.Error(err)
			}
		}
		if r > max {
			max = r
			d1 := fmt.Sprintf("[%s] new max ping: %s\r\n", time.Now().Format("2006-01-02 15:04:05"), max)
			fmt.Printf("%s", d1)
			_, err = f.Write([]byte(d1))
			if err != nil {
				logrus.Error(err)
			}
		}

		fmt.Printf("[%s] ping: %s\r\n", time.Now().Format("2006-01-02 15:04:05"), r.String())
		time.Sleep(time.Second)

		t = time.Now()

		if sendExtranonceSubscribe {
			send := `{"id": 7, "method": "mining.extranonce.subscribe", "params": []}`
			conn.Write([]byte(send + "\n"))
			// fmt.Println("<--", send)
		} else {
			conn.Write([]byte(subscribe + "\n"))
			// fmt.Println("<--", subscribe)
		}
		err = conn.SetReadDeadline(time.Now().Add(timeout)) // timeout
		if err != nil {
			logrus.Error("setReadDeadline failed:", err)
		}

	}

	r := time.Since(t)
	d1 = fmt.Sprintf("[%s] close index %d min %s max %s ReadDeadline set timeout %s time use %s\r\n", time.Now().Format("2006-01-02 15:04:05"), lost, min, max, timeout, r)
	logrus.Error("lost connection")
	fmt.Printf("%s", d1)
	_, err = f.Write([]byte(d1))
	if err != nil {
		logrus.Error(err)
	}
}
