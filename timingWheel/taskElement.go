package timingWheel

type taskElement struct {
	task func()
	// index in array
	pos   int
	cycle int
	key   string
}
