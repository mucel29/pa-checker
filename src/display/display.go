package display

type Display interface {
	Enable()
	Print(buffer string)
	Println(buffer string)
	PrintPage(index int, title string, buffer string)
	ReadLine() string
	IsInteractive() bool
}

func (bd *BasicDisplay) IsInteractive() bool {
	return false
}

func (id *InteractiveDisplay) IsInteractive() bool {
	return true
}
