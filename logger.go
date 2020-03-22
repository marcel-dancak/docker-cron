package dcron

import (
	"encoding/json"
	"io"
	"os"
)

type logMessage struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
}

func formatLog(w io.Writer, t string, p []byte) {
	msg := logMessage{string(p), t}
	json.NewEncoder(w).Encode(msg)
}

type dualLogger struct {
	Type   string
	Stream io.Writer
	Log    io.Writer
	Size   int
}

func (l *dualLogger) Write(p []byte) (n int, err error) {
	formatLog(l.Log, l.Type, p)
	n, err = l.Stream.Write(p)
	l.Size += n
	return
}

type dockerLogger struct {
	Log    io.Writer
	stdout *dualLogger
	stderr *dualLogger
}

func (l *dockerLogger) StdoutWriter() io.Writer {
	return l.stdout
}

func (l *dockerLogger) StderrWriter() io.Writer {
	return l.stderr
}

func newDockerLogger(out io.Writer) *dockerLogger {
	stdoutLogger := &dualLogger{"stdout", os.Stdout, out, 0}
	stderrLogger := &dualLogger{"stderr", os.Stderr, out, 0}
	return &dockerLogger{Log: out, stdout: stdoutLogger, stderr: stderrLogger}
}
