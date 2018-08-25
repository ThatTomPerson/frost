package task

import (
	"sync"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type Progress struct {
	p    *mpb.Progress
	bars map[int]*Bar
}

type Bar struct {
	b    *mpb.Bar
	p    *Progress
	name string
	task string
}

func New(wg *sync.WaitGroup) *Progress {
	p := mpb.New(mpb.WithWaitGroup(wg))

	return &Progress{p: p}
}

func (p *Progress) Add(name, task string, count int) *Bar {
	b := p.p.AddBar(int64(count),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.Name(task, decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
	)

	bar := Bar{
		b:    b,
		p:    p,
		name: name,
		task: task,
	}

	p.bars[bar.ID()] = &bar

	return &bar
}

func (b *Bar) ID() int {
	return b.b.ID()
}

func (b *Bar) Advance() {
	b.b.Increment()
}

func (b *Bar) SetName() {
	b.b.RemoveAllAppenders
}
