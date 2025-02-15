package main

import (
	"cloud-storage/desktop/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type CloudApp struct {
	client *api.Client
	window fyne.Window
	files  *widget.List
}

func main() {
	a := app.New()
	window := a.NewWindow("Cloud Storage")

	client := api.NewClient("http://localhost:8080")
	cloudApp := &CloudApp{
		client: client,
		window: window,
	}

	cloudApp.showLogin()
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}

func (a *CloudApp) showLogin() {
	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username", Widget: username},
			{Text: "Password", Widget: password},
		},
		OnSubmit: func() {
			err := a.client.Login(username.Text, password.Text)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			a.showMainView()
		},
	}

	a.window.SetContent(container.NewVBox(
		widget.NewLabel("Cloud Storage Login"),
		form,
	))
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

			err = a.client.UploadFile(reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}

			dialog.ShowInformation("Success", "File uploaded successfully", a.window)
		}, a.window)
		fd.Show()
	})

	a.window.SetContent(container.NewVBox(
		widget.NewLabel("Cloud Storage"),
		uploadBtn,
	))
}
