package display

import (
	"fmt"
)

type BasicDisplay struct {
}

func (*BasicDisplay) Enable() {}

func (display *BasicDisplay) Print(buffer string) {
	fmt.Print(buffer)
}

func (display *BasicDisplay) Println(buffer string) {
	display.Print(buffer)
	fmt.Println()
}

func (display *BasicDisplay) PrintPage(index int, title string, buffer string) {
	fmt.Println(title)
	fmt.Println(buffer)
	//slog.Warn("BasicDisplay does not support paging")
}
