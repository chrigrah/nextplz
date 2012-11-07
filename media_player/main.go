package media_commands

import (
	"fmt"
	"net"
//	"time"
	"flag"
	"strings"
	"os/exec"
)

type MediaPlayer interface {
	PlayFile(file string) error
	Disconnect()
}

type MediaPlayerInitInfo struct {
	Executable string
	Arguments string
}

func InitMediaPlayerFlagParser() *MediaPlayerInitInfo{
	var info MediaPlayerInitInfo
	flag.StringVar(&info.Executable, "exe", "", "The name of the media player executable (must be on system path)")
	flag.StringVar(&info.Arguments, "args", "", "Arguments to be passed to the media player (space separates, no escaping I'm afraid)")
	return &info
}

func (info *MediaPlayerInitInfo) CreateMediaPlayer() (MediaPlayer, error) {
	if info.Executable == "" {
		return CreateDefaultMediaPlayer()
	}

	executable, err := exec.LookPath(info.Executable)
	if err != nil {
		return nil, err
	}

	var mp CustomMediaPlayer
	mp.Executable = executable
	if info.Arguments == "" {
		mp.Args = make([]string, 1)
	} else {
		static_args := strings.Split(info.Arguments, " ")
		mp.Args = make([]string, len(static_args) + 1)
		copy(mp.Args, static_args)
	}

	return &mp, nil

}

func CreateDefaultMediaPlayer() (MediaPlayer, error) {
	executable, err := exec.LookPath("vlc")
	if err != nil {
		return nil, err
	}

	return MediaPlayer(&VLC{executable}), nil
}

type CustomMediaPlayer struct {
	Executable string
	Args []string
}

func (mp *CustomMediaPlayer) PlayFile(file string) error {
	var command exec.Cmd
	command.Path = mp.Executable
	mp.Args[len(mp.Args)-1] = file
	command.Args = mp.Args

	command.Start()
	return nil
}

func (mp *CustomMediaPlayer) Disconnect() {
}

////////////////////////////////////////////////////////////////////////
type VLC struct {
	executable string
}

func (vlc *VLC) TryQueue(file string) error {
	vlcConn, err := net.Dial("tcp", "localhost:47246")
	if err != nil {
		return err
	}

	result := make([]byte, 1024)
	err = vlc_rc_exec(vlcConn, fmt.Sprintf("add %s\n", file), result)
	if err != nil { return err; }

	vlcConn.Close()
	return nil
}

func (vlc *VLC) PlayFile(file string) error {
	err := vlc.TryQueue(file)
	if err == nil {
		return nil
	}

	command := exec.Command(vlc.executable, "--extraintf", "rc", "--rc-host", "127.0.0.1:47246", file)
	command.Start()

	return nil
}

func (vlc *VLC) Disconnect() {
}

func (vlc *VLC) Pause() error {
	vlcConn, err := net.Dial("tcp", "localhost:47246")
	if err != nil {
		return err
	}

	result := make([]byte, 1024)
	err = vlc_rc_exec(vlcConn, "pause\n", result)
	if err != nil { return err; }

	vlcConn.Close()
	return nil
}

func vlc_rc_exec(rc net.Conn, cmd string, result []byte) error {
	_, err := fmt.Fprint(rc, cmd)
	if err != nil {
		return err
	}

	if result != nil {
		_, err = rc.Read(result)
		return err
	}

	return nil
}
