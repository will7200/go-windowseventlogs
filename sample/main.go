package main

import (
	"fmt"
	"os"
	"path/filepath"

	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	. "github.com/will7200/go-windowsEventLogs"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	source     = "testProgram"
	addKeyName = `SYSTEM\CurrentControlSet\Services\EventLog\Application`
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())
	pwd, _ := os.Getwd()
	eventSourceFile := filepath.Join(pwd, "ExampleMessageFile.txt")
	exists := checkIfExists()
	if !exists {
		err := eventlog.Install(source, eventSourceFile, true, 4)
		if err != nil {
			log.Fatal(err)
		}
	}
	//eventlog.Remove(source)
}
func checkIfExists() bool {
	addKey := fmt.Sprintf(`%s\%s`, addKeyName, source)
	log.Infof("Looking for %s", addKey)
	appkey, err := registry.OpenKey(registry.LOCAL_MACHINE, addKey, registry.QUERY_VALUE)
	if err != nil {
		log.Debug(err)
		return false
	}
	t, _ := appkey.ReadValueNames(10)
	log.Debug(t)
	_, _, err = appkey.GetIntegerValue("CustomSource")
	if err != nil {
		log.Debug(err)
	}
	return true
	//log.Infof("Key found %d", s)
}
func main() {
	log.Info("Installed Correctly")
	t, e := eventlog.Open(source)
	if e != nil {
		log.Fatal(e)
	}
	t.Info(1, "TESTING FROM SOURCE")
	tt, ee := OpenEventLog(source)
	if ee != nil {
		log.Fatal(e)
	}
	tt.SetReadFlags(EVENTLOG_SEQUENTIAL_READ | EVENTLOG_BACKWARDS_READ)
	tt.ReadEventLog(0, 1000)
	tt.Print(0, 1000)
}
