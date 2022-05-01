# yg-render

This is a CLI tool for rendering Yahoo Groups archives exported using
[yahoo-group-archiver](https://github.com/IgnoredAmbience/yahoo-group-archiver).

This tool accepts a directory of RFC 822 `.eml` files as exported by
yahoo-group-archiver and builds a static site for browsing the archive.

![Screenshot of an example of the generated site](./docs/hha-screenshot.png)

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
- Supports minifying the generated HTML/CSS/JS.
- Dates and times are localized automatically on the client, with a fallback
  option for clients with scripts disabled.

## Usage

```
Render an exported Yahoo Groups archive as HTML

This accepts the path of the directory containing the .eml files.

Usage:
  yg-render [options] archive-path

Flags:
  -h, --help            help for yg-render
      --minify          Minify the output HTML/CSS/JS files
      --no-search       Disable the search functionality in the generated site
  -o, --output string   The directory to write the rendered output to (default ".")
      --page-size int   The maximum number of messages per page (default 50)
  -t, --title string    The title of the group (default "Yahoo Group")
  -v, --verbose         Print verbose output.
      --version         version for yg-render
```

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
- This tool doesn't attempt to handle attachments in messages.
- If a timestamp in the archive is missing a time zone offset, it is treated as
  UTC.
- The way the full-text search is implemented currently may not scale well to
  large archives. If performance is a problem, you can disable the search
  functionality at build time.
- Third-party fonts, CSS, and JS are included via CDNs. Hosting these assets
  yourself is not currently supported.
