package pool

import (
	"errors"
	"time"
)

// 调度器接口
type Scheduler interface {
	// 添加工作到线程池队列中
	// 如果队列已满,则返回失败
	Add(job Job) bool

	// 添加任务到线程池队列中
	// 如果队列已满,则返回失败
	Join(t Task, args ...interface{}) bool
	schedule()
}

// 调度器接口实现, 线程池
type pool struct {
	queue   Queue
	workers []*worker
}

// 添加工作到线程池队列中
// 如果队列已满,则返回失败
func (p *pool) Add(job Job) bool {
	return p.queue.add(job)
}

// 添加任务到线程池队列中
// 如果队列已满,则返回失败
func (p *pool) Join(t Task, args ...interface{}) bool {
	return p.queue.add(Job{
		Run:  t,
		Args: args,
	})
}

// 开启并发线程调度
// 线程池工作者worker的数量会动态增减
func (p *pool) schedule() {
	duration := time.Second
	workingNum := 0
	for ; ; time.Sleep(duration) {
		switch {
		case p.queue.length() > 0 && workingNum < len(p.workers):
			for _, w := range p.workers {
				if w.idle {
					w.start()
					duration = 100 * time.Millisecond
					workingNum++
					break
				}
			}
		case p.queue.length() == 0 && workingNum > 0:
			for _, w := range p.workers {
				if w.idle {
					w.stop()
					duration = 10 * time.Second
					workingNum--
					break
				}
			}
		default:
			duration = time.Second
		}
	}
}

// 创建默认的线程池队列调度器
// 队列长度: 1000
// 并发数量: 10
func Default() (s Scheduler) {
	s, _ = New(ArrayBlockingQueue, 1000, 10)
	return s
}

// 创建新的线程池队列调度器
// workerNum: 并发工作数量
func NewBlocking(workerNum int) (s Scheduler, err error) {
	if workerNum < 1 {
		return nil, errors.New("The number of worker is error.")
	}
	return New(LinkedBlockingQueue, workerNum, workerNum)
}

// 创建新的线程池队列调度器
// qType:     线程池队列类型
// queueCap:  队列最大容量
// workerNum: 并发工作数量
func New(qType QueueType, queueCap, workerNum int) (s Scheduler, err error) {
	if queueCap < 0 || workerNum < 1 {
		return nil, errors.New("The max queue capcity or max worker number are error.")
	}

	var q Queue
	bq := make(basequeue, queueCap)
	switch qType {
	case SynchronousQueue, PriorityBlockingQueue:
		// TODO 功能暂时未实现
		// q = synchronousqueue{bq}
		// q = priorityqueue{bq}
	case ArrayBlockingQueue:
		q = arrayqueue{bq}
	case LinkedBlockingQueue:
		q = linkedqueue{bq}
	default:
		q = bq
	}

	s = &pool{
		queue:   q,
		workers: newWorkers(workerNum, q),
	}

	go s.schedule()

	return s, nil
}
