package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

var (
	mntSource = os.Getenv("MNT_SOURCE")
	mntTarget = os.Getenv("MNT_TARGET")
	mntFSType = os.Getenv("MNT_FSTYPE")
)

func main() {
	// mount desired device and directory
	err := syscall.Mount(mntSource, mntTarget, mntFSType, syscall.MS_RELATIME, "")
	if err != nil {
		log.Fatal("error in mounting", err)
	}

	defer func(target string) {
		if err := syscall.Unmount(target, syscall.MNT_FORCE); err != nil {
			log.Println("error in unmounting", err)
		}
	}(mntTarget)

	// run rest-server command
	restserver := exec.Command(
		"/usr/local/bin/rest-server",
		os.Args...,
	)

	restserver.Env = append(restserver.Env, os.Environ()...)
	restserver.Stdin = os.Stdin
	restserver.Stdout = os.Stdout
	restserver.Stderr = os.Stderr
	err = restserver.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		} else {
			fmt.Fprintf(os.Stderr, "%v: %v\n", restserver.Args, err)
		}
	}

}
