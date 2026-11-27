package record

import (
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

const supportedSchema = 1

const configName = "config.yaml"

var positivePattern = regexp.MustCompile(`^[1-9][0-9]*$`)

var defaultLimits = ConfigLimits{
	Guide: 150,
	Brief: 60,
	Hook:  20,
}

var defaultReview = map[string]string{
	"convention":   "6mo",
	"architecture": "12mo",
	"product":      "12mo",
	"domain":       "12mo",
	"runbook":      "6mo",
	"decision":     "6mo",
}

type ConfigLimits struct {
	Guide int
	Brief int
	Hook  int
}

type Config struct {
	Schema   int
	Limits   ConfigLimits
	Review   map[string]memory.ReviewSpan
	Runners  map[string]Runner
	Packs    ConfigPacks
	Approved kernel.Approval
}

type Runner struct {
	Run      []string
	Baseline []kernel.PathGlob
}

func (s Store) ConfigFile() string {
	return filepath.Join(s.root, configName)
}

func (s Store) LoadConfig() (Config, []error) {
	file := s.ConfigFile()

	data, err := os.ReadFile(file)
	if err != nil {
		return Config{Limits: defaultLimits}, []error{
			kernel.NewRecordFailure(file, 0, err),
		}
	}

	return decodeConfig(file, data)
}

func decodeConfig(file string, data []byte) (Config, []error) {
	root, err := ParseRecord(file, data)
	if err != nil {
		return Config{Limits: defaultLimits}, parseFailure(file, err)
	}

	f := newFields(file, root)
	config := Config{
		Schema:   configSchema(f),
		Limits:   configLimits(f),
		Review:   configReview(f),
		Runners:  configRunners(f),
		Packs:    configPacks(f),
		Approved: approval(f),
	}

	if len(*f.failures) > 0 {
		return Config{Limits: defaultLimits}, *f.failures
	}

	return config, nil
}

func ParseConfig(file string, data []byte) (Config, []error) {
	return decodeConfig(file, data)
}

func configRunners(f *fields) map[string]Runner {
	section, found := f.mapping("runners")
	if !found {
		return map[string]Runner{}
	}

	runners := make(map[string]Runner, len(section.values))
	for name, node := range section.values {
		spec, valid := section.nested(node)
		if !valid {
			continue
		}

		run := many(spec, "run", parseArgument)
		if len(run) == 0 {
			spec.fail(node.Line, failAt(RecordRunnerInvalid, f.file, node.Line))
		}

		runners[name] = Runner{
			Run:      run,
			Baseline: many(spec, "baseline", kernel.ParsePathGlob),
		}
	}

	return runners
}

func parseArgument(raw string) (string, error) {
	if raw == "" {
		return "", recordErrors.Fail(RecordRunnerInvalid, nil)
	}

	return raw, nil
}

func configSchema(f *fields) int {
	node, found := f.values["schema"]
	if !found {
		line := f.root.Line
		f.fail(line, failAt(RecordSchemaMissing, f.file, line))
		return 0
	}

	return parseScalar(f, node, parseSchema)
}

func parseSchema(raw string) (int, error) {
	schema, err := parsePositive(raw)
	if err != nil {
		return 0, err
	}

	if schema > supportedSchema {
		details := map[string]any{
			"schema":    schema,
			"supported": supportedSchema,
		}
		return 0, recordErrors.Fail(RecordSchemaNewer, details)
	}

	return schema, nil
}

func parsePositive(raw string) (int, error) {
	if !positivePattern.MatchString(raw) {
		return 0, recordErrors.Fail(RecordFieldType, nil)
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, recordErrors.Fail(RecordFieldType, nil).WithCause(err)
	}

	return value, nil
}

func configLimits(f *fields) ConfigLimits {
	section, found := f.mapping("limits")
	if !found {
		return defaultLimits
	}

	return ConfigLimits{
		Guide: configLimit(section, "guide", defaultLimits.Guide),
		Brief: configLimit(section, "brief", defaultLimits.Brief),
		Hook:  configLimit(section, "hook", defaultLimits.Hook),
	}
}

func configLimit(section *fields, key string, fallback int) int {
	node, found := section.values[key]
	if !found {
		return fallback
	}

	return parseScalar(section, node, parsePositive)
}

func configReview(f *fields) map[string]memory.ReviewSpan {
	section, hasSection := f.mapping("review")
	review := make(map[string]memory.ReviewSpan, len(defaultReview))

	for _, key := range slices.Sorted(maps.Keys(defaultReview)) {
		span, err := memory.ParseReviewSpan(defaultReview[key])
		if err != nil {
			f.fail(f.root.Line, err)
			continue
		}

		if hasSection {
			if node, found := section.values[key]; found {
				span = parseScalar(section, node, memory.ParseReviewSpan)
			}
		}

		review[key] = span
	}

	return review
}
