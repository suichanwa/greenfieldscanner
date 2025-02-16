package ui

import (
	"cloud-storage/desktop/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowSyncScreen(client *api.Client, window fyne.Window) fyne.CanvasObject {
	statusLabel := widget.NewLabel("Not Synced")

	syncButton := widget.NewButton("Sync", func() {
		err := client.Sync()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		statusLabel.SetText("Synced Successfully")
	})

	return container.NewVBox(
		widget.NewLabel("Sync Files"),
		syncButton,
		statusLabel,
	)
}
