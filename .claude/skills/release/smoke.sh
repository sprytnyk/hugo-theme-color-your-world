#!/usr/bin/env bash
# Builds the test site in smoke-site/ with this theme and checks the output.
# Run it from anywhere inside the theme repo:
#
#   .claude/skills/release/smoke.sh <empty or missing folder>
#
# It builds three versions of the site:
#
#   main          giscus on for all languages
#   one-language  giscus on for Ukrainian only
#   root          main again, with baseURL = "/" (checked in Chromium, if found)
#
# It fails on any Hugo error or warning, and on any failed check.

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <output folder>" >&2
  exit 2
fi

out=$1
here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
theme=$(git -C "$here" rev-parse --show-toplevel)

if [[ -e $out && -n $(ls -A "$out") ]]; then
  echo "$out isn't empty" >&2
  exit 2
fi

# Copies the test site to $1 and links the theme in
make_site() {
  mkdir -p "$1/themes"
  cp -r "$here/smoke-site/." "$1/"
  ln -s "$theme" "$1/themes/color-your-world"
  cp "$theme/images/screenshot.png" "$1/content/posts/sample/cover.png"
}

# Builds the site in $1 into $2, with any extra Hugo flags after that
build() {
  local site=$1 public=$2 log=$2.log
  shift 2

  if ! hugo --source "$site" --destination "$public" --gc --minify "$@" >"$log" 2>&1; then
    cat "$log"
    echo "FAIL: the build of $site failed" >&2
    exit 1
  fi

  if grep -E '^(WARN|ERROR)' "$log"; then
    echo "FAIL: the build of $site printed warnings" >&2
    exit 1
  fi
}

hugo version
echo

make_site "$out/main"
cat "$here/giscus.toml" >>"$out/main/hugo.toml"
build "$out/main" "$out/main/public"

make_site "$out/one-language"
sed 's/^\[params\.giscus\]/[languages.uk.params.giscus]/' "$here/giscus.toml" \
  >>"$out/one-language/hugo.toml"
build "$out/one-language" "$out/one-language/public"

build "$out/main" "$out/root" --baseURL /

failed=0

check() {
  local name=$1
  shift

  if "$@" >/dev/null 2>&1; then
    echo "ok    $name"
  else
    echo "FAIL  $name"
    failed=1
  fi
}

public=$out/main/public
post=$public/posts/sample/index.html

# --minify drops quotes around attribute values, so they're optional here
giscus='class="?giscus[" >]'

check "post has comments" grep -Eq "$giscus" "$post"
check "comments: false hides them" \
  bash -c "! grep -Eq '$giscus' '$public/posts/quiet/index.html'"
check "standalone page has no comments" \
  bash -c "! grep -Eq '$giscus' '$public/about/index.html'"
check "giscus themes are built" \
  test -s "$public/css/giscus-dark.css" -a -s "$public/css/giscus-light.css"

check "images get WebP versions" grep -q 'image/webp' "$post"
check "cover has an Open Graph image" grep -q 'property="og:image"' "$post"
check "footnotes get a References heading" grep -q '<h2>References</h2>' "$post"
check "no raw footnote return arrows" bash -c "! grep -q '↩' '$post'"
check "KaTeX and its extensions load" \
  bash -c "grep -Eq 'src=\"?/libs/katex@[^ ]*/katex\.min\.' '$post' && grep -q 'mhchem\.min\.' '$post'"
check "no KaTeX on posts without math" bash -c "! grep -q katex '$public/posts/quiet/index.html'"
check "YouTube falls back to a link" grep -q 'youtube-link' "$post"
check "social and contact form render" \
  bash -c "grep -q 'mastodon.social' '$post' && grep -q 'formspree.io' '$post'"
check "Ukrainian translation is built" \
  grep -q '<h2>Джерела</h2>' "$public/uk/posts/sample/index.html"
check "search page is built" test -s "$public/search/index.html"
check "search form is on the post list" grep -Eq '<form role="?search' "$public/posts/index.html"

# main.js is built once for all languages, so it must load giscus even
# when only one language sets it up
one=$out/one-language/public
check "giscus for one language: no comments in English" \
  bash -c "! grep -Eq '$giscus' '$one/posts/sample/index.html'"
check "giscus for one language: comments in Ukrainian" \
  grep -Eq "$giscus" "$one/uk/posts/sample/index.html"
check "giscus for one language: main.js loads giscus" \
  bash -c "grep -q giscus.app '$one'/js/main.*.js"

# With baseURL = "/", the theme URLs are paths. giscus loads them from its
# own domain, so main.js has to turn them into full URLs on this site
browser=$(command -v chromium || command -v chromium-browser || command -v google-chrome || true)

if [[ -n $browser ]]; then
  port=$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')
  python3 -m http.server "$port" --bind 127.0.0.1 -d "$out/root" >/dev/null 2>&1 &
  server=$!
  trap 'kill "$server" 2>/dev/null' EXIT

  for _ in $(seq 20); do
    curl -s -o /dev/null "http://127.0.0.1:$port/" && break
    sleep 0.25
  done

  # No network beyond this machine, so nothing reaches giscus.app
  timeout 60 "$browser" --headless --disable-gpu \
    --host-resolver-rules='MAP * ~NOTFOUND, EXCLUDE 127.0.0.1' \
    --virtual-time-budget=5000 --dump-dom \
    "http://127.0.0.1:$port/posts/sample/" >"$out/root.dom" 2>/dev/null || true

  check "baseURL /: theme path stays a path in the page" \
    grep -Eq 'data-theme-dark="?/css/giscus-dark.css' "$out/root/posts/sample/index.html"
  check "baseURL /: giscus gets a full theme URL" \
    grep -q "data-theme=\"http://127.0.0.1:$port/css/giscus-" "$out/root.dom"
else
  echo "skip  baseURL /: no Chromium found"
fi

echo

if [[ $failed -ne 0 ]]; then
  echo "FAIL: some checks failed. The sites are in $out" >&2
  exit 1
fi

echo "All checks passed. The sites are in $out"
