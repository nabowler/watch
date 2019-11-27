// +build !windows

package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/kr/pty"
)

var defaultShell string

func init() {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "bash"
	}
	// TODO: detect running shell. This is just the default shell.
	// `ps -p $$ -oargs=`, though this likely gives this program, so we'd need the parent
	// Also may which to try to run `$SHELL -cli echo test` and check for an error

	// I checked the man page and/or tested ash, bash, csh, dash, ksh, fish, tcsh, zsh
	// The man pages all listed -cli, except ash and zsh which don't specify -l, but
	// still support it as far as I can tell
	defaultShell = sh + " -cli"
}

func cmdOutput(cmd *exec.Cmd, buf *bytes.Buffer) error {
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	err = pty.InheritSize(os.Stdin, ptmx)
	if err != nil {
		return err
	}

	_, err = io.Copy(buf, ptmx)
	if err != nil {
		// Linux kernel return EIO when attempting to read from a master pseudo
		// terminal which no longer has an open slave. So ignore error here.
		// See https://github.com/kr/pty/issues/21
		if pathErr, ok := err.(*os.PathError); !ok || pathErr.Err != syscall.EIO {
			return err
		}
	}

	return nil
}
