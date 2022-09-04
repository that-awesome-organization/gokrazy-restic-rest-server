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
			if err := mount(); err != nil {
				log.Println(err)
				os.Exit(125)
			}
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
		"/user/rest-server",
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
	source := mntSource
	log.Printf("initial mount parameters - mntSource: %q, mntTarget: %q, mntFSType: %q, mntData: %q", source, mntTarget, mntFSType, mntData)
	// get actual devices from UUID string
	if strings.HasPrefix(mntSource, "UUID=") {
		devices, err := getDevices(mntSource)
		if err != nil {
			return fmt.Errorf("error getting devices: %w", err)
		}

		// select first device to use it as source
		source = devices[0]
		if mntFSType == "btrfs" {
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
	}
	log.Printf("final mount parameters - mntSource: %q, mntTarget: %q, mntFSType: %q, mntData: %q", source, mntTarget, mntFSType, mntData)
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
	c := exec.Command("/usr/local/bin/blkid")
	c.Env = append(c.Env, "PATH=/usr/local/bin:/perm/bin:/usr/bin")
	o, err := c.Output()
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
		if !strings.Contains(sc.Text(), source) {
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
