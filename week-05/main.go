package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {

	var size int = 100
	var reqThreshold int = 200
	var failedThreshold float64 = 0.8
	var duration time.Duration = time.Second * 2

	// 初始化
	r := NewSlidWindow(size, reqThreshold, failedThreshold, duration)
	r.Launch()     // 启动
	r.Monitor()    // 监控
	r.ShowStatus() //查询当前是否处于熔断状态
}

type SlidWindow struct {
	sync.RWMutex
	broken          bool
	size            int
	buckets         []*Bucket
	reqThreshold    int       // 总数阈值
	failedThreshold float64   // 失败率阈值
	lastBreakTime   time.Time // 上次熔断时间
	seeker          bool
	brokeTimeGap    time.Duration // 恢复时间间隔
}

func NewSlidWindow(
	size int,
	reqThreshold int,
	failedThreshold float64,
	brokeTimeGap time.Duration,
) *SlidWindow {
	return &SlidWindow{
		size:            size,
		buckets:         make([]*Bucket, 0, size),
		reqThreshold:    reqThreshold,
		failedThreshold: failedThreshold,
		brokeTimeGap:    brokeTimeGap,
	}
}

func (r *SlidWindow) AppendBucket() {
	r.Lock()
	defer r.Unlock()
	r.buckets = append(r.buckets, NewBucket())
	if !(len(r.buckets) < r.size+1) {
		r.buckets = r.buckets[1:]
	}
}

// 获取最后一个桶
func (r *SlidWindow) GetBucket() *Bucket {
	if len(r.buckets) == 0 {
		r.AppendBucket()
	}
	return r.buckets[len(r.buckets)-1]
}

func (r *SlidWindow) RecordReqResult(result bool) {
	r.GetBucket().Record(result)
}

func (r *SlidWindow) ShowAllBucket() {
	for _, v := range r.buckets {
		fmt.Printf("id: [%v] | total: [%d] | failed: [%d]\n", v.Timestamp, v.Total, v.Failed)
	}
}

func (r *SlidWindow) Launch() {
	go func() {
		for {
			r.AppendBucket()
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (r *SlidWindow) BreakJudgement() bool {
	r.RLock()
	defer r.RUnlock()
	total := 0
	failed := 0

	for _, v := range r.buckets {
		total += v.Total
		failed += v.Failed
	}

	if float64(failed)/float64(total) > r.failedThreshold && total > r.reqThreshold {
		return true
	}

	return false
}

// 监控
func (r *SlidWindow) Monitor() {
	go func() {
		for {
			if r.broken {
				if r.OverBrokenTimeGap() {
					r.Lock()
					r.broken = false
					r.Unlock()
				}
				continue
			}

			if r.BreakJudgement() {
				r.Lock()
				r.broken = true
				r.lastBreakTime = time.Now()
				r.Unlock()
			}
		}
	}()
}

func (r *SlidWindow) OverBrokenTimeGap() bool {
	return time.Since(r.lastBreakTime) > r.brokeTimeGap
}

// 状态展示
func (r *SlidWindow) ShowStatus() {
	go func() {
		for {
			log.Println(r.broken)
			time.Sleep(time.Second)
		}
	}()
}

// 获取状态
func (r *SlidWindow) Broken() bool {
	return r.broken
}

func (r *SlidWindow) SetSeeker(status bool) {
	r.Lock()
	defer r.Unlock()
}

func (r *SlidWindow) Seeker() bool {
	return r.seeker
}

// bucket
type Bucket struct {
	sync.RWMutex
	Total     int // 请求总数
	Failed    int // 失败
	Timestamp time.Time
}

func NewBucket() *Bucket {
	return &Bucket{
		Timestamp: time.Now(),
	}
}

func (b *Bucket) Record(result bool) {
	b.Lock()
	defer b.Unlock()
	if !result {
		b.Failed++
	}
	b.Total++
}