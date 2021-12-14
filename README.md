# 📜 Touch

[Go Reference](https://pkg.go.dev/github.com/touchmarine/to) | [Playground](http://touchlabs.io/play) | [MIT License](LICENSE)

Touch is a lightweight markup language.
It uses puncutation characters to inscribe meaning to otherwise plain text.
The grammar is simple to parse, allowing for easy tooling.

Touch in a few bullet points:

- familiar syntax—follows existing conventions
- easily extensible via JSON for teams that need extra elements (removes the need for inline HTML or other flavors of the language)
- default config includes only common elements (including comments!) which should satisfy 80% of use cases
- comes with auto-formatting (think prettier or gofmt)
## Quick Start

Online playground: http://touchlabs.io/play

### Locally

1. Install the latest binary from [Releases](https://github.com/touchmarine/to/releases) or run ``go get github.com/touchmarine/to`` if you have Go installed.
1. Run ``to version`` to verify it&#39;s working.
1. Run ``to build html < file.to > file.html`` to convert Touch to HTML.
Use ``to help`` for details.

### Auto-Formatting

```bash
to fmt < file.to 1<> file.to                # 1<> to write to same file we read from
to fmt -linelength 80 < file.to 1<> file.to # hard-wrap at 80 columns
```

### Elements

See the [default config](config/to.extjson) for reference of all elements that come with Touch by default.

### Block Elements

This is a quick reference of some common block elements:

```to
/ this is a block comment

= Title
_ Subtitle

== h2
=== h3

/ numbered headings (prefixed with 1 and 1.1)
## h2
### h3

> blockquote
* note
- list
1. numbered list

/ code block
`js
function num() {
	return 1
}
`

/ preformatted block
'
      ___________________________
    < I'm an expert in my field. >
      ---------------------------
          \   ^__^
           \  (oo)\_______
              (__)\       )\/\
                  ||----w |
                  ||     ||
'
/ art from https://developer.mozilla.org/en-US/docs/Web/HTML/Element/pre#example
```

### Inline Elements

This is a quick reference of some common inline elements:

```to
// an inline comment //
__emphasis__ // italics //
**strong**   // bold //
``code``
((link))
[[link text]]((link URL))

a \ // line break //
b

/ autolinks
www.example.test
http://example.test
https://example.test
```

## Learn More

Checkout the [TOUR](TOUR.md).

## Get in Touch

Ha, get it?
In __Touch__?
Anyway, you can reach me at scout at touchlabs.io.
I would love to hear your thoughts.
