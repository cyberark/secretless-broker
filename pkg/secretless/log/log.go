package log

// Logger interface is used to emit logging information about the
// broker to common output(s).
type Logger interface {
	CopyWith(prefix string, isDebug bool) Logger
	DebugEnabled() bool
	Prefix() string

	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})

	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}
