package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
	"github.com/go-ini/ini"
	"github.com/kr/pty"
	gossh "golang.org/x/crypto/ssh"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	ssh.Handle(func(s ssh.Session) {
		if s.User() == "gui" {
			pwd, err := os.Executable()
			if err != nil {
				panic(err)
			}
			authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
			/*app := tview.NewApplication()
			modal := tview.NewModal().
				SetText("Akmey recognizes the key : " + string(authorizedKey)).
				AddButtons([]string{"Ok", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Ok" {
						app.Stop()
					}
				})
			if err := app.SetRoot(modal, false).SetFocus(modal).Run(); err != nil {
				panic(err)
			}*/
			cmd := exec.Command(cfg.Section("ssh").Key("uiexec").String(), string(authorizedKey))
			ptyReq, winCh, isPty := s.Pty()
			if isPty {
				cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
				cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", pwd))
				f, err := pty.Start(cmd)
				if err != nil {
					panic(err)
				}
				go func() {
					for win := range winCh {
						setWinsize(f, win.Width, win.Height)
					}
				}()
				go func() {
					io.Copy(f, s) // stdin
				}()
				io.Copy(s, f) // stdout
			} else {
				io.WriteString(s, "No PTY requested.\n")
				s.Exit(1)
			}
		} else {
			io.WriteString(s, "Wrong user, i accept `api` and `gui`.\n")
		}
	})

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true // allow all keys, or use ssh.KeysEqual() to compare against known keys
	})

	log.Println("starting ssh server on address : " + cfg.Section("ssh").Key("listen").String())
	log.Fatal(ssh.ListenAndServe(cfg.Section("ssh").Key("listen").String(), nil, ssh.HostKeyFile(cfg.Section("ssh").Key("hostkey").String()), publicKeyOption))
}
