package circ

type Data struct {
	I int
	S string
}

type CircBuff struct {
	Len       int
	j         int
	overlaped bool
	buff      []Data
}

func (cb *CircBuff) Add(s string, i int) {
	if cb.j >= cb.Len {
		return
	}
	if cb.j == cb.Len-1 {
		cb.buff[cb.j] = Data{i, s}
		cb.j = 0
		cb.overlaped = true
		return
	}
	cb.buff[cb.j] = Data{i, s}
	cb.j++
}

func (cb *CircBuff) Read() []Data {
	res := make([]Data, 0, cb.Len)

	start := 0
	if cb.overlaped {
		start = cb.j
	}

	newRound := false
	for {
		if cb.overlaped && start >= cb.Len {
			start = 0
			newRound = true
			continue
		} else if newRound && start >= cb.j {
			break
		} else if !cb.overlaped && start >= cb.j {
			break
		}

		res = append(res, cb.buff[start])
		start++
	}

	cb.j = 0
	cb.overlaped = false

	return res
}

func NewCirc(l int) *CircBuff {
	return &CircBuff{
		Len:  l,
		buff: make([]Data, l),
	}
}
