package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"testing"
)

func TestParseScheduleTime(t *testing.T) {
	tests := []struct {
		value string
		want  scheduleTime
		isErr bool
	}{
		{value: "09:30", want: scheduleTime{hour: 9, minute: 30}},
		{value: "8:30", want: scheduleTime{hour: 8, minute: 30}},
		{value: "23:59", want: scheduleTime{hour: 23, minute: 59}},
		{value: "24:00", isErr: true},
		{value: "12:60", isErr: true},
		{value: "08:3", isErr: true},
	}

	for _, test := range tests {
		got, err := parseScheduleTime(test.value)
		if test.isErr {
			if err == nil {
				t.Errorf("parseScheduleTime(%q) succeeded, want error", test.value)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseScheduleTime(%q) returned error: %v", test.value, err)
			continue
		}
		if got != test.want {
			t.Errorf("parseScheduleTime(%q) = %#v, want %#v", test.value, got, test.want)
		}
	}
}

func TestScheduleTimesFlagRejectsDuplicates(t *testing.T) {
	var times scheduleTimesFlag
	if err := times.Set("09:30"); err != nil {
		t.Fatalf("first Set returned error: %v", err)
	}
	if err := times.Set("09:30"); err == nil {
		t.Fatal("second Set succeeded, want duplicate error")
	}
}

func TestEncodeLaunchAgent(t *testing.T) {
	var output bytes.Buffer
	times := []scheduleTime{{hour: 9, minute: 30}, {hour: 15, minute: 0}}
	if err := encodeLaunchAgent(&output, "/Users/example/bin/par", times, "/Users/example/Library/Logs/par"); err != nil {
		t.Fatalf("encodeLaunchAgent returned error: %v", err)
	}

	var document interface{}
	if err := xml.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("generated plist is not valid XML: %v", err)
	}
	contents := output.String()
	for _, expected := range []string{
		"<key>Label</key><string>com.personal-attendance-record</string>",
		"<key>ProgramArguments</key><array><string>/Users/example/bin/par</string><string>--check</string></array>",
		"<key>Hour</key><integer>9</integer>",
		"<key>Minute</key><integer>30</integer>",
		"<key>Hour</key><integer>15</integer>",
		"<key>StandardOutPath</key><string>/Users/example/Library/Logs/par/stdout.log</string>",
		"<key>StandardErrorPath</key><string>/Users/example/Library/Logs/par/stderr.log</string>",
	} {
		if !strings.Contains(contents, expected) {
			t.Errorf("generated plist does not contain %q:\n%s", expected, contents)
		}
	}
}

func TestUnloadScheduleUsesQualifiedTarget(t *testing.T) {
	originalRunLaunchctl := runLaunchctl
	t.Cleanup(func() { runLaunchctl = originalRunLaunchctl })

	var gotArgs []string
	runLaunchctl = func(args ...string) ([]byte, error) {
		gotArgs = append([]string(nil), args...)
		return nil, nil
	}

	if err := unloadSchedule(); err != nil {
		t.Fatalf("unloadSchedule returned error: %v", err)
	}
	wantArgs := []string{"bootout", launchctlTarget()}
	if strings.Join(gotArgs, " ") != strings.Join(wantArgs, " ") {
		t.Fatalf("launchctl args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestUnloadScheduleAcceptsAbsentService(t *testing.T) {
	originalRunLaunchctl := runLaunchctl
	t.Cleanup(func() { runLaunchctl = originalRunLaunchctl })

	call := 0
	runLaunchctl = func(args ...string) ([]byte, error) {
		call++
		if call == 1 {
			return []byte("Could not find service"), errors.New("exit status 5")
		}
		return []byte("Could not find service"), errors.New("exit status 113")
	}

	if err := unloadSchedule(); err != nil {
		t.Fatalf("unloadSchedule returned error for absent service: %v", err)
	}
}

func TestUnloadScheduleReturnsRealFailure(t *testing.T) {
	originalRunLaunchctl := runLaunchctl
	t.Cleanup(func() { runLaunchctl = originalRunLaunchctl })

	call := 0
	runLaunchctl = func(args ...string) ([]byte, error) {
		call++
		if call == 1 {
			return []byte("permission denied"), errors.New("exit status 5")
		}
		return []byte("service is loaded"), nil
	}

	err := unloadSchedule()
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("unloadSchedule error = %v, want permission failure", err)
	}
}
