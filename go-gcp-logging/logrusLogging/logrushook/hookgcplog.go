package logrushook

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"
)

// 流用元：https://github.com/sirupsen/logrus/blob/master/hooks/writer/writer.go

// HookGCPLog is a hook that writes logs of specified LogLevels to specified Writer
type HookGCPLog struct {
	Writer      io.Writer
	LogLevels   []log.Level
	ErrorReport bool
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *HookGCPLog) Fire(entry *log.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		return err
	}
	if hook.ErrorReport == true {
		line, _ = insertErrorReportMark(line)
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *HookGCPLog) Levels() []log.Level {
	return hook.LogLevels
}

// insertErrorReportMark Cloud Error Reporting用の識別子を付与する
func insertErrorReportMark(line []byte) ([]byte, error) {
	var jsonData interface{}
	err := json.Unmarshal([]byte(line), &jsonData)
	if err != nil {
		return line, err
	}
	jsonKeyValue, ok := jsonData.(map[string]interface{})
	if ok == true {
		jsonKeyValue["@type"] = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
		jsonKeyValue["severity"] = "fatal" // Error Reportingへ挿入するseverityはfatal固定にする
	}
	blob, err := json.Marshal(jsonData)
	if err != nil {
		return line, err
	}
	blob = append(blob, '\n') // jsonUnmarshalで改行が取れてしまうため再度付与
	return blob, nil
}
