package media_commands

import (
	"os/exec"
)

func PlayFile(file string) {
	vlcBin, err := exec.LookPath("vlc")
	if err != nil {
		panic(err)
	}

	vlcCommand := exec.Command(vlcBin, file)
	vlcCommand.Start()
}
