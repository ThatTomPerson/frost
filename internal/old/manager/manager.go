package manager

import (
	"context"
	"runtime"
	"sync"

	"github.com/apex/log"
)

type Job func(context.Context) error
type Handler interface {
	Name() string
	Detect(string) bool
	Exec(context.Context, string, chan Job) error
	PostExec(context.Context, string) error
}

var handlers []Handler

func AddHandler(h Handler) {
	handlers = append(handlers, h)
}

func logError(h Handler, err error) {
	if err != nil {
		log.Errorf("%s: %v", h.Name(), err)
	}
}

func Run(root string) {
	workers := runtime.NumCPU() * 4

	c := make(chan Job, workers*2)
	wrokerWg := &sync.WaitGroup{}
	wrokerWg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker(c, wrokerWg)
	}

	handlersWg := &sync.WaitGroup{}

	for _, h := range handlers {
		if h.Detect(root) {
			handlersWg.Add(1)
			go func(h Handler) {
				log.Infof("running %s handler", h.Name())
				logError(h, h.Exec(context.Background(), root, c))
				log.Infof("done %s handler", h.Name())
				handlersWg.Done()
			}(h)
		}
	}
	handlersWg.Wait()
	close(c)
	wrokerWg.Wait()

	handlersWg = &sync.WaitGroup{}
	for _, h := range handlers {
		if h.Detect(root) {
			handlersWg.Add(1)
			go func(h Handler) {
				logError(h, h.PostExec(context.Background(), root))
				handlersWg.Done()
			}(h)
		}
	}

	handlersWg.Wait()

}

func worker(c chan Job, wg *sync.WaitGroup) {
	for job := range c {
		err := job(context.Background())
		if err != nil {
			log.Error(err.Error())
		}
	}

	wg.Done()
}
