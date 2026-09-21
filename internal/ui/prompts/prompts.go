package prompts

import (
	"fmt"
	"ndeploy/v2/internal/ssh"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
		keyType := strings.ToUpper(strings.TrimPrefix(info.Key.Type(), "ssh-"))

		intro := fmt.Sprintf("ndeploy has not connected to %s before.", host)
		if info.KnownHosts == ssh.KnownHostsMismatch {
			intro = fmt.Sprintf("WARNING: %s presented a different key than the one in your ~/.ssh/known_hosts. Unless the node was reinstalled, someone may be intercepting the connection.", host)
		}

		introLabel := widget.NewLabel(intro)
		introLabel.Wrapping = fyne.TextWrapWord

		fingerprint := widget.NewLabelWithStyle(
			fmt.Sprintf("%s key fingerprint:\n%s", keyType, ssh.Fingerprint(info.Key)),
			fyne.TextAlignLeading,
			fyne.TextStyle{Monospace: true},
		)

		help := widget.NewLabel(fmt.Sprintf("To verify, run this on the node and compare the output:\nssh-keygen -lf /etc/ssh/ssh_host_%s_key.pub", strings.ToLower(keyType)))
		help.Wrapping = fyne.TextWrapWord

		d := dialog.NewCustomConfirm(
			"Trust this host?", "Trust", "Cancel",
			container.NewVBox(introLabel, fingerprint, help),
			func(ok bool) { answers <- ok }, w,
		)

		d.Resize(fyne.NewSize(520, 0))
		d.Show()
	})

	return <-answers
}

func ShowError(w fyne.Window, err error) {
	fyne.Do(func() {
		dialog.ShowError(err, w)
	})
}
