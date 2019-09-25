package download

import (
	"fmt"
)

type Progresser interface {
	Init(size int64)
	Update(count int64, size int64)
}

type dummyProgresser struct {
	size  int64
	count int64
}

func (p *dummyProgresser) Init(size int64) {
	p.size = size
}

func (p *dummyProgresser) Update(count int64, size int64) {
	p.count = count
	p.size = size
	fmt.Printf("%.1f%% %d / %d   \n", float64(count)/float64(p.size)*100.0, count, size)
}
