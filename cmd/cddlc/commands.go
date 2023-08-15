package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/parser"
	"github.com/flowfunction/cddl/token"
	"github.com/urfave/cli/v2"
)

const (
	TAB = "    "
)

func readSource(filename string) (src []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	src, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	return
}

func checkArgs(cCtx *cli.Context, n int) bool {
	return cCtx.Args().Len() == n
}

func LexerCmd(cCtx *cli.Context) error {
	if !checkArgs(cCtx, 1) {
		return errors.New("filename required")
	}

	src, err := readSource(cCtx.Args().First())
	if err != nil {
		return err
	}
	lex := lexer.NewLexer(src)

	illegals := 0
	for {
		tok, pos, lit := lex.Scan()
		fmt.Printf("%s: %s -> %s\n", pos, tok, lit)
		if tok == token.EOF {
			break
		}
		if tok == token.ILLEGAL {
			illegals += 1
		}
	}

	if illegals > 0 {
		return fmt.Errorf("found %d illegal characters", illegals)
	}
	return nil
}

func ParseCmd(cCtx *cli.Context) error {
	if !checkArgs(cCtx, 1) {
		return errors.New("filename required")
	}

	src, err := readSource(cCtx.Args().First())
	if err != nil {
		return err
	}
	lex := lexer.NewLexer(src)
	prs := parser.NewParser(lex)

	ast, errs := prs.ParseFile()

	if len(errs) > 0 {
		outs := errorStringer(src, errs)
		fmt.Fprintln(os.Stderr)
		for _, out := range outs {
			fmt.Fprintln(os.Stderr, out)
		}
	}

	b, err := json.MarshalIndent(ast, "", "	")
	if err != nil {
		fmt.Printf("%+v", ast)
		return err
	}
	fmt.Println(string(b))
	return nil
}

func DoctorCmd(cCtx *cli.Context) error {
	out := ""

	rt := runtime.GOOS
	arch := runtime.GOARCH

	out += fmt.Sprintf("\nVersion\t\t: %s\nCommit\t\t: %s\nBuild Date\t: %s\nRuntime\t\t: %s\nArch\t\t: %s\n", version, commit, date, rt, arch)
	fmt.Println(out)

	return nil
}

func errorStringer(src []byte, errs parser.ErrorList) []string {
	lines := bytes.Split(src, []byte{'\n'})
	lCount := len(lines)

	outs := []string{}

	for _, err := range errs {
		pos := err.Start()
		if pos.Line <= lCount {
			line := string(lines[pos.Line-1])
			lPrefix := fmt.Sprintf("%s%d | ", TAB, pos.Line)
			outs = append(outs,
				fmt.Sprintf("%s\n%s%s\n%*s", err, lPrefix, line, pos.Column+len(lPrefix), "˜"),
			)
		}
	}
	return outs
}
