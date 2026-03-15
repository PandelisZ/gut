package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gut/native/common"
	guttesting "gut/testing"
)

type config struct {
	Format               string
	Mutable              bool
	RequiredCapabilities []common.Capability
}

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(stdout io.Writer, args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}

	report := guttesting.Evaluate(guttesting.Options{
		Mutable:              cfg.Mutable,
		RequiredCapabilities: cfg.RequiredCapabilities,
	})

	switch cfg.Format {
	case "text":
		_, err = fmt.Fprintln(stdout, guttesting.FormatReportText(report))
		return err
	case "json":
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report.CapabilityReport())
	default:
		return fmt.Errorf("unsupported format %q (expected text or json)", cfg.Format)
	}
}

func parseArgs(args []string) (config, error) {
	cfg := config{}

	flags := flag.NewFlagSet("gutenv", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	format := flags.String("format", "text", "output format: text or json")
	required := flags.String("require", "", "comma-separated required capabilities")
	mutable := flags.Bool("mutable", false, "include the mutation gate in readiness evaluation")

	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 {
		return cfg, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	cfg.Format = normalizeFormat(*format)
	if cfg.Format != "text" && cfg.Format != "json" {
		return cfg, fmt.Errorf("unsupported format %q (expected text or json)", *format)
	}

	cfg.Mutable = *mutable
	cfg.RequiredCapabilities = parseCapabilityList(*required)
	return cfg, nil
}

func normalizeFormat(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func parseCapabilityList(value string) []common.Capability {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	seen := make(map[common.Capability]struct{})
	capabilities := make([]common.Capability, 0)
	for _, item := range strings.Split(value, ",") {
		capability := common.Capability(strings.TrimSpace(item))
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		capabilities = append(capabilities, capability)
	}

	sort.Slice(capabilities, func(i, j int) bool {
		return capabilities[i] < capabilities[j]
	})
	return capabilities
}
