# slog: systemd journal handler

[![tag](https://img.shields.io/github/tag/tschaefer/slog-journal.svg)](https://github.com/tschaefer/slog-journal/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-%23007d9c)
[![GoDoc](https://godoc.org/github.com/tschaefer/slog-journal?status.svg)](https://pkg.go.dev/github.com/tschaefer/slog-journal)
![Build Status](https://github.com/tschaefer/slog-journal/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/tschaefer/slog-journal)](https://goreportcard.com/report/github.com/tschaefer/slog-journal)
[![Coverage](https://img.shields.io/codecov/c/github/tschaefer/slog-journal)](https://codecov.io/gh/tschaefer/slog-journal)
[![Contributors](https://img.shields.io/github/contributors/tschaefer/slog-journal)](https://github.com/tschaefer/slog-journal/graphs/contributors)
[![License](https://img.shields.io/github/license/tschaefer/slog-journal)](./LICENSE)

A systemd journal Handler for [slog](https://pkg.go.dev/log/slog) Go library.

## 🚀 Install

```sh
go get github.com/tschaefer/slog-journal
```

**Compatibility**: go >= 1.23

No breaking changes will be made to exported APIs before v1.0.0.

## 💡 Usage

GoDoc: [https://pkg.go.dev/github.com/tschaefer/slog-journal](https://pkg.go.dev/github.com/tschaefer/slog-journal)

### Handler options

```go
type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// optional: customize journal event builder
	Converter Converter
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}
```

Attributes will be injected in journal entry fields. Fields must be composed of
uppercase letters, numbers, and underscores, but must not start with an
underscore. Within these restrictions, any arbitrary field name may be used.
Some names have special significance: see the
[journalctl documentation](http://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html)
for more details. The converter will skip invalid attribute key names and
transform to upper case. Additionally the fields will be prefixed with
`SLOG_`.

Other global parameters:

```go
slogjournal.SourceKey = "source"
slogjournal.ErrorKeys = []string{"error", "err"}
slogjournal.FieldPrefix string
slogjournal.LogLevelToPriority = map[string]journal.Priority{
	"DEBUG": journal.PriDebug,
	"INFO":  journal.PriInfo,
	"WARN":  journal.PriWarning,
	"ERROR": journal.PriErr,
}
```

Use `slogjournal.FieldPrefix` to customize the fields prefix.

### Example

For further examples view the tests: [slogjournal_test.go](./slogjournal_test.go)

```go
import (
	"fmt"
	"log"
	"log/slog"
	"time"

	slogjournal "github.com/tschaefer/slog-journal"
)

func main() {
	logger := slog.New(slogjournal.Option{Level: slog.LevelDebug}.NewJournalHandler())
	logger = logger.
		With("environment", "dev").
		With("release", "v1.0.0")

	logger.
		With("category", "sql").
		With("query_statement", "SELECT COUNT(*) FROM users;").
		With("query_duration", 1*time.Second).
		With("error", fmt.Errorf("could not count users")).
		Error("caramba!")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		Info("user registration")
}
```

Output:

```shell
journaltl --output json-pretty --lines 2 --no-pager SLOG_LOGGER=tschaefer/slog-journal:v0.1.0
```

```json
{
  "_PID": "919303",
  "SLOG_QUERY_DURATION": "1s",
  "_BOOT_ID": "100da27bd8b94096b5c80cdac34d6063",
  "_SELINUX_CONTEXT": "unconfined\n",
  "_SYSTEMD_OWNER_UID": "1000",
  "_CAP_EFFECTIVE": "0",
  "_TRANSPORT": "journal",
  "SLOG_ERROR_ERROR": "could not count users",
  "_CMDLINE": "/tmp/go-build3177835552/b001/slog-journal.test -test.testlogfile=/tmp/go-build3177835552/b001/testlog.txt -test.paniconexit0 -test.timeout=10m0s -test.v=true",
  "SLOG_CATEGORY": "sql",
  "SLOG_ENVIRONMENT": "dev",
  "PRIORITY": "3",
  "_SYSTEMD_SLICE": "user-1000.slice",
  "__SEQNUM": "6791157",
  "_AUDIT_LOGINUID": "1000",
  "SLOG_QUERY_STATEMENT": "SELECT COUNT(*) FROM users;",
  "_SOURCE_REALTIME_TIMESTAMP": "1763322710754233",
  "SLOG_ERROR_KIND": "*errors.errorString",
  "_GID": "100",
  "_UID": "1000",
  "_AUDIT_SESSION": "1",
  "_SYSTEMD_UNIT": "session-1.scope",
  "MESSAGE": "caramba!",
  "_SYSTEMD_SESSION": "1",
  "SLOG_ERROR_STACK": "<nil>",
  "__SEQNUM_ID": "b3c7821dbfce47a59b06797aea9028ca",
  "SLOG_RELEASE": "v1.0.0",
  "_SYSTEMD_INVOCATION_ID": "021760b3373342b98aaeabf9d12d8d74",
  "__REALTIME_TIMESTAMP": "1763322710754273",
  "_RUNTIME_SCOPE": "system",
  "__CURSOR": "s=b3c7821dbfce47a59b06797aea9028ca;i=679ff5;b=100da27bd8b94096b5c80cdac34d6063;m=6c1d75d2e6;t=643bb8fcc83e1;x=ec6dbb930fa4a0cd",
  "_HOSTNAME": "bullseye",
  "_SYSTEMD_CGROUP": "/user.slice/user-1000.slice/session-1.scope",
  "_EXE": "/tmp/go-build3177835552/b001/slog-journal.test",
  "_MACHINE_ID": "75b649379b874beea04d95463e59c3a1",
  "__MONOTONIC_TIMESTAMP": "464350728934",
  "_SYSTEMD_USER_SLICE": "-.slice",
  "SLOG_LOGGER": "tschaefer/slog-journal:v0.1.0",
  "_COMM": "slog-journal.te"
}
{
  "_SYSTEMD_CGROUP": "/user.slice/user-1000.slice/session-1.scope",
  "SLOG_ENVIRONMENT": "dev",
  "_SYSTEMD_SESSION": "1",
  "_RUNTIME_SCOPE": "system",
  "_COMM": "slog-journal.te",
  "_SYSTEMD_SLICE": "user-1000.slice",
  "_TRANSPORT": "journal",
  "_SYSTEMD_UNIT": "session-1.scope",
  "_CMDLINE": "/tmp/go-build3177835552/b001/slog-journal.test -test.testlogfile=/tmp/go-build3177835552/b001/testlog.txt -test.paniconexit0 -test.timeout=10m0s -test.v=true",
  "__SEQNUM_ID": "b3c7821dbfce47a59b06797aea9028ca",
  "_SYSTEMD_OWNER_UID": "1000",
  "_UID": "1000",
  "SLOG_USER_ID": "user-123",
  "__MONOTONIC_TIMESTAMP": "464350730219",
  "SLOG_USER_CREATED_AT": "2025-11-16 20:51:50.753821198 +0100 CET",
  "_SELINUX_CONTEXT": "unconfined\n",
  "_BOOT_ID": "100da27bd8b94096b5c80cdac34d6063",
  "SLOG_RELEASE": "v1.0.0",
  "_SOURCE_REALTIME_TIMESTAMP": "1763322710754397",
  "_SYSTEMD_USER_SLICE": "-.slice",
  "_SYSTEMD_INVOCATION_ID": "021760b3373342b98aaeabf9d12d8d74",
  "_GID": "100",
  "_CAP_EFFECTIVE": "0",
  "_HOSTNAME": "bullseye",
  "SLOG_LOGGER": "tschaefer/slog-journal:v0.1.0",
  "_PID": "919303",
  "MESSAGE": "user registration",
  "_AUDIT_LOGINUID": "1000",
  "__REALTIME_TIMESTAMP": "1763322710755560",
  "_AUDIT_SESSION": "1",
  "__CURSOR": "s=b3c7821dbfce47a59b06797aea9028ca;i=679ff6;b=100da27bd8b94096b5c80cdac34d6063;m=6c1d75d7eb;t=643bb8fcc88e8;x=771c136fc02cf1ea",
  "_MACHINE_ID": "75b649379b874beea04d95463e59c3a1",
  "_EXE": "/tmp/go-build3177835552/b001/slog-journal.test",
  "__SEQNUM": "6791158",
  "PRIORITY": "6"
}
```
