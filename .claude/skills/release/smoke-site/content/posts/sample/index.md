---
title: "Sample post"
description: "Uses most of the theme's features."
date: 2026-01-01
lastmod: 2026-01-02
cover: "cover.png"
coverAlt: "Screenshot of the theme."
toc: true
katex: true
katexExtensions: [ mhchem, copy-tex ]
tags: [ test ]
categories: [ smoke ]
---

## Text

A footnote,[^1] the same footnote again,[^1] and keys: {{< kbd "Ctrl" "C" >}}. An {{< abbr "idk" "I don't know" >}}.

## Code

```go
func main() {
	fmt.Println("tabs are 4 columns")
}
```

> ```
> a plain code block inside a quote
> ```

* A list with code:

  ```
  a plain code block inside a list
  ```

## Table

| Name | Value |
|---|---|
| a | 1 |

## Math

Inline $x^2$, and a block:

{{< katex >}}
$$x_{1,2} = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$$
{{< /katex >}}

## Images

![Alt text.](cover.png)

{{< figure src="cover.png" caption="A *caption*." alt="Alt text." >}}

## Embeds

{{< youtube id="dQw4w9WgXcQ" title="A video" >}}

{{< social >}}

{{< contact-form >}}

[^1]: The note.
