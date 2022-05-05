# yg-render

This is a CLI tool for rendering Yahoo Groups archives exported using
[yahoo-group-archiver](https://github.com/IgnoredAmbience/yahoo-group-archiver).

This tool accepts a directory of RFC 822 `.eml` files as exported by
yahoo-group-archiver and builds a static site for browsing the archive.

[You can see example screenshots of the generated site here.](./docs/screnshots.md)

## Features

- Accessibility. A key design goal is to make the generated site as accessible
  as possible.
- Responsive and mobile-friendly.
- Parses the plain-text email markup used by Yahoo Groups into beautiful and
  semantic HTML.
- The site is paginated with a configurable page size set at build time.
- Client-side full-text search of the archive. This can be disabled at build
  time.
- Supports both a light and dark theme based on your browser preferences.
- All CSS, JavaScript, and fonts are downloaded, minified, optimized, and
  bundled with the site instead of relying in third-party CDNs.

## Usage

This tool has two components:

1. A Go program in `parser/` which parses the archive and builds the HTML.
2. A [gulp](https://gulpjs.com/) pipeline in `pipeline/` which builds,
   minifies, and optimizes all the necessary CSS, JavaScript, and fonts.

### Run the parser

To run the parser, you must first install [Go](https://go.dev/).

To run the parser:

```
cd ./parser
go run . ~/your-yahoo-group/email -t "Your Yahoo Group"
```

To see additional options for the parser:

```
go run . --help
```

This will produce a directory `../output` containing the generated HTML, but
you still need to run the asset pipeline to build the full site.

### Run the asset pipeline

To run the asset pipeline, you must first install
[npm](https://www.npmjs.com/).

To run the asset pipeline:

```
cd ./pipeline
npm install
npx gulp
```

This will produce a directory `../public` containing the generated static site.

## Gotchas

- This tool was written specifically for the Yahoo Group *Haven for the Human
  Amoeba*. While the tool was designed to be generalizable to other Yahoo
  Groups, it hasn't been tested with other data sets.
- The way this tool parses plain-text email markup is best-effort and often
  breaks. The markup used by Yahoo Groups is inconsistent and appears to have
  changed many times over the course of its history. This tool is designed to
  prefer false negatives (ignoring syntax constructs and leaving them as
  literal text) over false positives (potentially mangling text by treating it
  as markup).
- Some messages in Yahoo Groups archives are multipart messages containing both
  plain-text markup and HTML. However, given the long lifespan of Yahoo Groups,
  messages in older groups may use long-deprecated HTML features. For this
  reason, along with the security implications of rendering untrusted HTML and
  accessibility concerns, this tool ignores HTML messages and always attempts
  to parse the plain-text markup instead. Embedded HTML in plain-text markup is
  printed as literal text.
- This tool doesn't attempt to handle attachments in messages.
- If a timestamp in a message is missing a time zone offset, it is assumed to
  be UTC.
- The way the full-text search is implemented currently may not scale well to
  large archives. If performance is a problem, you can disable the search
  functionality at build time.
