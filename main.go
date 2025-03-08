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

	if *useInteractive {

		utils.Log("Interactive Display")
		d := display.NewInteractiveDisplay()

		checkerButton := tview.NewButton("Checker").SetSelectedFunc(func() {
			d.NewPage("Checker")

			refButton := tview.NewButton("Ref checker").SetSelectedFunc(func() {
				d.NewPage("Ref checker")
				d.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["ref_checker"].Display(d)
			})
			memoryButton := tview.NewButton("Memory checker").SetSelectedFunc(func() {
				d.NewPage("Memory checker")
				d.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["memory_checker"].Display(d)
			})
			commitButton := tview.NewButton("Commit checker").SetSelectedFunc(func() {
				d.NewPage("Commit checker")
				d.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["commit_checker"].Display(d)
			})

			styleButton := tview.NewButton("Style checker").SetSelectedFunc(func() {
				d.NewPage("Style checker")
				d.AddWritableContainer(display.NewWritableContainer(tview.FlexRow), 0, 1)
				checkermodules.AvailableModules["style_checker"].Display(d)
			})

			buttonContainer := tview.NewFlex().SetDirection(tview.FlexRow)
			buttonContainer.
				AddItem(refButton, 0, 1, false).
				AddItem(styleButton, 0, 1, false).
				AddItem(memoryButton, 0, 1, false).
				AddItem(commitButton, 0, 1, false)

			d.AddElement(&display.PageElement{Element: buttonContainer, Proportion: 1})

			d.UpdateDisplay()

		})

		optionsButton := tview.NewButton("Options").SetSelectedFunc(func() {
			d.NewPage("Options")

			form := tview.NewForm()
			form.AddInputField("Source Path", utils.Config.SourcePath, 0, nil, nil)
			form.AddInputField("Executable Path", utils.Config.ExecutablePath, 0, nil, nil)
			form.AddInputField("Output Path", utils.Config.OutputPath, 0, nil, nil)

			// TODO: change config settings when going back

			d.AddElement(&display.PageElement{Element: form, Proportion: 1})
			d.UpdateDisplay()
		})

		mainContainer := tview.NewFlex().SetDirection(tview.FlexRow)
		mainContainer.
			AddItem(checkerButton, 0, 1, false).
			AddItem(optionsButton, 0, 1, false).
			AddItem(tview.NewBox(), 0, 7, false)

		d.AddElement(&display.PageElement{Element: mainContainer, Proportion: 1})

		d.UpdateDisplay()

		d.Enable()

	} else {
		utils.Log("Basic Display")

		for _, module := range m.Modules {
			module.Dump()
		}

	}

	wg.Wait()
}
