package main

import (
	"flag"

	"github.com/bagaswh/lantas/pkg/config"
	"github.com/bagaswh/lantas/pkg/lantas"
	pkgzerolog "github.com/bagaswh/lantas/pkg/zerolog"
	"github.com/rs/zerolog/log"
)

// flags
var (
	configFile = flag.String("config-file", "", "Lantas config file")
	rootDir    = flag.String("root-dir", "", "Lantas will read any relative file path from this dir")

	requiredFlags = []string{
		"config-file",
		"root-dir",
	}
)

func main() {
	setupLogger("ERROR")
	runCmd()
}

func setupLogger(level string) {
	log.Logger = *pkgzerolog.SetupZeroLog(level)
}

func runCmd() {
	flag.Parse()
	mandatoryFlags()

	rt, readConfigErr := config.ReadConfig(*configFile)
	if readConfigErr != nil {
		log.Fatal().Msgf("cannot read config file %s: %w", *configFile, readConfigErr)
	}
	configValidationErr := rt.Validate()
	if configValidationErr != nil {
		log.Fatal().Msgf("failed to validate config: %s", configValidationErr)
	}

	lantasQ := lantas.NewLantas(rt, *rootDir)
	initErr := lantasQ.Init()
	if initErr != nil {
		log.Fatal().Err(initErr).Msg("failed to init Lantas")
	}
	lantasQ.Wait()
}

func mandatoryFlags() {
	seen := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != "" {
			seen[f.Name] = true
		}
	})
	for _, req := range requiredFlags {
		if !seen[req] {
			log.Fatal().Msgf("missing required -%s argument/flag\n", req)
		}
	}
}
