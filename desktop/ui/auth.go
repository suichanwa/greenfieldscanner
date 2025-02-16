package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowLoginForm creates and returns a login form.
// The onLogin callback is invoked with the username and password when the login button is clicked.
func ShowLoginForm(window fyne.Window, onLogin func(username, password string)) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	form := widget.NewForm(
		widget.NewFormItem("Username", usernameEntry),
		widget.NewFormItem("Password", passwordEntry),
	)

	loginButton := widget.NewButton("Login", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			dialog.ShowError(fmt.Errorf("please enter username and password"), window)
			return
		}
		onLogin(username, password)
	})

	content := container.NewVBox(
		widget.NewLabel("Login"),
		form,
		loginButton,
	)
	return content
}
