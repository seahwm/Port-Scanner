package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

type stat struct {
	port   int
	status bool
}

type Scanner struct {
}

// Port scan sequncially
func (s *Scanner) scanSeq() {
	// Sequencial
	for i := 21; i < 120; i++ {
		address := fmt.Sprintf("ming.com:%d", i)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Printf("%s 端口已关闭\n", address)
		} else {
			conn.Close()
			fmt.Printf("%s 端口打开了！！！\n", address)
		}
	}
}

// Port scan parallel
func (s *Scanner) scanParallel() {
	// Parallel
	var wg sync.WaitGroup
	start := time.Now()
	for i := 21; i < 120; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			dialer := &net.Dialer{
				Timeout: 3 * time.Second,
			}
			address := fmt.Sprintf("ming.com:%d", j)
			conn, err := dialer.Dial("tcp", address)
			if err != nil {
				fmt.Printf("%s 端口已关闭\n", address)
			} else {
				conn.Close()
				fmt.Printf("%s 端口打开了！！！\n", address)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start) / 1e9
	fmt.Printf("\n\n用时：%ds", elapsed)
}

func worker(port chan int, rsl chan stat) {
	for p := range port {
		dialer := &net.Dialer{
			Timeout: 3 * time.Second,
		}
		address := fmt.Sprintf("ming.com:%d", p)
		conn, err := dialer.Dial("tcp", address)
		if err != nil {
			rsl <- stat{status: false, port: p}
		} else {
			conn.Close()
			rsl <- stat{status: true, port: p}
		}
	}

}

func (s *Scanner) scanByWorker() {
	port := make(chan int, 100)
	rsl := make(chan stat)
	var success []int
	var fail []int

	// 分配100个worker，port有新数据就会随机分配给worker来执行
	for i := 0; i < cap(port); i++ {
		go worker(port, rsl)
	}

	// 由于worker函数里面会把数据放进result channel里面，而result channel并没有缓冲（只能存放1个数据）。
	// 所以当一个worker执行完毕后，其他worker不能放数据进去会照成堵塞，要把它放进另外一个线程，才不会打扰
	// 下面我们处理rsl数据。
	go func() {
		for i := 0; i < cap(port); i++ {
			port <- i
		}
	}()

	for i := 0; i < 100; i++ {
		p := <-rsl
		if p.status {
			success = append(success, p.port)
		} else {
			fail = append(fail, p.port)
		}
	}
	close(port)
	close(rsl)
	sort.Ints(success)
	sort.Ints(fail)
	for _, p := range success {
		fmt.Printf("%d 端口打开了！！！\n", p)
	}
	for _, p := range fail {
		fmt.Printf("%d 端口关闭了！！！\n", p)
	}
}

func main() {
	var s Scanner
	s.scanByWorker()
}
