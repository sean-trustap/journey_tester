// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/trustap/journey_tester/pkg/hashset"
)

// This tool works by discovering the files to be tested in `test` and using
// these to populate a `main.go` file, which is then executed.
func main() {
	exitCode, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

const (
	outputFile                = "target/run_journey_tests/main.go"
	testDefnDir               = "test"
	journeyTesterModule       = "github.com/trustap/journey_tester"
	journeyDefnSignature      = "^func [^(]*() \\*journey_tester.Journey {$"
	dataJourneysDefnSignature = "^var DataJourneys = "
	beforeEachTestSignature   = "^var BeforeEachTest = "
	funcNamePat               = "func %() *journey_tester.Journey {"
)

func run() (int, error) {
	projectRepo, err := readModulePath()
	if err != nil {
		return 0, fmt.Errorf("couldn't read module path: %w", err)
	}

	dataJourneyPkgs, err := findDataJourneyDefnPkgs(testDefnDir)
	if err != nil {
		return 0, fmt.Errorf("couldn't find data journey definitions: %w", err)
	}

	dataJourneyCalls := ""
	for _, pkg := range dataJourneyPkgs.AsSlice() {
		importDefn := newImportDefnFromPkg(projectRepo, pkg)
		dataJourneyCalls += fmt.Sprintf("%s.DataJourneys,\n", importDefn.Alias)
	}

	importPkgs := dataJourneyPkgs

	pkgFuncs, err := findJourneyDefnGroups(testDefnDir)
	if err != nil {
		return 0, fmt.Errorf("couldn't find journey definitions: %w", err)
	}

	journeyCalls := ""
	for pkg, funcNames := range pkgFuncs {
		importPkgs.Set(pkg)

		importDefn := newImportDefnFromPkg(projectRepo, pkg)
		for _, funcName := range funcNames {
			journeyCalls += fmt.Sprintf("%s.%s(),\n", importDefn.Alias, funcName)
		}
	}

	beforeEachTestPkgs, err := findBeforeEachTestPkgs(testDefnDir)
	if err != nil {
		return 0, fmt.Errorf("couldn't find BeforeEachTest definitions: %w", err)
	}

	beforeEachTestCalls := ""
	for _, pkg := range beforeEachTestPkgs.AsSlice() {
		importDefn := newImportDefnFromPkg(projectRepo, pkg)
		importPkgs.Set(pkg)
		beforeEachTestCalls += fmt.Sprintf("%s.BeforeEachTest,\n", importDefn.Alias)
	}

	imports := ""
	for _, pkg := range importPkgs.AsSlice() {
		importDefn := newImportDefnFromPkg(projectRepo, pkg)
		imports += fmt.Sprintf("%s \"%s\"\n", importDefn.Alias, importDefn.Path)
	}

	templateParams := map[string]string{
		"Imports":             imports,
		"DataJourneyCalls":    dataJourneyCalls,
		"JourneyCalls":        journeyCalls,
		"BeforeEachTestCalls": beforeEachTestCalls,
	}
	err = executeTemplateToFile(outputFile, rawMainGoTemplate, templateParams)
	if err != nil {
		return 0, fmt.Errorf("couldn't execute template to file for `main.go`: %w", err)
	}

	gofmt := exec.Command("gofmt", "-w", "-s", outputFile)
	output, err := gofmt.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("couldn't format the output file: %w:\n> %s", err, output)
	}

	rootArgs := []string{"run", outputFile}
	procState, err := runProcess("go", append(rootArgs, os.Args[1:]...))
	if err != nil {
		return 0, fmt.Errorf("couldn't run `go run`: %w", err)
	}
	return procState.ExitCode(), nil
}

func readModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("couldn't read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("couldn't find module declaration in go.mod")
}

func newImportDefnFromPkg(projectRepo, pkgRelativePath string) *ImportDefn {
	return &ImportDefn{
		Path:  path.Join(projectRepo, testDefnDir, pkgRelativePath),
		Alias: pkgPathToAlias(pkgRelativePath),
	}
}

type ImportDefn struct {
	Path  string
	Alias string
}

func pkgPathToAlias(s string) string {
	return strings.ReplaceAll(s, "/", "_")
}

func withCreateFile(
	filePath string,
	f func(io.Writer) error,
) error {
	dir := path.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("couldn't create `%s`: %w", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("couldn't create `main.go`: %w", err)
	}

	err = f(file)
	if err != nil {
		errFileClose := file.Close()
		if errFileClose != nil {
			return fmt.Errorf("couldn't close ('%w') while returning error: %w", errFileClose, err)
		}
		return fmt.Errorf("couldn't run function on file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("couldn't close: %w", err)
	}
	return nil
}

// `runProcess` emulates a Linux `exec` by passing the current process's streams
// to the child process and waiting for it to exit, thereby functioning as if
// the child was the current process (even though it is still a child process of
// the current process).
func runProcess(prog string, args []string) (*os.ProcessState, error) {
	progPath, err := exec.LookPath(prog)
	if err != nil {
		return nil, fmt.Errorf("couldn't find executable: %w", err)
	}

	inheritedStdio := []*os.File{os.Stdin, os.Stdout, os.Stderr}
	procAttr := &os.ProcAttr{Files: inheritedStdio}
	// The arguments passed to `StartProcess` must include the command as
	// the first element.
	allArgs := append([]string{progPath}, args...)
	proc, err := os.StartProcess(progPath, allArgs, procAttr)
	if err != nil {
		return nil, fmt.Errorf("couldn't start process: %w", err)
	}

	procState, err := proc.Wait()
	if err != nil {
		return nil, fmt.Errorf("couldn't start process: %w", err)
	}
	return procState, nil
}

func findDataJourneyDefnPkgs(testDefnDir string) (hashset.Set[string], error) {
	matches, err := grep(testDefnDir, dataJourneysDefnSignature)
	if err != nil {
		return nil, fmt.Errorf("couldn't grep: %w", err)
	}

	pkgs := hashset.NewSet[string]()
	for _, match := range matches {
		f := match.FilePath
		if strings.HasSuffix(f, "data.go") {
			pkgs.Set(path.Dir(f))
		}
	}

	return pkgs, nil
}

func findBeforeEachTestPkgs(testDefnDir string) (hashset.Set[string], error) {
	cmd := exec.Command("grep", "--recursive", beforeEachTestSignature, ".")
	cmd.Dir = testDefnDir
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return hashset.NewSet[string](), nil
		}
		return nil, fmt.Errorf("couldn't grep for BeforeEachTest definitions: %w", err)
	}

	pkgs := hashset.NewSet[string]()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[:len(lines)-1] {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		pkgs.Set(path.Dir(parts[0]))
	}
	return pkgs, nil
}

func grep(dir, pattern string) ([]*grepMatch, error) {
	cmd := exec.Command("grep", "--recursive", pattern, ".")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("couldn't find journeys: %w", err)
	}

	matches := []*grepMatch{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[:len(lines)-1] {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			// TODO This function should ideally be updated to
			// properly handle colons in either the file path or the
			// line, or both.
			msg := "dev error: `grep` line should only contain 1 `:` (line is '%s')"
			return nil, fmt.Errorf(msg, line)
		}

		match := &grepMatch{
			FilePath: parts[0],
			Line:     parts[1],
		}
		matches = append(matches, match)
	}

	return matches, nil
}

type grepMatch struct {
	FilePath string
	Line     string
}

// `findJourneyDefnGroups` returns a `map` from packages (relative to
// `testDefnDir`) to the names of journey test functions defined in those
// packages.
func findJourneyDefnGroups(testDefnDir string) (map[string][]string, error) {
	defns, err := findJourneyDefns(testDefnDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't find journey definitions: %w", err)
	}

	pkgFuncs := map[string][]string{}
	for _, defn := range defns {
		skip := !strings.HasSuffix(defn.FilePath, ".go") ||
			strings.HasSuffix(defn.FilePath, "data.go")
		if skip {
			continue
		}
		pkg := path.Dir(defn.FilePath)

		funcNames, ok := pkgFuncs[pkg]
		if !ok {
			// This is technically redundant because `append` works
			// with `nil`, but we assign a default value for
			// clarity.
			funcNames = []string{}
		}
		pkgFuncs[pkg] = append(funcNames, defn.FuncName)
	}

	return pkgFuncs, nil
}

func findJourneyDefns(testDefnDir string) ([]*JourneyDefinition, error) {
	matches, err := grep(testDefnDir, journeyDefnSignature)
	if err != nil {
		return nil, fmt.Errorf("couldn't grep: %w", err)
	}

	journeyDefns := []*JourneyDefinition{}
	for _, match := range matches {
		fn, err := extract(funcNamePat, match.Line)
		if err != nil {
			msg := "dev error: couldn't match function name pattern: '%s'"
			return nil, fmt.Errorf(msg, match.Line)
		}

		jd := &JourneyDefinition{
			FilePath: match.FilePath,
			FuncName: fn,
		}
		journeyDefns = append(journeyDefns, jd)
	}

	return journeyDefns, nil
}

type JourneyDefinition struct {
	FilePath string
	FuncName string
}

// `extract` takes a pattern `pat` containing `%` and a string `s` and returns
// the section of `s` corresponding to `%`.
//
// NOTE `%` is used as the match specifier for simplicity. This value could be
// taken as a parameter, but a more sophisticated mechanism such as Regex should
// likely be used in this place.
func extract(pat, s string) (string, error) {
	parts := strings.Split(pat, "%")
	if len(parts) != 2 {
		return "", fmt.Errorf("`pat` pattern must contain exactly one `%%`")
	}
	withoutPrefix := strings.TrimPrefix(s, parts[0])
	if withoutPrefix == s {
		return "", fmt.Errorf("pattern prefix didn't match")
	}
	extracted := strings.TrimSuffix(withoutPrefix, parts[1])
	if extracted == withoutPrefix {
		return "", fmt.Errorf("pattern suffix didn't match")
	}
	return extracted, nil
}

func executeTemplateToFile(outputFile, rawTemplate string, data any) error {
	t, err := template.New("").Parse(rawTemplate)
	if err != nil {
		return fmt.Errorf("couldn't parse template: %w", err)
	}

	err = withCreateFile(
		outputFile,
		func(w io.Writer) error {
			err := t.Execute(w, data)
			if err != nil {
				return fmt.Errorf("couldn't execute template: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't write template: %w", err)
	}

	return nil
}

// TODO Discover and populate data journeys automatically.
const rawMainGoTemplate = `
package main

import (
	"fmt"
	"os"

	"` + journeyTesterModule + `/pkg/journey_tester"
	{{.Imports}}
)

func main() {
	result, err := journey_tester.Main(journeys)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(2)
	} else if result == journey_tester.TestResultFail {
		os.Exit(1)
	}
}

var journeys = &journey_tester.Journeys{
	DataJourneys: concatDataJourneys([]*journey_tester.DataJourneys{
		{{.DataJourneyCalls}}
	}),

	TestJourneys: []*journey_tester.Journey{
		{{.JourneyCalls}}
	},

	BeforeEachTest: concatBeforeEachTestSteps([][]*journey_tester.Step{
		{{.BeforeEachTestCalls}}
	}),
}

func concatDataJourneys(djs []*journey_tester.DataJourneys) *journey_tester.DataJourneys {
	create := []*journey_tester.Journey{}
	verify := []*journey_tester.Journey{}
	load := []*journey_tester.Journey{}
	del := []*journey_tester.Journey{}

	for _, dj := range djs {
		create = append(create, dj.Create...)
		verify = append(verify, dj.Verify...)
		load = append(load, dj.Load...)
		del = append(del, dj.Delete...)
	}

	return &journey_tester.DataJourneys{
		Create: create,
		Verify: verify,
		Load:   load,
		Delete: del,
	}
}

func concatBeforeEachTestSteps(stepSlices [][]*journey_tester.Step) []*journey_tester.Step {
	result := []*journey_tester.Step{}
	for _, steps := range stepSlices {
		result = append(result, steps...)
	}
	return result
}
`
