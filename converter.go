package slogjournal

import (
	"fmt"
	"log/slog"
	"maps"
	"strconv"

	"github.com/coreos/go-systemd/v22/journal"
	slogcommon "github.com/samber/slog-common"
)

var SourceKey = "source"
var ErrorKeys = []string{"error", "err"}

var FieldPrefix string
var LogLevelToPriority = map[string]journal.Priority{
	"DEBUG": journal.PriDebug,
	"INFO":  journal.PriInfo,
	"WARN":  journal.PriWarning,
	"ERROR": journal.PriErr,
}

var appliedFieldPrefix string

type Converter func(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) (string, journal.Priority, map[string]string)

func DefaultConverter(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) (string, journal.Priority, map[string]string) {
	// aggregate all attributes
	attrs := slogcommon.AppendRecordAttrsToAttrs(loggerAttr, groups, record)

	// developer formatters
	attrs = slogcommon.ReplaceError(attrs, ErrorKeys...)
	if addSource {
		attrs = append(attrs, slogcommon.Source(SourceKey, record))
	}
	attrs = slogcommon.ReplaceAttrs(replaceAttr, []string{}, attrs...)
	attrs = slogcommon.RemoveEmptyAttrs(attrs)

	message := record.Message

	prio, ok := LogLevelToPriority[record.Level.String()]
	if !ok {
		prio = journal.PriDebug
	}

	fields := map[string]string{
		"SLOG_LOGGER": name + ":" + version,
	}
	maps.Copy(fields, attrsToFields(fieldPrefixer(), attrs))

	return message, prio, fields
}

func attrsToFields(prefix string, attrs []slog.Attr) map[string]string {
	nestedMap := slogcommon.AttrsToMap(attrs...)
	flatMap := flattenToMap(nestedMap)

	fields := map[string]string{}
	for k, v := range flatMap {
		skip := false
		var field string
		for _, c := range k {
			if ('A' > c || c > 'Z') && ('a' > c || c > 'z') && ('0' > c || c > '9') && c != '_' {
				skip = true
				break
			}

			if 'a' <= c && c <= 'z' {
				c -= 'a' - 'A'
			}
			field += string(c)
		}

		if skip {
			continue
		}

		s := fmt.Sprintf("%+v", v)
		fields[prefix+field] = s
	}

	return fields
}

func fieldPrefixer() string {
	if appliedFieldPrefix != "" && appliedFieldPrefix == FieldPrefix {
		return appliedFieldPrefix
	}

	appliedFieldPrefix = func() string {
		if FieldPrefix == "" {
			return "SLOG_"
		}

		if 'A' > FieldPrefix[0] || FieldPrefix[0] > 'Z' {
			return "SLOG_"
		}

		for _, c := range FieldPrefix {
			if ('A' > c || c > 'Z') && ('0' > c || c > '9') && c != '_' {
				return "SLOG_"
			}
		}

		return FieldPrefix + "_"
	}()
	FieldPrefix = appliedFieldPrefix

	return appliedFieldPrefix
}

func flattenToMap(nested map[string]any) map[string]any {
	flat := make(map[string]any)
	if nested == nil {
		return flat
	}

	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, val := range t {
				key := k
				if prefix != "" {
					key = prefix + "_" + k
				}
				walk(key, val)
			}
		case []any:
			for i, elem := range t {
				idx := strconv.Itoa(i)
				key := idx
				if prefix != "" {
					key = prefix + "_" + idx
				}
				walk(key, elem)
			}
		default:
			flat[prefix] = t
		}
	}

	for k, v := range nested {
		walk(k, v)
	}

	return flat
}
