# Tour of Touch

This document is best read sequentially, from top to bottom.

## Basic Syntax

Touch's syntax is similar to Markdown so it should be familiar.
However Touch is designed differently from Markdown and it's easier to think about it in the way its designed.

Markdown has special rules for each element it provides.
(For an example look at the [Emphasis and strong emphasis](https://spec.commonmark.org/0.30/#emphasis-and-strong-emphasis) section in the CommonMark Spec.)
Touch, on the other hand, defines 11 types of elements on which all elements are based on.
As such, each element of a certain type has the same rules as any other element of the same type.
And these rules are more explicit and thus simpler than Markdown's.

Let's look at an example.
One such element type is called "walled".
It is basically a generalized blockquote element from Markdown.
The walled element type simply states that any line starting with a single character is a walled element.
(It is a bit more detailed than that but not much).
Each walled element only differs from others in the delimiter it uses—the single character used to identify it.

Now, let's try to write some walled elements.
First, we will write a blockquote:

```to
> a
> > b
```

The 'b' blockquote is nested inside the 'a' blockquote.

Next, we will write a note.
It has the same rules as a blockquote, except that it uses a '*' delimiter instead of the '>':

```to
* a
* * b
```

Again, the 'b' note is nested inside the 'a' note.

While there are 11 element types, quite a few of them are just slightly different versions of one another.
They differ only in either the content they can contain or the form of the delimiter.
One such variation is the verbatimWalled element type.
It is like the walled element type but can contain only verbatim content.

Below is the BlockComment element of type verbatimWalled (delimiter='/').

```to
/ a
/ / b
```

Here, unlike in the walled examples above, 'b' isn't nested inside the 'a' BlockComment.
Remember that the verbatimWalled can contain only verbatim content.
So 'b' is not even an element.
The above example contains a single BlockComment with the following content:

```
 a
 / b
```

### Element Types

Element types are split into two groups:

- blocks—can be placed only at the start of a line or at the start of another block
- inlines—can be placed only inside blocks

Below are two tables of all block and inline types.
You can just skim over them now and come back to them later.

#### Block Types

|  Element Type   | Delimiter Chars |                                             Description                                              |
|-----------------|-----------------|------------------------------------------------------------------------------------------------------|
| walled          | 1               | each line must be prefixed with the delimiter (like md blockquote)                                   |
| verbatimWalled  | 1               | like walled but can contain only verbatim content                                                    |
| hanging         | 1               | prefixed first line, all subsequent must be indented after the end of delimiter (like md list items) |
| rankedHanging   | >=2             | like hanging but delimiter indicates the level/depth (like md heading)                               |
| fenced          | 1               | like md code block; only verbatim content                                                            |
| verbatimLine    | >=1             | one line with verbatim content                                                                       |
| leaf            | -               | implicit block, present in any non-verbatim content                                                  |

#### Inline Types

| Element Type | Delimiter Chars |                                         Description                                          |
|--------------|-----------------|----------------------------------------------------------------------------------------------|
| uniform      | 2               | starts with delimiter, ends with delimiter or at the end of any parent block (can be nested) |
| escaped      | 2               | like uniform but can contain only verbatim content (cannot be nested)                        |
| prefixed     | >=1             | used only for line break and autolinks (e.g. www.example.test)                               |
| text         | -               | implicit inline                                                                              |

### Elements

See the [default config](config/to.extjson) for reference of all elements that come with Touch by default.
Use the above tables for help, but first read on.

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

### Sticky Elements

There is another variation of elements we haven't discussed yet and that are sticky elements.
One of Touch's golden rules is composition.
And sticky elements fit right into it.

Sticky elements are simply elements that stick to other elements.
They can be either block or inline elements and can stick to any element or only to specific elements.
They offer a way to provide extra information to an existing element.
And depending on the sticky element, they can only be placed either before or after the element they stick to.

In the above element examples we actually used two sticky elements already.
The first was the Subtitle, which sticks only to elements before it:

```to
= Title
_ Subtitle
```

In HTML, this combination of elements is represented in a \<header>.

And the second sticky was the NamedLink:

```to
[[link text]]((link URL))
```

Here we can most clearly see the role composition plays in Touch.
We created an element from two different elements.
In similar ways we can create many elements that can serve many use-cases.

Below is the rest of sticky elements that come by default with Touch:

```to
! id="heading2" class="display" // Attributes
== Heading 2

? term
: description // Description

.image flowers.jpg
+ caption          // Caption
```

### Groups

The composition doesn't stop at sticky elements.
If you look at the [default config](config/to.extjson), you will find elements of type "paragraph", "list", and "sticky".
Looking at our tables of element types neither of these is in there.
That's because they are not element types but groups.

Groups are added by transformers and are created from elements.
Transformers are run after the elements are already parsed.
They traverse the node tree and add new elements (groups) to it.
They find groups by looking for simple patterns.

#### Paragraphs

Paragraphs are added to any leaf element that has a sibling.
They serve as an easy way to incorporate paragraphs into the language.

In languages like HTML we define paragraphs explicitly.
But any such notation would severely harm readability of a lightweight markup language.
It is one of the few cases where we break "the explicitness rule".

In Markdown or rather [CommonMark](https://spec.commonmark.org/0.30/#lists), a list is "loose" if any list item is separated by a blank line and "tight" otherwise.

```md
- a
- b

---

- a

- b
```
 
The example above converts to the following HTML (based on [commonmark.js dingus](https://spec.commonmark.org/dingus/)):

```html
<ul>
<li>a</li>
<li>b</li>
</ul>

---

<ul>
<li>
<p>a</p>
</li>
<li>
<p>b</p>
</li>
</ul>
```

Notice the added paragraphs in the second, "loose" list.
This approach is not intuitive and expected.
Meaning of the elements shouldn't change if a blank line is placed between them (orthogonality).

Touch solves "the paragraph problem" by using a very simple approach.
It adds a paragraph to any block of text that has a sibling element.
After all, we use paragraphs to separate blocks of text from other blocks which this does do.
While this is a complex case that breaks orthogonality, it is easy to predict and intuitive.
Additionally, if you don't like this behaviour, Touch offers an escape hatch through customization.

No paragraph:

```to
a
```

Two paragraphs:

```to
a

b
```

#### Lists

Lists are added around contiguous sequences of the same sibling elements.

```to
- a
- b

- c // blanks are allowed in between items so it's still part of the list
```

#### Stickies

See [Sticky Elements section](#sticky-elements).

### Diving deeper into the composition

Now that we know what stickies and groups are, we can see how they compose together.

```to
? term1
? term2
: description1
: description2
```

The above example represents a single element-StickyDescription (\<dl> in HTML).
First, the Terms and Descriptions are grouped into their own lists.
Then, the two lists are grouped together because the DescriptionList sticks to the TermList.
In short, elements are grouped into lists which are in turn grouped into stickies.

### Aggregates

There is one more way to make elements and it is using aggregators and aggregates.
Aggregators, like transformers, traverse the node tree after the elements are already parsed.
(Actually, they traverse the node tree even after the transformers.)

Aggregators aggregate (or collect) data we are interested in.
Their result is called an aggregate.
Aggregates are used by elements that need data (e.g. table of contents).

To see aggregates in action, you can add a TableOfContents element using the [toc.json](toc.json) config (placed in CWD):

```bash
to build html -config toc.json stdin
```

The toc.json config adds an aggregate that collectes and calculates the sequential numbers of NumberedHeadings:

```json
"Aggregates": {
	"numberedHeadings": {
		"Type": "sequentialNumber",
		"Elements": ["NumberedHeading"]
	}
}
```

This aggregate is used by the TableOfContents element to construct a table of contents.
You can change "NumberedHeading" to "Heading" to aggregate sequential numbers from normal headings instead of the numbered ones.

### Config

While Touch comes with a default set of elements, you can configure and extend it in anyway you want.

Simply create a new configuration file and supply it using the -config flag:

```bash
to build html -config to.json < file.to > file.html
to fmt html -config to.json < file.to 1<> file.to
```

Note: The `1<>` in the fmt command is so that we can write to the same file we read from.

Touch accepts only JSON config files.
However, writing templates in JSON strings is difficult.
As such, Touch configs are usually written in what I call extended JSON.
(Which is converted to JSON before usage; read below.)

#### Extended JSON

Extended JSON (extjson) is just old plain JSON but additionally supports raw multiline strings like you would find in TOML.
Raw multiline strings are denoted by `'''` and are converted to plain JSON strings.
Immediate newline after the opening delimiter is discarded if present.

This extjson:

```
"Templates": {
	"html": '''
<blockquote {{- template "HTMLAttributes" .}}>
	{{template "children" .}}
</blockquote>
'''
}
```

converts to this JSON:

```json
"Templates": {
	"html": "<blockquote {{- template \"HTMLAttributes\" .}}>\n\t{{template \"children\" .}}\n</blockquote>\n"
}
```

Notice the newline after the opening delimiter was removed.

Convert extjson to JSON:

```bash
to tool extjson < to.extjson > to.json
```

If you don't need the raw multiline strings, simply use plain JSON.
extjson provides literally no other benefits.

#### Config schema

Below is the config schema where variables are represented as `<name:type>` and types as `<type>`.

```
{
	"Templates": {
		"<format:string>": "<template:string>"
	},
	"Elements": {
		"<element name:string>": {
			"Disabled":  <bool>,   // disabled=as if the element wasn't present
			"Type":      <string>, // element or group type
			"Delimiter": <string>, // element delimiter (single char or exact)
			"Templates": {
				"<format:string>": "<template:string>"
			}
			// ... for more see config/config.go
		}
	},
	"Aggregates": {
		//... see config/config.go
	}
}
```

You should usually use a single character for the Delimiter, not an exact delimiter.
Touch will construct the actual delimiter based on the character you provide as the Delimiter and the given Type.
Provide exact delimiters only for the following element types:

- hanging
- verbatim line
- prefixed

Example:

```
// single character
"Type": "uniform",
"Delimiter": "*" // actual delimiter will be "**" as per Element Type tables

// exact
"Type": "verbatimLine",
"Delimiter": ".image" // actual delimiter will be ".image"
```

#### How to use configs

Touch accepts multiple configs using a comma-separated list of filepaths:

```bash
to build <format> -config to.json,extra.json stdin
```

The given configs are sequentially shallow merged into the default config (in the given order).
Shallow merge means that they can only add or override whole objects and cannot override specific properties.

#### How to remove default elements

To remove an element from the default config:

1. add an element with the element name you want to remove
1. add property `"Disabled": true`
For example, to remove the Blockquote, apply the following config:

```json
{
	"Elements": {
		"Blockquote": {
			"Disabled": true
		}
	}
}
```

A disabled element is treated as if it wasn't present.
