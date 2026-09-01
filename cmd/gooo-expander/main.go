package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/expander"
)

func main() {
	if len(os.Args) < 2 {
		fail("command is required: run or conformance")
	}
	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	case "conformance":
		conformance(os.Args[2:])
	default:
		fail(fmt.Sprintf("unknown command %q", os.Args[1]))
	}
}

func run(args []string) {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	contractPath := flags.String("contract", ".gooo/staged-quasiquote.gooo", "authoritative .gooo contract")
	scenarioID := flags.String("scenario", "", "declared scenario to expand")
	outputDir := flags.String("output-dir", "", "caller-owned output directory")
	_ = flags.Parse(args)
	if *scenarioID == "" || *outputDir == "" {
		fail("run requires --scenario and --output-dir")
	}
	c, _, digest, err := contract.Load(*contractPath)
	if err != nil {
		fail(err.Error())
	}
	report, _, err := expander.RunScenario(c, digest, *scenarioID, *outputDir, nil)
	if err != nil {
		fail(err.Error())
	}
	fmt.Printf("expander: scenario=%s decision=%s generated=%s\n", report.Scenario, report.ObservedStatus, report.GeneratedGoPath)
}

func conformance(args []string) {
	flags := flag.NewFlagSet("conformance", flag.ExitOnError)
	contractPath := flags.String("contract", ".gooo/staged-quasiquote.gooo", "authoritative .gooo contract")
	outputDir := flags.String("output-dir", "", "caller-owned output directory")
	_ = flags.Parse(args)
	if *outputDir == "" {
		fail("conformance requires --output-dir")
	}
	c, _, digest, err := contract.Load(*contractPath)
	if err != nil {
		fail(err.Error())
	}
	report, err := expander.RunConformance(c, digest, *outputDir)
	if err != nil {
		fail(err.Error())
	}
	fmt.Printf("conformance: scenarios=%d contract=%s corpus=%s\n", report.ScenarioCount, report.ContractDecision, report.CorpusResolution)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
