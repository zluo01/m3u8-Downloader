package hackpool

import (
	"github.com/cheggaaa/pb/v3"
	"sync"
)

// HackPool Modified from "github.com/greyh4t/hackpool"
type HackPool struct {
	numGo    int
	messages chan []interface{}
	function func(...interface{})
}

func New(numGoroutine int, function func(...interface{})) *HackPool {
	return &HackPool{
		numGo:    numGoroutine,
		messages: make(chan []interface{}),
		function: function,
	}
}

func (c *HackPool) Push(data ...interface{}) {
	c.messages <- data
}

func (c *HackPool) CloseQueue() {
	close(c.messages)
}

func (c *HackPool) Run(count int) {
	var wg sync.WaitGroup

	wg.Add(c.numGo)

	bar := pb.New(count)
	for i := 0; i < c.numGo; i++ {
		go func(bar *pb.ProgressBar) {
			bar.Start()
			for v := range c.messages {
				bar.Increment()
				c.function(v...)
			}
			bar.Finish()
			wg.Done()
		}(bar)
	}
	wg.Wait()
}
