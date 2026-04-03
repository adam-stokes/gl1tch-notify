package main

import (
	_ "embed"

	"github.com/getlantern/systray"
)

//go:embed assets/icon.png
var iconData []byte

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip("gl1tch")

	statusItem := systray.AddMenuItem("○ Disconnected", "")
	statusItem.Disable()

	systray.AddSeparator()

	quitItem := systray.AddMenuItem("Quit", "Quit gl1tch-notify")

	// Channel for bus connection state updates.
	statusCh := make(chan bool, 4)

	// Start the BUSD listener.
	go busLoop(statusCh)

	// Handle UI updates and quit.
	go func() {
		for {
			select {
			case connected := <-statusCh:
				if connected {
					statusItem.SetTitle("● Connected")
				} else {
					statusItem.SetTitle("○ Disconnected")
				}
			case <-quitItem.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {}
