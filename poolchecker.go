package main

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
)

type StratumStatus uint

const (
	StratumStatusUnknown StratumStatus = iota
	StratumStatusSubscribed
	StratumStatusAuthorized
)

const (
	INT64_MAX = int64(^uint64(0) >> 1)
	INT64_MIN = ^INT64_MAX
)

var (
	WriteTimeDuration    = time.Second * 1
	ConnectionTimeout    = time.Second * 3
	ConnReadWriteTimeout = time.Second * 3
)

type Timer struct {
	timeStart time.Time
	timeStop  time.Time
}

type PoolChecker struct {
	Mode        int
	IsRunning   bool
	Addr        string
	conn        net.Conn
	MinPing     time.Duration
	MaxPing     time.Duration
	AvgPing     time.Duration
	count       uint64
	BrokenCount uint64
	timer       *Timer
	CreatedAt   time.Time
	Status      StratumStatus
}

func RunPoolChecker(ctx context.Context, wg *sync.WaitGroup, addr string, Mode int) {
	var err error
	checker := PoolChecker{
		Addr:      addr,
		Mode:      Mode,
		MinPing:   time.Duration(INT64_MAX),
		timer:     NewTimer(),
		CreatedAt: time.Now(),
	}
	go func() {
		<-ctx.Done()
		checker.IsRunning = false
		checker.conn.Close()
	}()

	checker.IsRunning = true
	wg.Add(1)
	for {
		checker.conn, err = net.DialTimeout("tcp", addr, ConnectionTimeout)
		if err != nil {
			time.Sleep(ConnectionTimeout)
			continue
		}
		checker.handle()
		if !checker.IsRunning {
			glog.V(1).Infof("addr: %s\nstartAt: %s, endAt: %s\nlost connection times: %d\nmin/avg/max = %s/%s/%s\n",
				checker.Addr, checker.CreatedAt, time.Now(), checker.BrokenCount, checker.MinPing, checker.AvgPing, checker.MaxPing)
			wg.Done()
			return
		}
		glog.Error("lost connection")
		checker.BrokenCount++
	}
}
func (poolWatcher *PoolChecker) handle() {
	poolWatcher.Status = StratumStatusUnknown
	defer poolWatcher.conn.Close()
	for {
		err := poolWatcher.Write()
		if err != nil {
			if poolWatcher.IsRunning {
				glog.Error(err)
			}
			return
		}
		err = poolWatcher.Read()
		if err != nil {
			if poolWatcher.IsRunning {
				glog.Error(err)
			}
			return
		}
		time.Sleep(WriteTimeDuration)
	}
}

func (poolWatcher *PoolChecker) Write() (err error) {
	str := "{\"id\":1,\"method\":\"mining.subscribe\",\"params\":[\"__PoolWatcher__\"]}\n"
	switch poolWatcher.Mode {
	case 0: //bn.huobipool.com
	case 1: //hk.huobipool.com
		switch poolWatcher.Status {
		case StratumStatusSubscribed:
			str = "{\"id\": 7, \"method\": \"mining.extranonce.subscribe\", \"params\": []}\n"
		case StratumStatusAuthorized:
			str = "{\"id\": 7, \"method\": \"mining.extranonce.subscribe\", \"params\": []}\n"
		}
	}

	err = poolWatcher.conn.SetWriteDeadline(time.Now().Add(ConnReadWriteTimeout))
	if err != nil {
		return err
	}
	poolWatcher.timer.Reset()
	_, err = poolWatcher.conn.Write([]byte(str))
	if err != nil {
		return err
	}
	glog.V(4).Info("addr: ", poolWatcher.Addr, ", send: ", str)
	return nil
}
func (poolWatcher *PoolChecker) Read() (err error) {
	scanner := bufio.NewScanner(poolWatcher.conn)
	err = poolWatcher.conn.SetReadDeadline(time.Now().Add(ConnReadWriteTimeout))
	if err != nil {
		return err
	}
	if scanner.Scan() {
		glog.V(4).Info("addr: ", poolWatcher.Addr, ", recv: ", scanner.Text())
		if poolWatcher.Status == StratumStatusUnknown {
			poolWatcher.Status = StratumStatusSubscribed
		} else if poolWatcher.Status == StratumStatusSubscribed {
			poolWatcher.Status = StratumStatusAuthorized
		}
		duration := poolWatcher.timer.GetDuration()
		glog.V(1).Info("addr: ", poolWatcher.Addr, ", ping: ", duration)

		poolWatcher.count++
		if poolWatcher.count != 1 {
			poolWatcher.AvgPing = time.Duration(float64(poolWatcher.AvgPing)/float64(poolWatcher.count)*float64(poolWatcher.count-1) + float64(duration)/float64(poolWatcher.count))
		} else {
			poolWatcher.AvgPing = duration
		}
		glog.V(4).Info("addr: ", poolWatcher.Addr, ", avg ping: ", poolWatcher.AvgPing)
		if duration > poolWatcher.MaxPing {
			glog.V(2).Info("addr: ", poolWatcher.Addr, ", max ping: ", duration)
			poolWatcher.MaxPing = duration
		}
		if duration < poolWatcher.MinPing {
			glog.V(2).Info("addr: ", poolWatcher.Addr, ", min ping: ", duration)
			poolWatcher.MinPing = duration
		}
		return nil
	}
	return errors.New("poolWathcer read EOF")
}

func NewTimer() *Timer {
	return &Timer{}
}
func (timer *Timer) Reset() {
	timer.timeStart = time.Now()
}
func (timer *Timer) GetDuration() time.Duration {
	return time.Since(timer.timeStart)
}
