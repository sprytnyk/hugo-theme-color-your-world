package tests

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

const (
	post  = "posts/sample/index.html"
	quiet = "posts/quiet/index.html"

	// The giscus box, but not other classes that start with giscus
	giscusBox = `class="giscus[" >]`
)

func TestComments(t *testing.T) {
	t.Run("on posts", func(t *testing.T) {
		mainSite(t).has(post, giscusBox)
	})
	t.Run("off with comments: false", func(t *testing.T) {
		mainSite(t).lacks(quiet, giscusBox)
	})
	t.Run("off on standalone pages", func(t *testing.T) {
		mainSite(t).lacks("about/index.html", giscusBox)
	})
	t.Run("giscus themes are built", func(t *testing.T) {
		s := mainSite(t)
		s.exists("css/giscus-dark.css")
		s.exists("css/giscus-light.css")
	})
}

// main.js is built once for all languages, so it has to load giscus even when
// only one language sets it up
func TestCommentsForOneLanguage(t *testing.T) {
	t.Run("none in English", func(t *testing.T) {
		oneLanguageSite(t).lacks(post, giscusBox)
	})
	t.Run("on in Ukrainian", func(t *testing.T) {
		oneLanguageSite(t).has("uk/"+post, giscusBox)
	})
	t.Run("main.js loads giscus", func(t *testing.T) {
		s := oneLanguageSite(t)
		s.has(s.glob("js/main.*.js"), `giscus\.app`)
	})
}

// With baseURL = "/", the theme URLs are paths. giscus loads them from its own
// domain, so main.js has to turn them into full URLs on this site.
func TestRootBaseURL(t *testing.T) {
	t.Run("theme path stays a path in the page", func(t *testing.T) {
		rootSite(t).has(post, `data-theme-dark="/css/giscus-dark\.css`)
	})

	t.Run("in the browser", func(t *testing.T) {
		browser := findBrowser()
		if browser == "" {
			t.Skip("no Chromium found")
		}

		server := httptest.NewServer(http.FileServer(http.Dir(rootSite(t).public)))
		defer server.Close()

		theme := giscusTheme(t, browser, server.URL+"/posts/sample/")

		t.Run("giscus gets a full theme URL", func(t *testing.T) {
			if !strings.Contains(theme, `@import url("`+server.URL+`/css/giscus-`) {
				t.Errorf("no full theme URL in %q", theme)
			}
		})
		t.Run("giscus gets the accent color", func(t *testing.T) {
			if !regexp.MustCompile(`main\{--cyw-accent:#[0-9a-f]{6}\}`).MatchString(theme) {
				t.Errorf("no accent color in %q", theme)
			}
		})
	})
}

func TestImages(t *testing.T) {
	t.Run("get WebP versions", func(t *testing.T) {
		mainSite(t).has(post, `image/webp`)
	})
	t.Run("cover is the Open Graph image", func(t *testing.T) {
		mainSite(t).has(post, `property="og:image"`)
	})
}

func TestFootnotes(t *testing.T) {
	t.Run("get a References heading", func(t *testing.T) {
		mainSite(t).has(post, `<h2>References</h2>`)
	})
	t.Run("no raw return arrows", func(t *testing.T) {
		mainSite(t).lacks(post, `↩`)
	})
}

func TestKaTeX(t *testing.T) {
	t.Run("loads with its extensions", func(t *testing.T) {
		s := mainSite(t)
		s.has(post, `src="/libs/katex@[^ ]*/katex\.min\.`)
		s.has(post, `mhchem\.min\.`)
	})
	t.Run("not on posts without math", func(t *testing.T) {
		mainSite(t).lacks(quiet, `katex`)
	})
}

func TestShortcodes(t *testing.T) {
	t.Run("YouTube falls back to a link", func(t *testing.T) {
		mainSite(t).has(post, `youtube-link`)
	})
	t.Run("social links", func(t *testing.T) {
		mainSite(t).has(post, `mastodon\.social`)
	})
	t.Run("contact form", func(t *testing.T) {
		mainSite(t).has(post, `formspree\.io`)
	})
}

func TestTranslation(t *testing.T) {
	mainSite(t).has("uk/"+post, `<h2>Джерела</h2>`)
}

func TestSearch(t *testing.T) {
	t.Run("page is built", func(t *testing.T) {
		mainSite(t).exists("search/index.html")
	})
	t.Run("form is on the post list", func(t *testing.T) {
		mainSite(t).has("posts/index.html", `<form role="search`)
	})
}
