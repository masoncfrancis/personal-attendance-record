package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const launchAgentLabel = "com.personal-attendance-record"

type scheduleTime struct {
	hour   int
	minute int
}

func (t scheduleTime) String() string {
	return fmt.Sprintf("%02d:%02d", t.hour, t.minute)
}

type scheduleTimesFlag struct {
	values []scheduleTime
}

func (f *scheduleTimesFlag) String() string {
	values := make([]string, len(f.values))
	for i, value := range f.values {
		values[i] = value.String()
	}
	return strings.Join(values, ",")
}

func (f *scheduleTimesFlag) Set(value string) error {
	parsed, err := parseScheduleTime(value)
	if err != nil {
		return err
	}
	for _, existing := range f.values {
		if existing == parsed {
			return fmt.Errorf("duplicate schedule time %q", value)
		}
	}
	f.values = append(f.values, parsed)
	return nil
}

func handleSchedule(args []string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("schedule commands are only supported on macOS")
	}
	if len(args) == 0 {
		return errors.New("missing schedule command (expected install, status, or uninstall)")
	}

	switch args[0] {
	case "install":
		return installSchedule(args[1:])
	case "status":
		if len(args) != 1 {
			return errors.New("schedule status does not accept arguments")
		}
		return scheduleStatus()
	case "uninstall":
		if len(args) != 1 {
			return errors.New("schedule uninstall does not accept arguments")
		}
		return uninstallSchedule()
	default:
		return fmt.Errorf("unknown schedule command %q", args[0])
	}
}

func installSchedule(args []string) error {
	flags := flag.NewFlagSet("schedule install", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var times scheduleTimesFlag
	flags.Var(&times, "time", "Daily run time in HH:MM format (repeatable)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if len(times.values) == 0 {
		return errors.New("at least one --time value is required")
	}
	sort.Slice(times.values, func(i, j int) bool {
		if times.values[i].hour == times.values[j].hour {
			return times.values[i].minute < times.values[j].minute
		}
		return times.values[i].hour < times.values[j].hour
	})

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	plistPath, err := launchAgentPath()
	if err != nil {
		return err
	}
	logDir, err := launchLogDirectory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("create LaunchAgents directory: %w", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	if err := unloadSchedule(); err != nil {
		return err
	}
	if err := writeLaunchAgent(plistPath, executable, times.values, logDir); err != nil {
		return err
	}
	if _, err := runLaunchctl("bootstrap", launchctlDomain(), plistPath); err != nil {
		return fmt.Errorf("load LaunchAgent: %w", err)
	}
	fmt.Printf("Installed %s at %s\n", launchAgentLabel, plistPath)
	return nil
}

func scheduleStatus() error {
	output, err := runLaunchctl("print", launchctlDomain()+"/"+launchAgentLabel)
	if err != nil {
		return fmt.Errorf("LaunchAgent %s is not loaded: %s", launchAgentLabel, launchctlError(err, output))
	}
	fmt.Print(string(output))
	return nil
}

func uninstallSchedule() error {
	if err := unloadSchedule(); err != nil {
		return err
	}
	plistPath, err := launchAgentPath()
	if err != nil {
		return err
	}
	if err := os.Remove(plistPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove LaunchAgent plist: %w", err)
	}
	fmt.Printf("Uninstalled %s\n", launchAgentLabel)
	return nil
}

func unloadSchedule() error {
	target := launchctlTarget()
	output, err := runLaunchctl("bootout", target)
	if err == nil {
		return nil
	}

	statusOutput, statusErr := runLaunchctl("print", target)
	if statusErr != nil && isAbsentLaunchctlService(statusOutput) {
		return nil
	}
	return fmt.Errorf("unload LaunchAgent: %s", launchctlError(err, output))
}

func launchctlTarget() string {
	return launchctlDomain() + "/" + launchAgentLabel
}

func isAbsentLaunchctlService(output []byte) bool {
	text := strings.ToLower(strings.Join(strings.Fields(string(output)), " "))
	return strings.Contains(text, "could not find service") || strings.Contains(text, "service not found")
}

func launchctlError(err error, output []byte) string {
	details := strings.TrimSpace(string(output))
	if details == "" {
		return err.Error()
	}
	return fmt.Sprintf("%v: %s", err, details)
}

func parseScheduleTime(value string) (scheduleTime, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 || (len(parts[0]) != 1 && len(parts[0]) != 2) || len(parts[1]) != 2 {
		return scheduleTime{}, fmt.Errorf("invalid schedule time %q: expected H:MM or HH:MM", value)
	}
	hour, hourErr := strconv.Atoi(parts[0])
	minute, minuteErr := strconv.Atoi(parts[1])
	if hourErr != nil || minuteErr != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return scheduleTime{}, fmt.Errorf("invalid schedule time %q: expected H:MM or HH:MM", value)
	}
	return scheduleTime{hour: hour, minute: minute}, nil
}

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist"), nil
}

func launchLogDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Logs", "personal-attendance-record"), nil
}

func launchctlDomain() string {
	currentUser, err := user.Current()
	if err != nil {
		return "gui/0"
	}
	return "gui/" + currentUser.Uid
}

var runLaunchctl = func(args ...string) ([]byte, error) {
	return exec.Command("/bin/launchctl", args...).CombinedOutput()
}

func writeLaunchAgent(filename, executable string, times []scheduleTime, logDir string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create LaunchAgent plist: %w", err)
	}
	defer file.Close()

	if err := encodeLaunchAgent(file, executable, times, logDir); err != nil {
		return fmt.Errorf("write LaunchAgent plist: %w", err)
	}
	return nil
}

func encodeLaunchAgent(writer io.Writer, executable string, times []scheduleTime, logDir string) error {
	encoder := xml.NewEncoder(writer)
	if err := encoder.EncodeToken(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)}); err != nil {
		return err
	}
	if err := encoder.EncodeToken(xml.CharData("\n")); err != nil {
		return err
	}
	if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "plist"}, Attr: []xml.Attr{{Name: xml.Name{Local: "version"}, Value: "1.0"}}}); err != nil {
		return err
	}
	if err := writeLaunchAgentDict(encoder, executable, times, logDir); err != nil {
		return err
	}
	if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "plist"}}); err != nil {
		return err
	}
	return encoder.Flush()
}

func writeLaunchAgentDict(encoder *xml.Encoder, executable string, times []scheduleTime, logDir string) error {
	if err := startElement(encoder, "dict"); err != nil {
		return err
	}
	if err := writeKeyValue(encoder, "Label", func() error { return writeString(encoder, launchAgentLabel) }); err != nil {
		return err
	}
	if err := writeKeyValue(encoder, "ProgramArguments", func() error {
		if err := startElement(encoder, "array"); err != nil {
			return err
		}
		if err := writeString(encoder, executable); err != nil {
			return err
		}
		if err := writeString(encoder, "--check"); err != nil {
			return err
		}
		return endElement(encoder, "array")
	}); err != nil {
		return err
	}
	if err := writeKeyValue(encoder, "StartCalendarInterval", func() error {
		if err := startElement(encoder, "array"); err != nil {
			return err
		}
		for _, schedule := range times {
			if err := startElement(encoder, "dict"); err != nil {
				return err
			}
			if err := writeKeyValue(encoder, "Hour", func() error { return writeInteger(encoder, schedule.hour) }); err != nil {
				return err
			}
			if err := writeKeyValue(encoder, "Minute", func() error { return writeInteger(encoder, schedule.minute) }); err != nil {
				return err
			}
			if err := endElement(encoder, "dict"); err != nil {
				return err
			}
		}
		return endElement(encoder, "array")
	}); err != nil {
		return err
	}
	if err := writeKeyValue(encoder, "StandardOutPath", func() error {
		return writeString(encoder, filepath.Join(logDir, "stdout.log"))
	}); err != nil {
		return err
	}
	if err := writeKeyValue(encoder, "StandardErrorPath", func() error {
		return writeString(encoder, filepath.Join(logDir, "stderr.log"))
	}); err != nil {
		return err
	}
	return endElement(encoder, "dict")
}

func writeKeyValue(encoder *xml.Encoder, key string, writeValue func() error) error {
	if err := writeKey(encoder, key); err != nil {
		return err
	}
	return writeValue()
}

func writeKey(encoder *xml.Encoder, key string) error {
	return encoder.EncodeElement(key, xml.StartElement{Name: xml.Name{Local: "key"}})
}

func writeString(encoder *xml.Encoder, value string) error {
	return encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: "string"}})
}

func writeInteger(encoder *xml.Encoder, value int) error {
	return encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: "integer"}})
}

func startElement(encoder *xml.Encoder, name string) error {
	return encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: name}})
}

func endElement(encoder *xml.Encoder, name string) error {
	return encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: name}})
}
