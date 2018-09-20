package project

var Handlers []Handler

type Handler interface {
	Execute()
	Name() string
}

func AddHandler(h Handler) {
	Handlers = append(Handlers, h)
}
