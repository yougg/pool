package pool

import "time"

// 任务执行者
type worker struct {
	idle bool      // worker是否空闲
	end  chan bool // worker是否停止
	q    Queue     // worker接入的队列
}

// 开工
func (w *worker) start() {
	go w.execute()
}

// 停工
func (w *worker) stop() {
	w.end <- true
}

// 循环执行任务
func (w *worker) execute() {
	c := time.NewTicker(time.Second).C
	for {
		select {
		case job := <-w.q.poll():
			w.idle = false
			if nil != job.Run {
				job.Run(job.Args...)
			}
		case <-c:
			w.idle = true
		case <-w.end:
			return
		}
	}
}

// 初始化指定数量的worker, 并接入队列
func newWorkers(num int, q Queue) (ws []*worker) {
	for i := 0; i < num; i++ {
		ws = append(ws, &worker{
			idle: true,
			end:  make(chan bool),
			q:    q,
		})
	}
	return
}
