package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/verify"
)

func main() {
	flags := flag.NewFlagSet("verify", flag.ExitOnError)
	contractPath := flags.String("contract", ".gooo/staged-quasiquote.gooo", "authoritative .gooo contract")
	conformanceDir := flags.String("conformance-dir", "", "caller-owned conformance output directory")
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "verify" {
		args = args[1:]
	}
	_ = flags.Parse(args)
	if *conformanceDir == "" {
		fail("verify requires --conformance-dir")
	}
	c, _, digest, err := contract.Load(*contractPath)
	if err != nil {
		fail(err.Error())
	}
	if err := verify.Verify(c, digest, *conformanceDir); err != nil {
		fail(err.Error())
	}
	fmt.Println("verifier: CLOSED")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
