---
name: release
description: Release a new version of the Color Your World Hugo theme. Commits what's pending, builds a test site, checks translations and the README, makes a signed tag on master, pushes it, writes release notes and publishes the GitHub release. Use when the user asks to release, ship, or "ship it as usual".
argument-hint: '[version, e.g. 2.1.0; picked from the commits if left out]'
disable-model-invocation: true
---

# Release

Running this skill means the user wants you to commit, tag and push. Their global rules still
apply. Ask before any force-push, retag or rewrite of pushed history. No AI attribution in
commits or tags, no skipping hooks, and no IPs or secrets in the release notes.

If a step fails, stop and tell the user. Don't work around it.

The theme has no build output and no CI. Sites pull it in as a Git submodule or a clone, so a
release is a signed tag on `master` and release notes on GitHub. No file holds the version,
only the tag does, so there's no release commit.

`gh` is installed and logged in. Use it to publish the release.

## 1. Before you tag

- Be on `master`. Run `git fetch origin`, and make sure `master` isn't behind or diverged.
- **Commit pending work first.** One Conventional Commit per change, signed with
  `git commit -S` like the rest of the history. Never stage secret files, `public/`,
  `resources/` or scratch files. Leave `.claude/` out unless the user says otherwise. If you're
  not sure a change belongs in the release, ask.
- **pre-commit.** The hooks run on every commit from `.git/hooks/pre-commit`. If that file is
  missing, ask the user to run `! pre-commit install`, since you can't change `.git/hooks`
  yourself. The first commit on a machine downloads the hooks, which is third-party code. If a
  hook changes files, look at the change, stage it again and commit again. If gitleaks or
  `detect-private-key` complains, stop and tell the user.
- **Every new commit is signed.** Run `git log --format='%h %G? %s' <last-tag>..HEAD`. Every
  line must show `G`. To sign unpushed commits that show `N`, run
  `git rebase --exec 'git commit --amend --no-edit -S' <last pushed commit>`. It only adds
  signatures. If an unsigned commit is already pushed, stop and ask.
- **Hugo.** `hugo version` must say Extended, and must be at least the `min` under
  `[module.hugoVersion]` in `config.toml`.
- **Test site.** Build the test site from this skill's folder:
  ```bash
  .claude/skills/release/smoke.sh <scratchpad>/smoke
  ```
  The folder must be empty or missing. Every check must pass, and no build may print `WARN`
  or `ERROR` lines. The site turns on most features, comes in English and Ukrainian, and has
  no network, so YouTube has to fall back to its placeholder. The script builds it three
  times: as is, with giscus set for Ukrainian only, and with `baseURL = "/"`. The last one
  needs Chromium. Without it, that check says `skip`, and you mention it in the report. If a
  commit since the last tag adds a feature the test site doesn't use, add it to `smoke-site/`
  and add a check to `smoke.sh`, in a `chore(skills): ...` commit.
- **Tested Hugo version.** The Requirements section of `README.md` says which Hugo version the
  theme was tested up to. If yours is newer and the test site passed, update that line in a
  `docs(readme): ...` commit.
- **Minimum Hugo version.** If a commit since the last tag uses a Hugo function, method or
  setting for the first time, check Hugo's docs to see if the minimum version has it. If it
  doesn't, or you can't tell, ask the user. They can raise the minimum, which breaks sites and
  means changing `config.toml`, the README badge, Requirements and Upgrading. Or they can test
  with the minimum Hugo, which needs a download they approve.
- **Translations.** All four files need the same keys:
  ```bash
  for f in i18n/*.toml; do grep -o '^\[[^]]*\]' "$f" | sort > "<scratchpad>/keys-$(basename "$f")"; done
  diff <scratchpad>/keys-en.toml <scratchpad>/keys-de.toml
  diff <scratchpad>/keys-en.toml <scratchpad>/keys-pt.toml
  diff <scratchpad>/keys-en.toml <scratchpad>/keys-uk.toml
  ```
  If `diff` prints anything, a translation is missing. Ask the user for it, or ask if English
  is fine for now.
- **README.** Go through `git log <last-tag>..HEAD` and check that `README.md` covers it, in
  Features, Requirements, Appearance, Writing content, Search, Comments, Shortcodes,
  Customization, Configuration reference, and Upgrading from the original theme. Every new or
  changed param goes in Configuration reference or its own section, with its default. Front
  matter params go in the Front matter table. Files meant for overriding go in the
  Customization table. Anything that breaks sites gets a line in Upgrading. Add what's missing
  in a `docs(readme): ...` commit. Only update the screenshots in `images/` when the user asks.

## 2. Version

- The last tag is `git describe --tags --abbrev=0`.
- Use the version in `$ARGUMENTS`. If there isn't one, pick it from the commits since the last
  tag. A `!` or `BREAKING CHANGE` means a major bump, any `feat` means minor, and anything else
  means patch. Say in one line which version you picked and why.
- A change breaks sites when it renames or removes a param, raises the minimum Hugo version,
  changes the variables in a file sites override (`variables.scss`, `font-face.scss`,
  `giscus.scss`), or moves a layout or partial sites override. If a commit like that has no
  `!`, say so and ask before you pick the version.
- `git status --porcelain --untracked-files=no` must print nothing.

## 3. Signed tag

Put the tag on the latest commit of `master`. It's annotated and signed. The message is the
version, a blank line, and a short summary of the release in plain words, wrapped at 72
columns. `git cat-file -p v2.0.0` shows an example. Write the message to a file in the
scratchpad:

```bash
git tag -s vX.Y.Z -F <scratchpad>/tag-message
git tag -v vX.Y.Z
```

The signature must be Good, from `0xC00C0F0D647BD3C6`. If gpg times out waiting for the
passphrase, ask the user to run the command themselves with a `!` in front.

## 4. Push

```bash
git push origin master
git push origin vX.Y.Z
```

If either push gets rejected, stop and ask. Never force-push. There's no CI, so nothing runs
after the push.

## 5. Release notes

Write them from the commits since the last tag, in plain words. When it isn't obvious why a
change matters, say why. Start with a one-line summary, and only keep the sections that have
something in them. Save them to `dist/vX.Y.Z/RELEASE_NOTES.md`, without the outer fence below.
`dist/` is in the user's global gitignore. Leave everything else in it alone, like
`dist/config/`.

````markdown
One-line summary of the release.

### Breaking
- What changed, why, and what sites have to change: params, overridden files or Hugo version.

### Highlights
- New features, the ones site owners notice first.

### Fixes
- What was broken and what happens now.

### Changes
- Other changes site owners may notice, like defaults, styles or translations.

### Removed
- Params, shortcodes or files that are gone.

### Upgrade
```bash
git -C themes/color-your-world fetch --tags
git -C themes/color-your-world checkout vX.Y.Z
git -C themes/color-your-world tag -v vX.Y.Z
```

**Full Changelog**: https://github.com/sprytnyk/hugo-theme-color-your-world/compare/<last-tag>...vX.Y.Z
````

## 6. GitHub release

A release is public, so ask first with a pop-up: publish now, publish as a draft, or leave it
to the user. The title is `Color Your World vX.Y.Z`. GitHub adds the source archives on its
own, so there's nothing to upload:

```bash
gh release create vX.Y.Z --verify-tag --title "Color Your World vX.Y.Z" \
  --notes-file dist/vX.Y.Z/RELEASE_NOTES.md
gh release view vX.Y.Z --json url,isDraft,name
```

Add `--draft` if the user picks a draft. Tags up to v2.0.2 have no GitHub release, and that's
fine.

## 7. Report

Keep it short:

- the version, and that the tag signature is Good
- the commits in the release
- that pre-commit and the test site passed, and which Hugo version built it
- the release link, or the release notes in one fenced `markdown` block if the user publishes
  them
- what's left over, like the test site in the scratchpad, and any check you skipped
- last, the commands that update the user's site to the new version, ready to paste:
  ```bash
  git submodule update --remote themes/color-your-world
  git -C themes/color-your-world describe --tags   # should print vX.Y.Z
  git add themes/color-your-world
  git commit -S -m "chore(theme): update to vX.Y.Z"
  git push
  ```
