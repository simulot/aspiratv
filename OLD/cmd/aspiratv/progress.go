package main

import (
	"context"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

// ProgressBarNoop is a no-op progress bar for headless mod
type ProgressBarNoop struct {
}

// Stage set the name of the bar
func (ProgressBarNoop) Stage(string) {}

// Total set the total of the task
func (ProgressBarNoop) Total(int) {}

// Update set the curent value of the task
func (ProgressBarNoop) Update(int) {}

// Done to terminate
func (ProgressBarNoop) Done() {}

type barContainer struct {
	*mpb.Progress
}

// NewBarContainer create a mpb containter for other bars
func NewBarContainer(ctx context.Context) *barContainer {
	p := &barContainer{
		Progress: mpb.NewWithContext(
			ctx,
			mpb.WithWidth(64),
		),
	}
	return p
}

func (c *barContainer) Done() {
	c.Wait()
}

type bar struct {
	*mpb.Bar
	name string
}

// NewProviderBar create a bar for following downloaded items per provider
func (b barContainer) NewProviderBar(name string, id int) *bar {
	newB := &bar{
		Bar: b.AddBar(0,
			mpb.BarWidth(12),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Counters(0, "  %3d/%3d"), "completed"),
				decor.Name("      Pulling "+name),
			),
		),
		name: name,
	}
	newB.SetPriority(id)
	return newB
}

func (b *bar) Stage(s string) {}
func (b *bar) Total(total int) {
	b.SetTotal(int64(total), false)
}
func (b *bar) Update(current int) {
	b.SetCurrent(int64(current))
}
func (b *bar) Done() {
	b.Abort(true)

}

type downloadBar struct {
	*mpb.Bar
	name      string
	startTime time.Time
	lastSize  int
}

// NewDownloadBar create a progression bar attached to the container
func (b barContainer) NewDownloadBar(name string, id int) *downloadBar {
	newB := &downloadBar{
		Bar: b.AddBar(100*1024*1024*1024,
			mpb.BarWidth(12),
			mpb.AppendDecorators(
				decor.AverageSpeed(decor.UnitKB, " %.1f", decor.WC{W: 15, C: decor.DidentRight}),
				decor.Name(name),
			),
			mpb.BarRemoveOnComplete(),
		),
		name:      name,
		startTime: time.Now(),
	}
	newB.SetPriority(int(id))
	return newB
}

func (b *downloadBar) Stage(s string) {}
func (b *downloadBar) Total(total int) {
	b.SetTotal(int64(total), false)
}
func (b *downloadBar) Update(current int) {

	b.IncrBy(current - b.lastSize)
	b.DecoratorEwmaUpdate(time.Since(b.startTime))
	b.lastSize = current
}
func (b *downloadBar) Done() {
	b.Abort(true)

}
