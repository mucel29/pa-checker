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

const (
	labelColor   = tcell.ColorWhite
	labelBgColor = tcell.ColorDefault
	fieldColor   = tcell.ColorWhite
	fieldBgColor = tcell.ColorBlue

	fieldSelColor   = tcell.ColorBlue
	fieldSelBgColor = tcell.ColorWhite

	labelWidth = 20
	fieldWidth = 30
)

func (m *Menu) displayOptions() {

	m.CurrentContainer().Clear()
	m.CurrentContainer().Title("Options", tview.AlignLeft)

	exPath := tview.NewInputField()
	exPath.SetLabel("Executable Path")
	exPath.SetText(utils.Config.ExecutablePath)
	exPath.SetFieldWidth(fieldWidth)

	sPath := tview.NewInputField()
	sPath.SetLabel("Source Path")
	sPath.SetText(utils.Config.SourcePath)
	sPath.SetFieldWidth(fieldWidth)

	oPath := tview.NewInputField()
	oPath.SetLabel("Output Path")
	oPath.SetText(utils.Config.OutputPath)
	oPath.SetFieldWidth(fieldWidth)

	rPath := tview.NewInputField()
	rPath.SetLabel("Ref Path")
	rPath.SetText(utils.Config.RefPath)
	rPath.SetFieldWidth(fieldWidth)

	form := tview.NewFlex()
	form.SetDirection(tview.FlexRow)
	form.AddItem(exPath, 0, 1, false)
	form.AddItem(sPath, 0, 1, false)
	form.AddItem(oPath, 0, 1, false)
	form.AddItem(rPath, 0, 1, false)

	items := []*tview.InputField{
		exPath, sPath, oPath, rPath,
	}

	itemIndex := 0
	firstLaunch := true

	for i, item := range items {
		item.SetFormAttributes(labelWidth, labelColor, labelBgColor, fieldColor, fieldBgColor)
		// Update selected item on click
		item.SetFocusFunc(func() {
			firstLaunch = false
			items[itemIndex].SetFormAttributes(labelWidth, labelColor, labelBgColor, fieldColor, fieldBgColor)
			itemIndex = i
			item.SetFormAttributes(labelWidth, labelColor, labelBgColor, fieldSelColor, fieldSelBgColor)
		})
	}

	// Stop the input fields from changing cursor the end / beginning
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			return nil
		}

		return event
	})

	m.CurrentContainer().AddInputCallback(func(event *tcell.EventKey) {

		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			oldItem := items[itemIndex]
			oldItem.SetFormAttributes(labelWidth, labelColor, labelBgColor, fieldColor, fieldBgColor)
		}

		if !firstLaunch {
			switch event.Key() {
			case tcell.KeyUp:
				itemIndex--
			case tcell.KeyDown:
				itemIndex++
			default:
				return
			}
		} else {
			firstLaunch = false
		}

		if itemIndex < 0 {
			itemIndex = len(items) - 1
		} else if itemIndex >= len(items) {
			itemIndex = 0
		}

		currItem := items[itemIndex]
		m.App.SetFocus(currItem)
		currItem.SetFormAttributes(labelWidth, labelColor, labelBgColor, fieldSelColor, fieldSelBgColor)

		m.App.ForceDraw()
	})

	// TODO: maybe make a better way to save the fields

	m.CurrentContainer().AddChangeCallback(func() {
		// Save the new values
		utils.Config.UserConfig.ExecutablePath = exPath.GetText()
		utils.Config.UserConfig.SourcePath = sPath.GetText()
		utils.Config.UserConfig.OutputPath = oPath.GetText()
		utils.Config.UserConfig.RefPath = rPath.GetText()

		// TODO: trigger manager

		// Save the current config
		utils.SaveUserConfig()
	})

	m.CurrentContainer().AddPrimitive(form, false, 0, 1)
	m.UpdateDisplay()
}

func (m *Menu) displayRef() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["ref_checker"].Display(m.Display)
}

func (m *Menu) displayStyle() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["style_checker"].Display(m.Display)
	m.CurrentContainer().WrapInput(m.CurrentContainer().Sections[0])
}

func (m *Menu) displayMemory() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["memory_checker"].Display(m.Display)
	m.CurrentContainer().WrapInput(m.CurrentContainer().Sections[0])
}

func (m *Menu) displayCommits() {
	m.CurrentContainer().Clear()
	checker_modules.AvailableModules["commit_checker"].Display(m.Display)
	m.CurrentContainer().WrapInput(m.CurrentContainer().Sections[0])
}

func (m *Menu) displayHome() {
	m.CurrentContainer().Clear()
	m.CurrentContainer().Title("Home", tview.AlignLeft)
	m.CurrentContainer().SetDirection(tview.FlexRow)
	// TODO: set this to info about the modules, redraw the page when the manager triggers
	m.CurrentContainer().PrintIndex(0, "Summary", "Placeholder for module summary")
	m.CurrentContainer().PrintIndex(1, "Tutorial", "TUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\nTUTORIAL\n")
}

func (m *Menu) createMainMenu() {

	mainContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	mainContainer.AddItem(buttonContainer, 0, 1, false)

	infoContainer := m.CurrentContainer().Container
	mainContainer.AddItem(infoContainer, 0, 3, false)

	m.displayHome()

	// TODO: fix bug where clicking the nav buttons they switch to the wrong tab

	m.nav = tview.NewList()
	m.nav.SetBorderPadding(1, 1, 1, 1)

	m.nav.SetTitle("Navigation").
		SetTitleAlign(tview.AlignLeft)

	// This prevents double selections
	clicked := false

	m.nav.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			clicked = true
		}

		return action, event
	})

	m.nav.SetSelectedFunc(func(_ int, _ string, _ string, _ rune) {
		// m.CurrentContainer().Title("", 0)
		// m.App.SetFocus(m.CurrentContainer().Container)
		// utils.Log(fmt.Sprintf("Navigation: %s", main))
	})

	m.nav.SetChangedFunc(func(_ int, _ string, _ string, _ rune) {
		// utils.Log("changed func")
		if clicked {
			clicked = false
			return
		}
		if m.launched {
			m.nav.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
		}
	})

	m.nav.AddItem("home", "", 0, func() {
		m.displayHome()
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
			// IDK if this should be called, like save the options and stuff
			// m.CurrentContainer().TriggerChange()
			m.PreviousPage()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			if m.nav.HasFocus() {
				m.App.SetFocus(m.CurrentContainer().Container)
			} else if m.CurrentPageIndex() == 0 {
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
