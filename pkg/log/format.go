package log

//ClarkezoneWriter implements the interface of the writers
type ClarkezoneWriter interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Success(format string, args ...interface{})
}

const (
	//TTYFormat represents a tty logger
	TTYFormat string = "tty"
)

func (l *logger) getWriter(format string) ClarkezoneWriter {
	switch format {
	case TTYFormat:
		l.outputMode = TTYFormat
		return newTTYWriter(l.out, l.file)
	default:
		Debugf("could not load %s. Callback to 'tty'", format)
		l.outputMode = TTYFormat
		return newTTYWriter(l.out, l.file)
	}
}
