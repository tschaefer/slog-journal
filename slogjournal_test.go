package slogjournal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func IsRunningInGitHubActions() bool {
	v, ok := os.LookupEnv("GITHUB_ACTIONS")
	return ok && v == "true"
}

func jsonFormatter(entry *sdjournal.JournalEntry) (string, error) {
	b, err := json.Marshal(entry.Fields)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readJournal(t *testing.T, field, value string) string {
	r, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Matches: []sdjournal.Match{
			{Field: field, Value: value},
		},
		Formatter: jsonFormatter,
	})
	assert.NoError(t, err)
	defer func() {
		_ = r.Close()
	}()

	b := make([]byte, 65536)
	for {
		i, err := r.Read(b)
		if err != nil || i == 0 {
			break
		}
	}

	s := string(b)
	s = strings.TrimRight(s, "\x00")

	return s
}

func Test_JournalHandler(t *testing.T) {
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithCustomFieldPrefix(t *testing.T) {
	FieldPrefix = "CUSTOM"
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "CUSTOM_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["CUSTOM_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithInvalidFieldPrefix(t *testing.T) {
	FieldPrefix = "INVALID-PREFIX"
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)

	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithInvalidFieldPrefixPrefix(t *testing.T) {
	FieldPrefix = "1INVALID_PREFIX"
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)

	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithInvalidField(t *testing.T) {
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid, "%invalid_field%", "value")

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
	assert.NotContains(t, fields, "%invalid_field%")
}

func Test_JournalHandlerAddSource(t *testing.T) {
	o := &Option{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
	assert.Contains(t, fields, "SLOG_SOURCE_FILE")
	assert.Contains(t, fields, "SLOG_SOURCE_LINE")
	assert.Contains(t, fields, "SLOG_SOURCE_FUNCTION")
}

func Test_JournalHandlerWithErrorField(t *testing.T) {
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	errE := errors.New("something went wrong")
	l.Info(name+":"+version, "uuid", uuid, "error", errE)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
	assert.Equal(t, errE.Error(), fields["SLOG_ERROR_ERROR"])
	assert.Equal(t, fmt.Sprintf("%T", errE), fields["SLOG_ERROR_KIND"])
	assert.Equal(t, "<nil>", fields["SLOG_ERROR_STACK"])
}

func Test_JournalHandlerPriorityIsLogLevel(t *testing.T) {
	o := &Option{
		Level: slog.LevelDebug,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	// Info level
	uuidStr := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuidStr)

	e := readJournal(t, "SLOG_UUID", uuidStr)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuidStr, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])

	// Debug level
	uuidStr = uuid.New().String()
	l.Debug(name+":"+version, "uuid", uuidStr)

	e = readJournal(t, "SLOG_UUID", uuidStr)
	assert.NotEmpty(t, e)
	err = json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuidStr, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriDebug), fields["PRIORITY"])
	assert.Equal(t, "7", fields["PRIORITY"])

	// Error level
	uuidStr = uuid.New().String()
	l.Error(name+":"+version, "uuid", uuidStr)

	e = readJournal(t, "SLOG_UUID", uuidStr)
	assert.NotEmpty(t, e)
	err = json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuidStr, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriErr), fields["PRIORITY"])
	assert.Equal(t, "3", fields["PRIORITY"])

	// Warn level
	uuidStr = uuid.New().String()
	l.Warn(name+":"+version, "uuid", uuidStr)

	e = readJournal(t, "SLOG_UUID", uuidStr)
	assert.NotEmpty(t, e)
	err = json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuidStr, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriWarning), fields["PRIORITY"])
}

func Test_JournalHandlerMissingLogLevelToPriority(t *testing.T) {
	LogLevelToPriority = map[string]journal.Priority{}
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriDebug), fields["PRIORITY"])

	LogLevelToPriority = map[string]journal.Priority{
		"DEBUG": journal.PriDebug,
		"INFO":  journal.PriInfo,
		"WARN":  journal.PriWarning,
		"ERROR": journal.PriErr,
	}
}

func Test_JournalHandlerCustomLogLevelToPriority(t *testing.T) {
	LevelTrace := slog.Level(-8)
	LogLevelToPriority = map[string]journal.Priority{
		"DEBUG-4": journal.PriNotice,
	}
	o := &Option{
		Level: LevelTrace,
	}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Log(context.Background(), LevelTrace, name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriNotice), fields["PRIORITY"])

	LogLevelToPriority = map[string]journal.Priority{
		"DEBUG": journal.PriDebug,
		"INFO":  journal.PriInfo,
		"WARN":  journal.PriWarning,
		"ERROR": journal.PriErr,
	}
}

func Test_JournalHandlerNoLogLevel(t *testing.T) {

	o := &Option{}
	h := o.NewJournalHandler()
	assert.IsType(t, &JournalHandler{}, h)

	l := slog.New(h)
	assert.NotNil(t, l)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithGroup(t *testing.T) {
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler().WithGroup("group")
	assert.IsType(t, &JournalHandler{}, h)

	l := slog.New(h)
	assert.NotNil(t, l)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_GROUP_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_GROUP_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}

func Test_JournalHandlerWithAttr(t *testing.T) {
	o := &Option{
		Level: slog.LevelInfo,
	}
	h := o.NewJournalHandler().WithAttrs([]slog.Attr{slog.String("attr", "extra")})
	assert.IsType(t, &JournalHandler{}, h)

	if !journal.Enabled() || IsRunningInGitHubActions() {
		t.Skip("systemd journal is not enabled")
	}

	l := slog.New(h)
	assert.NotNil(t, l)

	uuid := uuid.New().String()
	l.Info(name+":"+version, "uuid", uuid)

	e := readJournal(t, "SLOG_UUID", uuid)
	assert.NotEmpty(t, e)
	var fields map[string]string
	err := json.Unmarshal([]byte(e), &fields)
	assert.NoError(t, err)

	assert.Equal(t, uuid, fields["SLOG_UUID"])
	assert.Equal(t, name+":"+version, fields["MESSAGE"])
	assert.Equal(t, name+":"+version, fields["SLOG_LOGGER"])
	assert.Equal(t, fmt.Sprintf("%d", journal.PriInfo), fields["PRIORITY"])
}
