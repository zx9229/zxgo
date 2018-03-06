package zxgo

import (
	"sync"
)

type QueueCallbackFun func(interface{})

type Queue struct {
	cache []interface{}
	mutex *sync.Mutex
	cbFun QueueCallbackFun //回调函数时有效
	cond  *sync.Cond       //回调函数时有效
	alive bool             //回调函数时有效
}

func NewQueue(callbackFun QueueCallbackFun) *Queue {
	var queue *Queue = nil
	if callbackFun == nil {
		queue = &Queue{cache: make([]interface{}, 0), mutex: new(sync.Mutex)}
	} else {
		queue = &Queue{make([]interface{}, 0), new(sync.Mutex), callbackFun, sync.NewCond(new(sync.Mutex)), true}
		go func() {
			var data interface{} = nil
			var isOk bool = false
			for queue.alive {
				if data, isOk = queue.popData(); isOk {
					queue.cbFun(data)
					continue
				} else {
					queue.cond.L.Lock()   //获取锁
					queue.cond.Wait()     //等待通知,暂时阻塞
					queue.cond.L.Unlock() //释放锁
				}
			}
		}()
	}
	return queue
}

func (self *Queue) ExitGoroutine() {
	self.alive = false
	if self.cbFun != nil {
		self.cond.Broadcast()
	}
}

func (self *Queue) HasCbFun() bool {
	if self.cbFun != nil {
		return true
	} else {
		return false
	}
}

func (self *Queue) Size() int {
	var data int = 0
	self.mutex.Lock()
	data = len(self.cache)
	self.mutex.Unlock()
	return data
}

func (self *Queue) Push(data interface{}) {
	self.mutex.Lock()
	self.cache = append(self.cache, data)
	self.mutex.Unlock()

	if self.cbFun != nil {
		self.cond.Signal()
	}
}

func (self *Queue) popData() (data interface{}, ok bool) {

	self.mutex.Lock()
	if 0 < len(self.cache) {
		data = self.cache[0]
		ok = true
		self.cache = self.cache[1:]
	}
	self.mutex.Unlock()

	return
}

func (self *Queue) Pop() (data interface{}, ok bool) {
	if self.HasCbFun() {
		return
	}
	data, ok = self.popData()
	return
}

/* 这是一个例子
type ExampleData struct {
	data int
	que  *Queue
}

func NewExampleData() *ExampleData {
	example := &ExampleData{}
	cbFun := func(iData interface{}) {
		fmt.Println("CallBackFunc", iData)
		data := iData.(int)
		example.data += data
	}
	example.que = NewQueue(cbFun)
	return example
}

func (self *ExampleData) FeedData(data int) {
	self.que.Push(data)
}

func (self *ExampleData) Print() {
	fmt.Println(self.data)
}

func main() {
	example := NewExampleData()
	example.FeedData(1)
	example.FeedData(2)
	example.FeedData(3)
	time.Sleep(time.Second * 3)
	example.Print()
	time.Sleep(time.Second * 3)
}*/
