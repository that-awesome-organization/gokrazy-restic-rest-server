package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var (
	mntSource = os.Getenv("MNT_SOURCE")
	mntTarget = os.Getenv("MNT_TARGET")
	mntFSType = os.Getenv("MNT_FSTYPE")
	mntData   = os.Getenv("MNT_DATA")
)

func main() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	// catch sigterm
	go func() {
		sig := <-sigs
		log.Println("got signal", sig)
		done <- true
	}()

	// run the command
	go func() {
		// mount desired device and directory
		if mntSource != "" || mntTarget != "" || mntFSType != "" {
			mount()
		}
		err := run()
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}()

	// wait for signal
	<-done
	unmount()
}

func run() error {
	// run rest-server command
	restserver := exec.Command(
		"/usr/local/bin/rest-server",
		os.Args[1:]...,
	)

	restserver.Env = append(restserver.Env, os.Environ()...)
	restserver.Stdin = os.Stdin
	restserver.Stdout = os.Stdout
	restserver.Stderr = os.Stderr
	err := restserver.Run()
	if err != nil {
		return err
	}
	return nil
}

func mount() error {
	log.Printf("initial mount parameters - mntSource: %q, mntTarget: %q, mntFSType: %q, mntData: %q", mntSource, mntTarget, mntFSType, mntData)
	source := mntSource
	// get actual devices from UUID string
	if strings.HasPrefix(mntSource, "UUID=") {
		devices, err := getDevices(mntSource)
		if err != nil {
			return fmt.Errorf("error getting devices: %w", err)
		}
		source = devices[0]
		mntDataStrings := []string{}
		for _, d := range devices {
			mntDataStrings = append(mntDataStrings, "device="+d)
		}
		if mntData == "" {
			mntData = strings.Join(mntDataStrings, ",")
		} else {
			mntData = mntData + "," + strings.Join(mntDataStrings, ",")
		}
	}
	log.Printf("final mount parameters - mntSource: %q, mntTarget: %q, mntFSType: %q, mntData: %q", mntSource, mntTarget, mntFSType, mntData)
	err := syscall.Mount(source, mntTarget, mntFSType, syscall.MS_RELATIME, mntData)
	if err != nil {
		return fmt.Errorf("error in mounting: %w", err)
	}
	return nil
}

func unmount() {
	if err := syscall.Unmount(mntTarget, syscall.MNT_FORCE); err != nil {
		log.Println("error in unmounting", err)
	}
}

func getDevices(source string) ([]string, error) {
	o, err := exec.Command("blkid").Output()
	if err != nil {
		return nil, fmt.Errorf("error running blkid: %w", err)
	}

	sc := bufio.NewScanner(bytes.NewBuffer(o))
	devices := []string{}
	for sc.Scan() {
		if sc.Err() == io.EOF {
			break
		}
		if sc.Text() == "" {
			continue
		}
		if !strings.Contains(source, sc.Text()) {
			continue
		}

		splitString := strings.Split(sc.Text(), ":")
		if len(splitString) < 2 {
			return nil, fmt.Errorf("invalid blkid string %q", sc.Text())
		}

		devices = append(devices, splitString[0])
	}

	return devices, nil
}
