package main

import (
	"fmt"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	redisPool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   1000,
		IdleTimeout: 60,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}
	defer redisPool.Close()

	reConn := redisPool.Get()
	defer reConn.Close()

	// 初始化
	var value string = "1"
	for j := 0; j < 10000; {
		value += "1"
		j++
	}

	var avg uint32
	for i := 0; i < 50; i++ {
		var tempValue = ""
		beforeMem := MemStat()

		// 1w, 2w, 3w ..... 50w
		for k := -1; k < i; k++ {
			tempValue += value
		}

		_, err := reConn.Do("set", "key"+strconv.Itoa(i+1), tempValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		tempValue = ""

		// 获取内存信息
		time.Sleep(1 * time.Second)
		afterMem := MemStat()

		var used uint32 = 0
	
		if afterMem.Used < beforeMem.Used {
			used = 0
		} else {
			used = afterMem.Used - beforeMem.Used
		}

		fmt.Println("插入前key", strconv.Itoa(i+1), " 内存使用: ", beforeMem.Used, "b, 插入后key", strconv.Itoa(i+1), " 内存使用: ", afterMem.Used, "b, 插入占用内存: ", used, "b")

		avg += used
	}

	avg = avg / 50 / 1000 // k
	fmt.Println("\n50个key平均占用内存 ", avg, "k")

	// 清除key
	for i := 0; i < 50; i++ {
		_, err := reConn.Do("del", "key"+strconv.Itoa(i+1))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

type MemStatus struct {
	All  uint32 `json:"all"`
	Used uint32 `json:"used"`
	Free uint32 `json:"free"`
	Self uint64 `json:"self"`
}



func MemStat() MemStatus {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	mem := MemStatus{}
	mem.Self = memStat.Alloc

	sysInfo := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysInfo)
	if err == nil {
		mem.All = uint32(sysInfo.Totalram) * uint32(syscall.Getpagesize())
		mem.Free = uint32(sysInfo.Freeram) * uint32(syscall.Getpagesize())
		mem.Used = mem.All - mem.Free
	}

	return mem
}