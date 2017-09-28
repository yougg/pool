package pool

import (
	"errors"
)

// 调度器接口
type Scheduler interface {
	// 添加工作到线程池队列中
	// 如果队列已满,则返回失败
	Add(job Job) bool

	// 添加任务到线程池队列中
	// 当队列容量已满时:
	// 有界队列, 超过队列容量时新工作入队直接失败
	// 无界队列, 新工作入队时一直等待直到入队成功
	Join(t Task, args ...interface{}) bool
	schedule()
}

// 调度器接口实现, 线程池
type pool struct {
	queue           Queue
	inputForWorkers Queue
	workers         chan worker
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
	for {
		p.inputForWorkers.add(<-p.queue.poll())
		select {
		case idleWorker := <-p.workers:
			idleWorker.start()
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

	var dataQueue Queue
	var dispatcherQueue Queue
	//dispatcher can hold more than number of workers
	//i.e when all workers busy, act like a buffer
	dataq := make(basequeue, queueCap)
	dispatchq := make(basequeue, queueCap*10)
	switch qType {
	case SynchronousQueue, PriorityBlockingQueue:
		// TODO 功能暂时未实现
		dataQueue = dataq
		dispatcherQueue = dispatchq
	case ArrayBlockingQueue:
		dataQueue = arrayqueue{dataq}
		dispatcherQueue = arrayqueue{dispatchq}
	case LinkedBlockingQueue:
		dataQueue = linkedqueue{dataq}
		dispatcherQueue = linkedqueue{dispatchq}
	default:
		dataQueue = dataq
		dispatcherQueue = dispatchq
	}
	workers := newWorkers(workerNum, dataQueue)
	s = &pool{
		queue:           dispatcherQueue,
		workers:         workers,
		inputForWorkers: dataQueue,
	}

	go s.schedule()

	return s, nil
}
