package main

import (
	checkermodules "checker-pa/src/checker-modules"
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/utils"
	_ "embed"
	"flag"
	"github.com/rivo/tview"
	"sync"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUserConfigStr string

func main() {

	utils.InitConfig(defaultUserConfigStr, moduleConfigStr)
	useInteractive := flag.Bool("i", false, "Interactive mode")

	var wg sync.WaitGroup

	// TODO: check for user config

	m, err := manager.NewManager()

	if err != nil {
		panic(err)
	}

	flag.Parse()

	m.Run()

	var d display.Display
	if *useInteractive {

		utils.Log("Interactive Display")
		iDisplay := display.NewInteractiveDisplay()
		d = iDisplay

		checkerButton := tview.NewButton("Checker").SetSelectedFunc(func() {
			iDisplay.NewPage("Checker")

			refButton := tview.NewButton("Ref checker").SetSelectedFunc(func() {
				iDisplay.NewPage("Ref checker")
				iDisplay.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["ref_checker"].Details(d)
			})
			memoryButton := tview.NewButton("Memory checker").SetSelectedFunc(func() {
				iDisplay.NewPage("Memory checker")
				iDisplay.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["memory_checker"].Details(d)
			})
			commitButton := tview.NewButton("Commit checker").SetSelectedFunc(func() {
				iDisplay.NewPage("Commit checker")
				iDisplay.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["commit_checker"].Details(d)
			})

			styleButton := tview.NewButton("Style checker").SetSelectedFunc(func() {
				iDisplay.NewPage("Style checker")
				iDisplay.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["style_checker"].Details(d)
			})

			buttonContainer := tview.NewFlex().SetDirection(tview.FlexRow)
			buttonContainer.
				AddItem(refButton, 0, 1, false).
				AddItem(styleButton, 0, 1, false).
				AddItem(memoryButton, 0, 1, false).
				AddItem(commitButton, 0, 1, false)

			iDisplay.AddElement(&display.PageElement{Element: buttonContainer, Proportion: 1})

			iDisplay.UpdateDisplay()

		})

		optionsButton := tview.NewButton("Options").SetSelectedFunc(func() {
			iDisplay.NewPage("Options")

			form := tview.NewForm()
			form.AddInputField("Source Path", utils.Config.SourcePath, 0, nil, nil)
			form.AddInputField("Executable Path", utils.Config.ExecutablePath, 0, nil, nil)
			form.AddInputField("Output Path", utils.Config.OutputPath, 0, nil, nil)

			// TODO: change config settings when going back

			iDisplay.AddElement(&display.PageElement{Element: form, Proportion: 1})
			iDisplay.UpdateDisplay()
		})

		mainContainer := tview.NewFlex().SetDirection(tview.FlexRow)
		mainContainer.
			AddItem(checkerButton, 0, 1, false).
			AddItem(optionsButton, 0, 1, false).
			AddItem(tview.NewBox(), 0, 7, false)

		iDisplay.AddElement(&display.PageElement{Element: mainContainer, Proportion: 1})

		iDisplay.UpdateDisplay()

	} else {
		d = &display.BasicDisplay{}
		utils.Log("Basic Display")

		for _, module := range m.Modules {
			module.Details(d)
		}

	}

	d.Enable()
	wg.Wait()
}
