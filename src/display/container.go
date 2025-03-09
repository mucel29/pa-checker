package display

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

type WritableContainer struct {
	Container *tview.Flex
	Sections  []*tview.TextView

	inputCaptures []tview.Primitive
	synced        bool

	parent *Display
}

func NewWritableContainer(d *Display) *WritableContainer {
	container := WritableContainer{}
	container.Container = tview.NewFlex()
	container.Container.SetDirection(tview.FlexRow)
	container.parent = d

	container.Container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		for _, capture := range container.inputCaptures {
			if !capture.HasFocus() {
				capture.InputHandler()(event, nil)
			}
		}
		return event
	})

	container.Container.SetBorderPadding(1, 1, 1, 1)

	return &container
}

func (wc *WritableContainer) SetDirection(direction int) {
	wc.Container.SetDirection(direction)
}

func (wc *WritableContainer) WrapInput(primitive tview.Primitive) {
	if primitive.InputHandler() == nil {
		return
	}

	wc.inputCaptures = append(wc.inputCaptures, primitive)
}

func (wc *WritableContainer) AddPrimitive(primitive tview.Primitive, wrapInput bool, fixed int, proportion int) {
	wc.Container.AddItem(primitive, fixed, proportion, false)
	if wrapInput {
		wc.WrapInput(primitive)
	}
}

func (wc *WritableContainer) AddSection(title string, fixed int, proportion int) *tview.TextView {
	newView := tview.NewTextView()

	newView.SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetBorder(true)

	newView.SetChangedFunc(func() {
		wc.parent.App.ForceDraw()
	})

	newView.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {

		if action == tview.MouseScrollUp || action == tview.MouseScrollDown {
			if wc.synced {
				go func() {
					// Short delay to wait for the scroll action to apply
					time.Sleep(5 * time.Millisecond)

					wc.parent.App.QueueUpdateDraw(func() {
						for _, section := range wc.Sections {
							if section == newView {
								continue
							}
							section.ScrollTo(newView.GetScrollOffset())
						}
					})
				}()
			}
		}

		return action, event
	})

	if title != "" {
		newView.SetTitle(title)
	}

	wc.Sections = append(wc.Sections, newView)
	wc.Container.AddItem(newView, fixed, proportion, false)

	return newView
}

func (wc *WritableContainer) GetSection(index int) *tview.TextView {
	// Create intermediary sections until the desired index
	currentLen := len(wc.Sections)
	if currentLen <= index {
		for i := currentLen; i <= index; i++ {
			wc.AddSection("", 0, 1)
		}
	}
	return wc.Sections[index]
}

func (wc *WritableContainer) PrintIndex(index int, title string, buffer string) {
	if title == "$nb" {
		wc.GetSection(index).SetBorder(false)
	} else if title != "" {
		wc.GetSection(index).SetBorder(true).SetTitle(title)
	}

	_, err := fmt.Fprint(tview.ANSIWriter(wc.GetSection(index)), buffer)
	if err != nil {
		panic(err)
	}
}

func (wc *WritableContainer) Print(buffer string) {
	wc.PrintIndex(0, "", buffer)
}

func (wc *WritableContainer) Title(title string, align int) {
	if title != "" {
		wc.Container.SetTitle(title).SetTitleAlign(align)
		wc.Container.SetBorder(true)
	} else {
		wc.Container.SetBorder(false)
	}

}

func (wc *WritableContainer) SyncSections(sync bool) {
	wc.synced = sync
}

func (wc *WritableContainer) Clear() {
	wc.Container.Clear()
	wc.Sections = []*tview.TextView{}
	wc.Title("", 0)
}
