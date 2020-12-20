package renderer_test

import (
	"testing"
	"to/parser"
	"to/renderer"
)

func TestHTML(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "lines",
			input: `
Tibsey is eating eucalyptus leaves.
Tibsey is going shopping.
Tibsey likes to sleep.
`,
			want: `<p>Tibsey is eating eucalyptus leaves.<br>
Tibsey is going shopping.<br>
Tibsey likes to sleep.</p>
`,
		},
		{
			name:  "emphasis",
			input: "Tibsey __loves__ sleeping.",
			want:  "<p>Tibsey <em>loves</em> sleeping.</p>\n",
		},
		{
			name:  "strong",
			input: "Tibsey **loves** sleeping.",
			want:  "<p>Tibsey <strong>loves</strong> sleeping.</p>\n",
		},
		{
			name:  "strong in emphasis",
			input: "Tibsey **__loves__** sleeping.",
			want:  "<p>Tibsey <strong><em>loves</em></strong> sleeping.</p>\n",
		},
		{
			name:  "heading 1",
			input: "= Koalas",
			want:  "<h1>Koalas</h1>\n",
		},
		{
			name:  "heading 3",
			input: "=== Koalas",
			want:  "<h3>Koalas</h3>\n",
		},
		{
			name:  "heading 8",
			input: "======== Koalas",
			want: `<div role="heading" aria-level="8">Koalas</div>
`,
		},
		{
			name:  "link",
			input: "<Koalas, facts, and photos><https://www.nationalgeographic.com/animals/mammals/k/koala/>",
			want: `<p><a href="https://www.nationalgeographic.com/animals/mammals/k/koala/">Koalas, facts, and photos</a></p>
`,
		},
		{
			name: "code block",
			input: `
` + "``" + `ts, button.ts
function displayButton(): void {
	const button = document.querySelector("button")
	button.style.display = "block"
	// ...
` + "``",
			want: `<div>
	button.ts
	<pre><code>
function displayButton(): void {
	const button = document.querySelector("button")
	button.style.display = "block"
	// ...
	</code></pre>
</div>
`,
		},
		{
			name: "unordered list",
			input: `
- Tuesday:
	- milk
	- sugar
	- bananas
	 5 or 6?
`,
			// each text has a space in front as in To
			want: `<ul>
	<li>
		 Tuesday:
		<ul>
			<li>
				 milk
			</li>
			<li>
				 sugar
			</li>
			<li>
				 bananas
				5 or 6?
			</li>
		</ul>

	</li>
</ul>
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(tc.input)
			doc := p.ParseDocument()
			html := renderer.HTML(doc, 0)

			if html != tc.want {
				t.Errorf("\ngot:\n%swant:\n%s", html, tc.want)
			}
		})
	}
}
