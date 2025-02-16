package main

import (
	"cloud-storage/desktop/api"
	"cloud-storage/desktop/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type CloudApp struct {
	client *api.Client
	window fyne.Window
}

func main() {
	a := app.NewWithID("com.suiseika.cloudstorage")
	window := a.NewWindow("Cloud Storage")

	client := api.NewClient("http://localhost:8080")
	cloudApp := &CloudApp{
		client: client,
		window: window,
	}

	// Use the login form from the UI package
	window.SetContent(ui.ShowLoginForm(window, func(username, password string) {
		err := cloudApp.client.Login(username, password)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		cloudApp.showMainView()
	}))

	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}

func (a *CloudApp) showMainView() {
	uploadBtn := widget.NewButton("Upload File", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			err = a.client.UploadFile(reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}

			dialog.ShowInformation("Success", "File uploaded successfully", a.window)
		}, a.window)
		fd.Show()
	})

	showFilesBtn := widget.NewButton("Show Files", func() {
		fileListUI := ui.ShowFileList(a.client, a.window)
		a.window.SetContent(fileListUI)
	})

	showSyncBtn := widget.NewButton("Sync Files", func() {
		syncUI := ui.ShowSyncScreen(a.client, a.window)
		a.window.SetContent(
			container.NewVBox(
				widget.NewLabel("Sync Files"),
				syncUI,
				widget.NewButton("Back", func() {
					a.showMainView()
				}),
			),
		)
	})

	mainContainer := container.NewVBox(
		widget.NewLabel("Cloud Storage"),
		uploadBtn,
		showFilesBtn,
		showSyncBtn,
	)

	a.window.SetContent(mainContainer)
}
