package display

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strconv"
)

type PageStack []Page

func (ps *PageStack) Push() {

	*ps = append(*ps, Page{})
}

func (ps *PageStack) Pop() {
	*ps = (*ps)[:len(*ps)-1]
}

func (ps *PageStack) Page() *Page {
	return &(*ps)[len(*ps)-1]
}

func (ps *PageStack) Len() int {
	return len(*ps)
}

type WritableContainer struct {
	Container *tview.Flex
	Sections  []*tview.TextView
}

func NewWritableContainer(flexDirection int) *WritableContainer {
	container := WritableContainer{}
	container.Container = tview.NewFlex()
	container.Container.SetDirection(flexDirection)

	return &container
}

func (display *InteractiveDisplay) addWritableSection(title string, fixed int, proportion int) {
	newView := tview.NewTextView()

	newView.SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetBorder(true)

	newView.SetTitle(title)

	newView.SetChangedFunc(func() {
		display.app.ForceDraw()
	})

	// Add the view to the container
	display.pageStack.Page().WritableContainer.Sections = append(display.pageStack.Page().WritableContainer.Sections, newView)

	// Add the container to the flex container
	display.pageStack.Page().WritableContainer.Container.AddItem(newView, fixed, proportion, false)

}

func (display *InteractiveDisplay) currentPage() *Page {
	return display.pageStack.Page()
}

type PageElement struct {
	Element    tview.Primitive
	Fixed      int
	Proportion int
	Hidden     bool
	Focused    bool
}

type Page struct {
	Title    string
	Elements []*PageElement
	*WritableContainer
}

type InteractiveDisplay struct {
	pageStack *PageStack
	root      *tview.Flex
	app       *tview.Application
}

func NewInteractiveDisplay() *InteractiveDisplay {
	// Create the display
	display := &InteractiveDisplay{}

	// Initialize the page stack
	display.pageStack = &PageStack{}
	display.pageStack.Push()

	// Configure the root element
	display.root = tview.NewFlex()
	display.root.SetBorder(true)
	display.root.SetDirection(tview.FlexRow)
	display.root.SetTitle("[yellow]Interactive Display")

	// Configure app
	display.app = tview.NewApplication().SetRoot(display.root, true)
	display.app.EnableMouse(true)

	display.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// PreviousPage page
		if event.Key() == tcell.KeyEscape {
			display.PreviousPage()
		}
		// Exit app
		if event.Key() == tcell.KeyCtrlC {
			display.app.Stop()
		}
		return event
	})

	return display
}

func (display *InteractiveDisplay) getSection(index int) *tview.TextView {
	if display.currentPage().WritableContainer == nil {
		wc := NewWritableContainer(tview.FlexColumn)
		display.AddWritableContainer(wc, 0, 1)
	}

	// Create intermediary sections until the desired index
	currentLen := len(display.pageStack.Page().WritableContainer.Sections)
	if currentLen <= index {
		for i := currentLen; i <= index; i++ {
			display.addWritableSection("Section - "+strconv.Itoa(i), 0, 1)
		}
	}
	display.UpdateDisplay()
	return display.pageStack.Page().WritableContainer.Sections[index]
}

func (display *InteractiveDisplay) UpdateDisplay() {
	// Clear the screen
	display.root.Clear()

	if display.currentPage().Title == "" {
		display.root.SetTitle("[yellow]Interactive Display")
	} else {
		display.root.SetTitle(display.currentPage().Title)
	}

	// Add back the elements in the Page
	for _, element := range display.currentPage().Elements {
		if element.Hidden {
			continue
		}
		display.root.AddItem(element.Element, element.Fixed, element.Proportion, element.Focused)
	}
	if display.app != nil {
		display.app.ForceDraw()
	}
}

func (display *InteractiveDisplay) Enable() {

	if err := display.app.Run(); err != nil {
		panic(err)
	}

}

func (display *InteractiveDisplay) Print(buffer string) {

	_, err := fmt.Fprintf(tview.ANSIWriter(display.getSection(0)), "%s", buffer)
	if err != nil {
		panic(err)
	}

}

func (display *InteractiveDisplay) Println(buffer string) {
	display.Print(buffer)
	display.Print("\n")
}

// PrintPage Maybe add an option to print line numbers to the side automatically
func (display *InteractiveDisplay) PrintPage(index int, title string, buffer string) {

	display.getSection(index).SetTitle(title)

	_, err := fmt.Fprintf(tview.ANSIWriter(display.getSection(index)), "%s", buffer)
	if err != nil {
		panic(err)
	}

}

func (display *InteractiveDisplay) NewPage(title string) {
	display.pageStack.Push()
	display.currentPage().Title = title
	display.UpdateDisplay()
}

func (display *InteractiveDisplay) PreviousPage() {

	// Stop the app if the current page is the main menu
	if display.pageStack.Len() == 1 {
		display.Stop()
	}

	// Pop the current page and reload the last
	display.pageStack.Pop()
	display.UpdateDisplay()

}

func (display *InteractiveDisplay) AddElement(element *PageElement) {
	display.currentPage().Elements = append(display.currentPage().Elements, element)
}

func (display *InteractiveDisplay) AddWritableContainer(container *WritableContainer, fixed int, proportion int) {
	element := &PageElement{}
	element.Element = container.Container
	element.Fixed = fixed
	element.Proportion = proportion

	display.AddElement(element)
	display.currentPage().WritableContainer = container
}

func (display *InteractiveDisplay) Stop() {
	display.app.Stop()
	os.Exit(0)
}
