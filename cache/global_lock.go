package cache

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

var isStarted bool

// SystemLock starts a system rather than application lock. This way multiple
// app instances don't cause race conditions when working in the cache.
func SystemLock() error {
	if isStarted {
		return nil
	}
	err := waitOnLock()
	if err != nil {
		return err
	}
	err = startLock()
	isStarted = true
	return err
}

// SystemUnlock removes the system wide Glide cache lock.
func SystemUnlock() {
	lockdone <- struct{}{}
	os.Remove(lockFileName)
}

var lockdone = make(chan struct{}, 1)

type lockdata struct {
	Comment string `json:"comment"`
	Pid     int    `json:"pid"`
	Time    string `json:"time"`
}

var lockFileName = filepath.Join(gpath.Home(), "lock.json")

// Write a lock for now.
func writeLock() error {
	ld := &lockdata{
		Comment: "File managed by Glide (https://glide.sh)",
		Pid:     os.Getpid(),
		Time:    time.Now().Format(time.RFC3339Nano),
	}

	out, err := json.Marshal(ld)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(lockFileName, out, 0755)
	return err
}

func startLock() error {
	err := writeLock()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-lockdone:
				return
			default:
				time.Sleep(10 * time.Second)
				err := writeLock()
				if err != nil {
					msg.Die("Error using Glide lock: %s", err)
				}
			}
		}
	}()

	return nil
}

func waitOnLock() error {
	var announced bool
	for {
		fi, err := os.Stat(lockFileName)
		if err != nil && os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}

		diff := time.Now().Sub(fi.ModTime())
		if diff.Seconds() > 15 {
			return nil
		}

		if !announced {
			announced = true
			msg.Info("Waiting on Glide global cache access")
		}

		// Check on the lock file every 15 seconds.
		// TODO(mattfarina): should this be a different length?
		time.Sleep(time.Second)
	}
}
