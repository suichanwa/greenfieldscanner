package ui

import (
	"cloud-storage/desktop/api"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func ShowFileList(client *api.Client, window fyne.Window) fyne.CanvasObject {
	var fileList []api.FileInfo
	var filteredList []api.FileInfo
	var refresh func()
	var list *widget.List
	var expandedID = -1

	createFileItem := func(file api.FileInfo, isExpanded bool) fyne.CanvasObject {
		// Basic info row
		basicInfo := container.NewHBox(
			widget.NewIcon(theme.DocumentIcon()),
			widget.NewLabelWithStyle(file.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(formatSize(file.Size)),
		)

		// Detailed info (shown when expanded)
		detailsContainer := container.NewVBox(
			widget.NewCard("Details",
				"",
				container.NewVBox(
					widget.NewLabel("Last Modified: "+formatTime(file.LastModified)),
					container.NewHBox(
						widget.NewButtonWithIcon("Download", theme.DownloadIcon(), nil),
						widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), nil),
					),
				),
			),
		)

		if !isExpanded {
			detailsContainer.Hide()
		}

		return container.NewVBox(basicInfo, detailsContainer)
	}

	list = widget.NewList(
		func() int {
			return len(filteredList)
		},
		func() fyne.CanvasObject {
			return createFileItem(api.FileInfo{}, false)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			file := filteredList[id]
			vbox := item.(*fyne.Container)
			basicInfo := vbox.Objects[0].(*fyne.Container)
			detailsContainer := vbox.Objects[1].(*fyne.Container)

			// Update basic info
			nameLabel := basicInfo.Objects[1].(*widget.Label)
			nameLabel.SetText(file.Name)

			sizeLabel := basicInfo.Objects[2].(*widget.Label)
			sizeLabel.SetText(formatSize(file.Size))

			// Update details container
			card := detailsContainer.Objects[0].(*widget.Card)
			cardContent := card.Content.(*fyne.Container)
			timeLabel := cardContent.Objects[0].(*widget.Label)
			timeLabel.SetText("Last Modified: " + formatTime(file.LastModified))

			buttons := cardContent.Objects[1].(*fyne.Container)
			downloadBtn := buttons.Objects[0].(*widget.Button)
			deleteBtn := buttons.Objects[1].(*widget.Button)

			downloadBtn.OnTapped = func() {
				downloadDialog := dialog.NewProgressInfinite("Downloading", "Downloading "+file.Name, window)
				downloadDialog.Show()

				go func() {
					err := client.DownloadFile(fmt.Sprint(file.ID), file.Name)
					downloadDialog.Hide()
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					dialog.ShowInformation("Success", "File downloaded successfully", window)
				}()
			}

			deleteBtn.OnTapped = func() {
				dialog.ShowConfirm("Delete File",
					"Are you sure you want to delete "+file.Name+"?",
					func(ok bool) {
						if ok {
							err := client.DeleteFile(fmt.Sprint(file.ID))
							if err != nil {
								dialog.ShowError(err, window)
								return
							}
							refresh()
							dialog.ShowInformation("Success", "File deleted successfully", window)
						}
					}, window)
			}

			// Show/hide details based on expanded state
			if id == expandedID {
				detailsContainer.Show()
			} else {
				detailsContainer.Hide()
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id == expandedID {
			expandedID = -1 // Collapse if clicking the same item
		} else {
			expandedID = id // Expand the clicked item
		}
		list.Refresh()
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search files...")
	searchEntry.OnChanged = func(text string) {
		if text == "" {
			filteredList = fileList
		} else {
			filteredList = nil
			searchLower := strings.ToLower(text)
			for _, file := range fileList {
				if strings.Contains(strings.ToLower(file.Name), searchLower) {
					filteredList = append(filteredList, file)
				}
			}
		}
		list.Refresh()
	}

	refresh = func() {
		files, err := client.ListFiles()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		fileList = files
		filteredList = files
		list.Refresh()
	}

	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("Upload", theme.UploadIcon(), func() {
			fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()

				path := reader.URI().Path()
				if len(path) > 0 && path[0] == '/' {
					path = path[1:]
				}

				uploadDialog := dialog.NewProgressInfinite("Uploading", "Uploading "+path, window)
				uploadDialog.Show()

				go func() {
					err = client.UploadFile(path)
					uploadDialog.Hide()
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					refresh()
					dialog.ShowInformation("Success", "File uploaded successfully", window)
				}()
			}, window)
			fd.Show()
		}),
		widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), refresh),
		searchEntry,
	)

	refresh()

	return container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewScroll(list),
	)
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatTime(timestamp string) string {
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return timestamp
}
