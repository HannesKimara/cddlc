package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/HannesKimara/cddlc/config"
	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	gogen "github.com/HannesKimara/cddlc/transforms/codegen/golang"

	"github.com/urfave/cli/v2"
)

const (
	TAB = "    "
)

var (
	errNoConfiguration = errors.New("could not find either cddlc.json or cddlc.yaml in current dir")
)

func readSource(filename string) (src []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	src, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	return
}

func checkArgs(cCtx *cli.Context, n int) bool {
	return cCtx.Args().Len() == n
}

func GenerateCmd(cCtx *cli.Context) error {
	var conf *config.Config

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	conf, err = loadConfigFromDir(wd)
	if errors.Is(err, errNoConfiguration) {
		fmt.Print(err.Error() + " `" + wd + "`")
		fmt.Println(" ...falling back to default config")
		conf = config.NewDefaultConfig()
	} else if err != nil {
		return err
	}

	for _, build := range conf.Builds {
		stat, err := os.Stat(build.SourceDir)
		if err != nil {
			return errors.New("failed to get stat info on source directory `" + build.SourceDir + "` with err: `" + err.Error() + " `")
		}

		if !stat.IsDir() {
			return errors.New("`" + build.SourceDir + "`: is not a directory")
		}
		files, err := os.ReadDir(build.SourceDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if filepath.Ext(file.Name()) != ".cddl" {
				continue
			}
			if _, err := os.Stat(build.OutDir); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(build.OutDir, os.ModePerm)
				if err != nil {
					return errors.New("failed to create output directory `" + build.OutDir + "` with err: `" + err.Error() + " `")
				}
			} else if err != nil {
				return errors.New("failed to get stat info on output directory `" + build.OutDir + "` with err: " + err.Error())
			}
			outPath := filepath.Join(build.OutDir, strings.TrimSuffix(file.Name(), ".cddl")+".go")
			err := generateFile(cCtx, build.Package, filepath.Join(build.SourceDir, file.Name()), outPath)
			if err != nil /*&& skip failed not set*/ {
				return err
			}
		}

	}

	return nil
}

func loadConfigFromDir(dir string) (*config.Config, error) {
	files, err := filepath.Glob(filepath.Join(dir, "cddlc.*"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errNoConfiguration
	}

	return config.LoadConfig(files[0])
}

func generateFile(cCtx *cli.Context, pkgName, filepath, outPath string) error {
	gen := gogen.NewGenerator(pkgName)
	src, err := readSource(filepath)
	if err != nil {
		return err
	}

	// check if output file can be opened/created before generating. Fail otherwise
	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer out.Close()

	lex := lexer.NewLexer(src)
	p := parser.NewParser(lex)
	cddl, errs := p.ParseFile()

	if len(errs) > 0 {
		outs := errorStringer(src, errs)
		fmt.Fprintln(os.Stderr)
		for _, out := range outs {
			fmt.Fprintln(os.Stderr, out)
		}
		return errors.New("parser failed with errors above")
	}

	gen.Visit(cddl)

	err = addBuildHeader(out)
	if err != nil {
		return err
	}

	return nil
}

func addBuildHeader(f io.Writer) error {
	header := "/*\n  File generated using `" + filepath.Base(os.Args[0]) + " " + strings.Join(os.Args[1:], " ") + "`. DO NOT EDIT\n*/\n\n"

	_, err := f.Write([]byte(header))
	if err != nil {
		return err
	}
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
