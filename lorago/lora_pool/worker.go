package lora_pool

import (
	"github.com/LorraineWen/lorago/lora_log"
	"time"
)

/*
*@Author: LorraineWen
*定义worker
*支持任务执行，支持task执行时的异常处理
 */
type Worker struct {
	pool         *Pool       //父类，一般都是Pool里面调用work，这个work就可以通过pool成员获取到父亲pool的属性
	taskChannel  chan func() //一个worker对应一个任务，所以taskChannel的长度在Submit中设置为1
	lastTaskTime time.Time   //最后一次执行任务的时间
}

func (w *Worker) run() {
	w.pool.AddRunningWorkerNum()
	go w.runTask()
}

func (w *Worker) runTask() {
	//task执行的时候可能出现错误
	defer func() {
		w.pool.SubRunningWorkerNum()
		w.pool.workCache.Put(w)
		if err := recover(); err != nil {
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler() //用户自定义异常处理方式
			} else {
				lora_log.NewLogger().Error(err)
			}
		}
		w.pool.workerCon.Signal() //如果不加上，freeWrokerCheck函数就永远阻塞了，出现了一次异常，其他的task就都无法执行了
	}()
	for task := range w.taskChannel {
		if task == nil {
			return
		}
		task()
		//任务运行完成，worker空闲
		w.pool.PutWorker(w)
	}
}
