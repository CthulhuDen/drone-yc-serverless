package main

import (
	log "github.com/sirupsen/logrus"
)

type HasFields interface {
	GetFields() log.Fields
}

type withFields struct {
	error
	f log.Fields
}

func WithFields(err error, f log.Fields) error {
	return &withFields{
		error: err,
		f:     f,
	}
}

func (wf *withFields) GetFields() log.Fields {
	return wf.f
}

func Info(msg string, ctx ...log.Fields) {
	var l log.FieldLogger = log.StandardLogger()
	for _, c := range ctx {
		l = l.WithFields(c)
	}

	l.Info(msg)
}
