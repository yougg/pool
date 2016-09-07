package pool

// 队列类型
type QueueType int

const (
	SynchronousQueue      QueueType = iota // 同步阻塞队列
	ArrayBlockingQueue                     // 有界队列, 超过队列容量时新工作入队直接失败
	LinkedBlockingQueue                    // 无界队列, 新工作入队时一直等待直到入队成功
	PriorityBlockingQueue                  // 优先级队列
)

// 队列接口
type Queue interface {
	add(job Job) bool
	poll() <-chan Job
	length() int
}

type basequeue chan Job

// 加入工作到队列中
// 直到加入成功才返回
func (q basequeue) add(job Job) (added bool) {
	q <- job
	return true
}

// 从队列中取出工作
func (q basequeue) poll() <-chan Job {
	return q
}

// 当前队列的长度
func (q basequeue) length() int {
	return len(q)
}

// 阻塞队列实现
type arrayqueue struct {
	basequeue
}

// 加入工作到队列中
// 队列已满则加入失败
func (q arrayqueue) add(job Job) (added bool) {
	select {
	case q.basequeue <- job:
		added = true
	default:
		added = false
	}
	return
}

// 阻塞队列实现
type linkedqueue struct {
	basequeue
}
