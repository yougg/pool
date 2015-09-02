package pool

import (
	"sync"
	"testing"
	"time"
)

func MyTask(args ...interface{}) {
	// fmt.Print(args, ` `)
	time.Sleep(100 * time.Millisecond)
}

func TestNewPool(t *testing.T) {
	t.Log("Test array blocking queue scheduled pool")

	p, err := New(ARRAY_BLOCKING_QUEUE, -1, 5)
	if nil == err {
		t.Error("Create new array blocking queue failed.", err)
		return
	}

	// 初始化一个队列容量10,并发数量5的线程池队列调度器
	p, err = New(ARRAY_BLOCKING_QUEUE, 10, 5)
	if nil != err {
		t.Error("Create new array blocking queue failed.", err)
		return
	}
	result := make(map[int]bool)
	// 一旦队列容量满, 后续加入的任务会直接失败返回
	for i := 1; i < 50; i++ {
		result[i] = p.Join(MyTask, i)
		time.Sleep(10 * time.Millisecond)
	}
	t.Log("Array Blocking Result:", result)
}

func TestBlockingPool(t *testing.T) {
	t.Log("Test linked blocking queue scheduled pool")
	// 初始化一个并发数量10的线程池队列调度器
	p, err := NewBlocking(10)
	if nil != err {
		t.Error("Create new linked blocking queue failed.", err)
		return
	}
	result := make(map[int]bool)
	// 所有的任务都会在添加成功后再返回, 注意: 添加任务成功不等于任务执行成功
	for i := 101; i < 150; i++ {
		result[i] = p.Join(MyTask, i)
	}
	t.Log("Linked Blocking Result:", result)
}

func TestWaitAll(t *testing.T) {
	p, err := New(ARRAY_BLOCKING_QUEUE, 10, 5)
	if nil != err {
		t.Error("Create new array blocking queue failed.", err)
		return
	}
	// 创建10个并行任务, 等待所有任务执行完成后再退出
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		p.Add(Job{
			Run: func(args ...interface{}) {
				if x, ok := args[0].(int); ok {
					time.Sleep(time.Duration(x * 100) * time.Millisecond)
				}
				if wg, ok := args[1].(*sync.WaitGroup); ok {
					wg.Done()
				}
			},
			Args: []interface{}{i, wg},
		})
	}
	wg.Wait()
	t.Log("Wait 10 jobs finish.")
}

func TestGetAllResult(t *testing.T) {
	t.Log("Sum 1 to 100 use 10 tasks")
	p, err := NewBlocking(10)
	if nil != err {
		t.Error("Create new linked blocking queue failed.", err)
		return
	}
	// 使用10个并行任务计算1到100的和,最后统计总和
	ch := make(chan int)
	for i := 1; i < 100; i += 10 {
		p.Join(func(args ...interface{}) {
			var i int
			if x, ok := args[0].(int); ok {
				i = x
			}
			if ch, ok := args[1].(chan int); ok {
				var s int
				for j := i; j < (i + 10); j++ {
					s += j
				}
				ch <- s
			}
		}, i, ch)
	}

	sum, count := 0, 0
	for s := range ch {
		sum += s
		if count++; count >= 10 {
			break
		}
	}

	t.Log("Sum 1 to 100 result:", sum)
}

func BenchmarkNewPool(b *testing.B) {
	// 初始化一个队列容量10,并发数量5的线程池队列调度器
	p, _ := New(ARRAY_BLOCKING_QUEUE, 10, 5)
	p.Join(func(args ...interface{}) {
		x, sum := 0, 0
		var ok bool
		if x, ok = args[0].(int); ok {
			for j := 0; j <= x; j++ {
				sum += j
			}
		}
		b.Logf("sum 0 to %d :\t%d\n", x, sum)
	}, b.N)
}
