package prompts

import (
	"fmt"
	"ndeploy/v2/internal/ssh"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func Passphrase(w fyne.Window) ssh.PassphrasePrompt {
	return func(path, fingerprint string, attempt int) ([]byte, error) {
		type answer struct {
			passphrase []byte
			ok         bool
		}

		answers := make(chan answer, 1)

		fyne.Do(func() {
			title := "Unlock" + filepath.Base(path)
			if attempt > 1 {
				title = fmt.Sprintf("Wrong passphrase, try again (attempt %d)", attempt)
			}

			entry := widget.NewPasswordEntry()
			entry.SetPlaceHolder("passphrase for " + filepath.Base(path))

			item := widget.NewFormItem("Passphrase", entry)
			item.HintText = fingerprint

			d := dialog.NewForm(title, "Unlock", "Cancel", []*widget.FormItem{item},
				func(ok bool) {
					answers <- answer{passphrase: []byte(entry.Text), ok: ok}
				}, w)

			entry.OnSubmitted = func(string) { d.Submit() }

			d.Resize(fyne.NewSize(460, 0))
			d.Show()
			w.Canvas().Focus(entry)
		})

		a := <-answers
		if !a.ok {
			return nil, ssh.ErrPromptCanceled
		}

		return a.passphrase, nil
	}
}

func ConfirmHostKey(w fyne.Window, host string, info *ssh.HostKeyInfo) bool {
	answers := make(chan bool, 1)

	fyne.Do(func() {

	})
}
