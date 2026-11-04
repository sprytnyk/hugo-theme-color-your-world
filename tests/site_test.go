package tests

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	// The theme repo, one folder up from here
	themeDir string

	// Where the sites get built. TestMain deletes it unless a test fails
	workDir string
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	version, err := exec.Command("hugo", "version").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't run hugo:", err)
		return 1
	}
	if !bytes.Contains(version, []byte("+extended")) {
		fmt.Fprintf(os.Stderr, "Hugo has to be the Extended version: %s", version)
		return 1
	}
	fmt.Printf("%s\n", version)

	if themeDir, err = filepath.Abs(".."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if workDir, err = os.MkdirTemp("", "cyw-tests-"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	code := m.Run()
	if code != 0 {
		fmt.Println("The sites are in", workDir)
	} else {
		os.RemoveAll(workDir)
	}
	return code
}

// giscus returns the giscus settings under the given TOML table
func giscus(table string) string {
	return fmt.Sprintf(`
[%s]
  repo       = "owner/repo"
  repoId     = "R_test"
  category   = "Announcements"
  categoryId = "DIC_test"
`, table)
}

// The sites the tests check. The first test that asks for a site builds it,
// and later tests reuse that build.
var (
	// giscus on for all languages
	mainSite = fixture("main", giscus("params.giscus"))

	// giscus on for Ukrainian only
	oneLanguageSite = fixture("one-language", giscus("languages.uk.params.giscus"))

	// mainSite with baseURL = "/"
	rootSite = fixture("root", giscus("params.giscus"), "--baseURL", "/")
)

// fixture returns a function that hands a built site to a test. It builds the
// site on the first call. config goes at the end of hugo.toml, and flags go to
// Hugo.
func fixture(name, config string, flags ...string) func(*testing.T) *site {
	once := sync.OnceValues(func() (string, error) {
		return build(name, config, flags)
	})

	return func(t *testing.T) *site {
		t.Helper()
		public, err := once()
		if err != nil {
			t.Fatal(err)
		}
		return &site{t: t, public: public}
	}
}

// build copies the test site to workDir/name, links the theme in and builds it.
// It returns the folder with the output.
func build(name, config string, flags []string) (string, error) {
	dir := filepath.Join(workDir, name)

	if err := os.CopyFS(dir, os.DirFS("testdata/site")); err != nil {
		return "", err
	}
	if err := os.Mkdir(filepath.Join(dir, "themes"), 0o755); err != nil {
		return "", err
	}
	if err := os.Symlink(themeDir, filepath.Join(dir, "themes", "color-your-world")); err != nil {
		return "", err
	}

	cover, err := os.ReadFile(filepath.Join(themeDir, "images", "screenshot.png"))
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "content", "posts", "sample", "cover.png"), cover, 0o644); err != nil {
		return "", err
	}

	f, err := os.OpenFile(filepath.Join(dir, "hugo.toml"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(config)
	if err := f.Close(); err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}

	public := filepath.Join(dir, "public")
	args := append([]string{"--source", dir, "--destination", public, "--gc", "--minify"}, flags...)
	out, err := exec.Command("hugo", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("the %s build failed: %v\n%s", name, err, out)
	}

	if warnings := regexp.MustCompile(`(?m)^(WARN|ERROR).*`).FindAll(out, -1); warnings != nil {
		return "", fmt.Errorf("the %s build printed warnings:\n%s", name, bytes.Join(warnings, []byte("\n")))
	}

	return public, nil
}

// site is a built test site, as seen by one test
type site struct {
	t      *testing.T
	public string
}

// read returns a file from the output
func (s *site) read(path string) string {
	s.t.Helper()
	data, err := os.ReadFile(filepath.Join(s.public, path))
	if err != nil {
		s.t.Fatal(err)
	}
	return string(data)
}

// has checks that the file at path matches pattern, a regexp written as for
// unminified HTML. See htmlPattern.
func (s *site) has(path, pattern string) {
	s.t.Helper()
	if !htmlPattern(pattern).MatchString(s.read(path)) {
		s.t.Errorf("%s doesn't have %s", path, pattern)
	}
}

// lacks checks that the file at path doesn't match pattern
func (s *site) lacks(path, pattern string) {
	s.t.Helper()
	if htmlPattern(pattern).MatchString(s.read(path)) {
		s.t.Errorf("%s has %s", path, pattern)
	}
}

// exists checks that the file at path is there and isn't empty
func (s *site) exists(path string) {
	s.t.Helper()
	info, err := os.Stat(filepath.Join(s.public, path))
	if err != nil {
		s.t.Error(err)
	} else if info.Size() == 0 {
		s.t.Errorf("%s is empty", path)
	}
}

// glob returns the one file in the output that matches pattern
func (s *site) glob(pattern string) string {
	s.t.Helper()
	matches, err := filepath.Glob(filepath.Join(s.public, pattern))
	if err != nil || len(matches) != 1 {
		s.t.Fatalf("want one file for %s, got %v", pattern, matches)
	}
	path, err := filepath.Rel(s.public, matches[0])
	if err != nil {
		s.t.Fatal(err)
	}
	return path
}

// htmlPattern compiles a regexp for minified HTML. --minify drops the quotes
// around some attribute values, so each " outside of [...] may be missing.
func htmlPattern(pattern string) *regexp.Regexp {
	var b strings.Builder
	inClass := false

	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '\\' && i+1 < len(pattern):
			b.WriteByte(c)
			i++
			c = pattern[i]
		case c == '[':
			inClass = true
		case c == ']':
			inClass = false
		case c == '"' && !inClass:
			b.WriteString(`"?`)
			continue
		}
		b.WriteByte(c)
	}

	return regexp.MustCompile(b.String())
}

// findBrowser returns the path to Chromium, or "" if it isn't installed
func findBrowser() string {
	for _, name := range []string{"chromium", "chromium-browser", "google-chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// giscusTheme opens url in headless Chromium, with no network beyond this
// machine, and returns the stylesheet main.js hands to giscus. giscus gets it
// as a base64 data: URL.
func giscusTheme(t *testing.T, browser, url string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dom, err := exec.CommandContext(ctx, browser,
		"--headless", "--disable-gpu",
		"--user-data-dir="+t.TempDir(),
		"--host-resolver-rules=MAP * ~NOTFOUND, EXCLUDE 127.0.0.1",
		"--virtual-time-budget=5000",
		"--dump-dom", url,
	).Output()
	if err != nil {
		t.Logf("Chromium: %v", err)
	}

	m := regexp.MustCompile(`data-theme="data:text/css;base64,([^"]*)"`).FindSubmatch(dom)
	if m == nil {
		t.Fatal("giscus didn't get a data: URL theme")
	}

	css, err := base64.StdEncoding.DecodeString(string(m[1]))
	if err != nil {
		t.Fatal(err)
	}
	return string(css)
}
