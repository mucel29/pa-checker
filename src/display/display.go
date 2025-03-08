package display

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strconv"
)

//type Display interface {
//	Enable()
//	Print(buffer string)
//	Println(buffer string)
//	PrintPage(index int, title string, buffer string)
//	ReadLine() string
//	IsInteractive() bool
//}
//
//func (bd *BasicDisplay) IsInteractive() bool {
//	return false
//}
//
//func (id *Display) IsInteractive() bool {
//	return true
//}

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

func (display *Display) addWritableSection(title string, fixed int, proportion int) {
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

func (display *Display) currentPage() *Page {
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

type Display struct {
	pageStack *PageStack
	root      *tview.Flex
	app       *tview.Application
}

func NewInteractiveDisplay() *Display {
	// Create the display
	display := &Display{}

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

func (display *Display) getSection(index int) *tview.TextView {
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

func (display *Display) UpdateDisplay() {
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

func (display *Display) Enable() {

	if err := display.app.Run(); err != nil {
		panic(err)
	}

}

func (display *Display) Print(buffer string) {

	_, err := fmt.Fprintf(tview.ANSIWriter(display.getSection(0)), "%s", buffer)
	if err != nil {
		panic(err)
	}

}

func (display *Display) Println(buffer string) {
	display.Print(buffer)
	display.Print("\n")
}

// PrintPage Maybe add an option to print line numbers to the side automatically
func (display *Display) PrintPage(index int, title string, buffer string) {

	display.getSection(index).SetTitle(title)

	_, err := fmt.Fprintf(tview.ANSIWriter(display.getSection(index)), "%s", buffer)
	if err != nil {
		panic(err)
	}

}

func (display *Display) NewPage(title string) {
	display.pageStack.Push()
	display.currentPage().Title = title
	display.UpdateDisplay()
}

func (display *Display) PreviousPage() {

	// Stop the app if the current page is the main menu
	if display.pageStack.Len() == 1 {
		display.Stop()
	}

	// Pop the current page and reload the last
	display.pageStack.Pop()
	display.UpdateDisplay()

}

func (display *Display) AddElement(element *PageElement) {
	display.currentPage().Elements = append(display.currentPage().Elements, element)
}

func (display *Display) AddWritableContainer(container *WritableContainer, fixed int, proportion int) {
	element := &PageElement{}
	element.Element = container.Container
	element.Fixed = fixed
	element.Proportion = proportion

	display.AddElement(element)
	display.currentPage().WritableContainer = container
}

func (display *Display) Stop() {
	display.app.Stop()
	os.Exit(0)
}

func (display *Display) ReadLine() string {
	// For interactive display, we don't want to interrupt the app flow
	// Just return an empty string as a fallback
	return "1" // Default to checking the first file
}

// Add this method to support the side-by-side display
func (display *Display) NewWritableContainer(flexDirection int) *WritableContainer {
	return NewWritableContainer(flexDirection)
}

// New method to handle file selection via keyboard
func (display *Display) SetupFileSelectionHandler(callback func(int), validFiles []int) {
	// Create a map of valid file numbers for quick lookup
	validFilesMap := make(map[int]bool)
	for _, num := range validFiles {
		validFilesMap[num] = true
	}

	// Add an input capture handler that checks for numeric keys
	prevHandler := display.app.GetInputCapture()
	display.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if a number key was pressed
		if event.Key() >= tcell.KeyRune && event.Key() <= tcell.KeyRune {
			r := event.Rune()
			if r >= '0' && r <= '9' {
				fileNum := int(r - '0')
				// Check if this is a valid file number
				if validFilesMap[fileNum] {
					// Call the callback with the selected file number
					callback(fileNum)
					return nil // Consume the event
				}
			}
		}

		// Pass the event to the previous handler
		if prevHandler != nil {
			return prevHandler(event)
		}
		return event
	})
}

// Add this new method to allow setting custom content for a page
func (display *Display) SetCustomContent(pageIndex int, content tview.Primitive) {
	// Ensure the page exists
	if display.currentPage().WritableContainer == nil {
		wc := NewWritableContainer(tview.FlexColumn)
		display.AddWritableContainer(wc, 0, 1)
	}

	// Clear existing sections at this index if any
	if pageIndex < len(display.pageStack.Page().WritableContainer.Sections) {
		display.pageStack.Page().WritableContainer.Container.RemoveItem(
			display.pageStack.Page().WritableContainer.Sections[pageIndex])
	}

	// Create a new page element for the custom content
	element := &PageElement{
		Element:    content,
		Proportion: 1,
		Focused:    true,
	}

	// Replace or add the element to the page
	if pageIndex < len(display.pageStack.Page().Elements) {
		display.pageStack.Page().Elements[pageIndex] = element
	} else {
		display.AddElement(element)
	}

	// Update the display
	display.UpdateDisplay()
}
