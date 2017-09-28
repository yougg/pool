package pool

// 任务执行者
type worker struct {
	jobConsumeQueue Queue // worker接入的队列
	notifyChan      chan worker
}

// 开工
func (w *worker) start() {
	go w.execute()
}

// 循环执行任务
func (w *worker) execute() {
	killNextCycle := make(chan bool, 1)
	for {
		select {
		case job := <-w.jobConsumeQueue.poll():
			if nil != job.Run {
				job.Run(job.Args...)
			}
		case <-killNextCycle:
			w.notifyChan <- *w
			return
		}
		//situation when u are spawned but some other worker read your data
		//hence do not wait forever for data, if you dont have it now,
		//try to die in next cycle, cos select is random when all ready
		killNextCycle <- true
	}
}

// 初始化指定数量的worker, 并接入队列
func newWorkers(num int, q Queue) (workers chan worker) {
	workers = make(chan worker, num)
	for i := 0; i < num; i++ {
		worker := worker{
			jobConsumeQueue: q,
			notifyChan:      workers,
		}
		workers <- worker
	}
	return
}
