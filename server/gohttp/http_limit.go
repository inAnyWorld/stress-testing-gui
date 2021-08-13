package gohttp

type GoLimit struct {
	Num int
	C   chan struct{}
}

func NewGLimit(num int) *GoLimit {
	return &GoLimit{
		Num: num,
		C : make(chan struct{}, num),
	}
}

func (g *GoLimit) Run(f func()) {
	g.C <- struct{}{}
	go func() {
		f()
		<-g.C
	}()
}

