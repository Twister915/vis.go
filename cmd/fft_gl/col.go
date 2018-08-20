package main

type col [4]float32
type cols []col

func (cs cols) toDec() {
	for i, c := range cs {
		cs[i] = c.toDec()
	}
	return
}

func (c col) toDec() col {
	return col{c[0] / 255, c[1] / 255, c[2] / 255, c[3] / 255}
}
