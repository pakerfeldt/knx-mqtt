package utils

import (
	golog "log"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vapourismo/knx-go/knx/dpt"
	"github.com/vapourismo/knx-go/knx/util"
)

func StringWithoutSuffix(dpt dpt.Datapoint) string {
	return strings.Trim(strings.TrimSuffix(dpt.String(), dpt.Unit()), " ")
}

func SetupLogging(logLevel string, enableKnxLogs bool) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Error().Msgf("%s", err)
		level = zerolog.NoLevel
	}

	log.Logger = log.Output(zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.FieldsOrder = []string{"protocol"}
		},
	)).Level(level)

	if enableKnxLogs {
		// Enables stdout logging in knx-go
		util.Logger = golog.New(os.Stdout, "", golog.LstdFlags)
	}
}
