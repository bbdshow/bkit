package typ

type IsLimit int

func (i IsLimit) Val() int {
	return int(i)
}

func (i IsLimit) All() bool {
	return int(i) == -1
}
