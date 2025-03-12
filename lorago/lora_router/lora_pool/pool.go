package lora_pool

/*
*@Author: LorraineWen
*定义go程池，一个pool包含多个worker，一个worker对应一个go程
*支持任务提交，支持超时回收空闲worker
 */
import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 3

type Pool struct {
	capacity         int32         //go程池容量
	runningWorkerNum int32         //正在运行的worker的数量
	workers          []*Worker     //空闲的worker
	expireTime       time.Duration //如果空闲的worker存在超过了这个时间，就回收掉这个worker
	workerLock       sync.Mutex    //对worker进行读写的时候需要加锁
	workerCon        *sync.Cond    //条件变量，当出现了空闲worker时，通知worker回收go程去回收worker
	releaseOnce      sync.Once     //专门用于释放go程池的资源
	releaseFlag      chan struct{} //一个信号，如果release中有值，那么说明资源已经释放完毕了
	workCache        sync.Pool     //用来提前创建worker
	PanicHandler     func()        //用户可以自定义的异常处理函数
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, errors.New("go程池的容量必须大于0")
	}
	if expire <= 0 {
		return nil, errors.New("超时时间必须大于0")
	}
	p := &Pool{
		capacity:    int32(cap),
		expireTime:  time.Duration(expire) * time.Second,
		releaseFlag: make(chan struct{}, 1),
	}
	p.workCache.New = func() interface{} {
		return &Worker{
			pool:        p,
			taskChannel: make(chan func(), 1),
		}
	}
	p.workerCon = sync.NewCond(&p.workerLock) //需要配合锁来使用
	go p.expireWorker()                       //单独开一个go程清理超时并且空闲的worker
	return p, nil
}

// 定期清理空闲的worker
func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expireTime)
	for range ticker.C {
		if len(p.releaseFlag) > 0 {
			break
		}
		p.workerLock.Lock()
		freeWorkers := p.workers
		n := len(freeWorkers) - 1
		if n >= 0 {
			var clearN = -1
			for i, w := range freeWorkers {
				//如果wroker的最后执行任务时间和当前时间的差值小于expire就释放掉
				if time.Now().Sub(w.lastTaskTime) <= p.expireTime {
					break
				}
				//需要清除的
				clearN = i
				w.taskChannel <- nil
				freeWorkers[i] = nil
			}
			if clearN != -1 {
				if clearN >= len(freeWorkers)-1 {
					p.workers = freeWorkers[:0]
				} else {
					p.workers = freeWorkers[clearN+1:]
				}
				fmt.Printf("清除完成,running:%d, workers:%v \n", p.runningWorkerNum, p.workers)
			}
		}
		p.workerLock.Unlock()
	}
}

// 提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.releaseFlag) > 0 {
		return errors.New("go程池已经关闭")
	}
	//获取池里面的一个worker，然后执行任务就可以了
	w := p.GetWorker()
	w.taskChannel <- task
	return nil
}

// 获取pool里面的worker
func (p *Pool) GetWorker() (w *Worker) {
	//1. 目的获取pool里面的worker
	//2. 优先使用workCache中的
	readyWroker := func() {
		w = p.workCache.Get().(*Worker)
		w.run()
	}
	p.workerLock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n > 0 {
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
		p.workerLock.Unlock()
		return w
	}
	//3. 如果没有空闲的worker，要新建一个worker
	if p.runningWorkerNum < p.capacity {
		p.workerLock.Unlock()
		readyWroker()
		return
	}
	p.workerLock.Unlock()
	//4. 如果正在运行的workers 如果大于pool容量，阻塞等待，worker释放
	return p.waitFreeWorker()
}
func (p *Pool) waitFreeWorker() *Worker {
	p.workerLock.Lock()
	p.workerCon.Wait()

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		p.workerLock.Unlock()
		if p.runningWorkerNum < p.capacity {
			//还不够pool的容量，直接新建一个
			c := p.workCache.Get()
			var w *Worker
			if c == nil {
				w = &Worker{
					pool:        p,
					taskChannel: make(chan func(), 1),
				}
			} else {
				w = c.(*Worker)
			}
			w.run()
			return w
		}
		return p.waitFreeWorker()
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.workerLock.Unlock()
	return w
}
func (p *Pool) AddRunningWorkerNum() {
	atomic.AddInt32(&p.runningWorkerNum, 1)
}
func (p *Pool) SubRunningWorkerNum() {
	atomic.AddInt32(&p.runningWorkerNum, -1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTaskTime = time.Now()
	p.workerLock.Lock()
	p.workers = append(p.workers, w)
	p.workerCon.Signal()
	p.workerLock.Unlock()
}

// 释放go程池里面的资源，只需要释放一次
func (p *Pool) Release() {
	p.releaseOnce.Do(func() {
		p.workerLock.Lock()
		//将所有的worker里面的资源都释放了
		workers := p.workers
		for i, w := range workers {
			if w == nil {
				continue
			}
			w.taskChannel = nil
			w.pool = nil
			workers[i] = nil
		}
		//将pool的所有资源都释放了
		p.workers = nil
		p.workerLock.Unlock()
		p.releaseFlag <- struct{}{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.releaseFlag) > 0
}

func (p *Pool) Restart() bool {
	if len(p.releaseFlag) <= 0 {
		return true
	}
	_ = <-p.releaseFlag
	go p.expireWorker()
	return true
}
func (p *Pool) GetRunningNum() int {
	return int(atomic.LoadInt32(&p.runningWorkerNum))
}
func (p *Pool) GetFreeNum() int {
	return int(p.capacity - p.runningWorkerNum)
}
