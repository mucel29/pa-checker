package display

type Display interface {
	Enable()
	Print(buffer string)
	Println(buffer string)
	PrintPage(index int, title string, buffer string)
	ReadLine() string
}
