package core

type Throttle struct {
	Queue         chan bool
	MaxConcurrent int
}

func (t *Throttle) New(size int) Throttle {
	t.MaxConcurrent = size
	t.Queue = make(chan bool, size)
	return *t
}

func (t *Throttle) WaitForDone() {
	for i := 0; i < cap(t.Queue); i++ {
		t.Queue <- true
	}
	for i := 0; i < cap(t.Queue); i++ {
		<-t.Queue
	}
}

func (t *Throttle) Done() {
	<-t.Queue
}

func (t *Throttle) WaitForSpot() {
	t.Queue <- true
}
