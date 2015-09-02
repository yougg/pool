package pool

// 任务
type Task func(args ...interface{})

// 工作
type Job struct {
	Run  Task
	Args []interface{}
}
