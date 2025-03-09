package menu

import (
	checker_modules "checker-pa/src/checker-modules"
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Menu struct {
	*display.Display
	*manager.Manager
	nav      *tview.List
	launched bool
}

func (m *Menu) displayOptions() {

	m.CurrentContainer().Clear()
	m.CurrentContainer().Title("Options", tview.AlignLeft)

	form := tview.NewForm()
	form.AddInputField("Source Path", utils.Config.SourcePath, 0, nil, nil)
	form.AddInputField("Executable Path", utils.Config.ExecutablePath, 0, nil, nil)
	form.AddInputField("Output Path", utils.Config.OutputPath, 0, nil, nil)

	// TODO: change config settings when going back

	m.CurrentContainer().AddPrimitive(form, true, 0, 1)
	m.UpdateDisplay()
}

func (m *Menu) displayRef() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["ref_checker"].Display(m.Display)
}

func (m *Menu) displayStyle() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["style_checker"].Display(m.Display)
}

func (m *Menu) displayMemory() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["memory_checker"].Display(m.Display)

}

func (m *Menu) displayCommits() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["commit_checker"].Display(m.Display)
}

func (m *Menu) displayTutorial() {
	m.CurrentContainer().Clear()
	m.CurrentContainer().Title("Tutorial", tview.AlignLeft)
	m.CurrentContainer().Print("TUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\n")
}

func (m *Menu) createMainMenu() {

	mainContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	mainContainer.AddItem(buttonContainer, 0, 1, false)

	infoContainer := m.CurrentContainer().Container
	mainContainer.AddItem(infoContainer, 0, 3, false)

	m.displayTutorial()

	m.nav = tview.NewList()

	m.nav.SetTitle("Navigation").
		SetTitleAlign(tview.AlignLeft)

	m.nav.SetSelectedFunc(func(_ int, _ string, _ string, _ rune) {
		// m.CurrentContainer().Title("", 0)
		// m.App.SetFocus(m.CurrentContainer().Container)
	})

	m.nav.SetChangedFunc(func(_ int, _ string, _ string, _ rune) {
		// utils.Log("changed func")
		if m.launched {
			m.nav.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 'e', tcell.ModNone), nil)
		}
	})

	m.nav.AddItem("ref", "", 0, func() {
		m.displayRef()
	})
	m.nav.AddItem("style", "", 0, func() {
		m.displayStyle()
	})
	m.nav.AddItem("memory", "", 0, func() {
		m.displayMemory()
	})
	m.nav.AddItem("commit", "", 0, func() {
		m.displayCommits()
	})
	m.nav.AddItem("options", "", 0, func() {
		m.displayOptions()
	})

	m.nav.SetBorder(true)

	buttonContainer.AddItem(m.nav, 0, 4, true)

	scoreContainer := tview.NewTextView().SetText("80 / 100")
	scoreContainer.SetTitle("Score").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)

	buttonContainer.AddItem(scoreContainer, 0, 1, true)

	infoBox := tview.NewTextView().SetText("mucel\nCiprian\nsteffe\nHoricuz").SetTextAlign(tview.AlignCenter)
	infoBox.SetTitle("Credits").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)
	infoBox.SetBorder(true)

	buttonContainer.AddItem(infoBox, 0, 1, false)

	m.AddElement(&display.PageElement{Element: mainContainer, Proportion: 1, Focused: false})
	m.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.PreviousPage()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			if m.nav.HasFocus() {
				m.App.SetFocus(m.CurrentContainer().Container)
			} else {
				m.App.SetFocus(m.nav)
			}

			return nil
		}

		return event
	})
	m.App.SetFocus(m.nav)

}

func (m *Menu) Launch() {
	m.createMainMenu()
	m.launched = true
	m.UpdateDisplay()
}
