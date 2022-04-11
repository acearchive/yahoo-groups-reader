package render

import (
	"html/template"
)

const templateString = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Haven for the Human Amoeba</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Noto+Sans:wght@300;400;500&display=swap" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossorigin="anonymous">
    <link href="thread.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p" crossorigin="anonymous"></script>
  </head>
  <body>
    <main class="message-thread">
      <h1 class="thread-title">{{ .Title }}</h1>
      {{ range $message := .Messages -}}
      <div id="{{ $message.ID }}" class="message">
        <div class="message-header">
          <time class="message-date" datetime="{{ $message.Timestamp }}">{{ $message.FormattedDatetime }}</time>
          <span class="message-count">{{ $message.Index }} / {{ $message.TotalCount }}</span>
        </div>
        <div class="d-flex align-items-start">
          <a class="message-link d-none d-sm-inline" href="#{{ $message.ID }}">
            <span class="visually-hidden">Permalink</span>
            <div aria-hidden="true">
              <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" fill="currentColor" class="bi bi-link-45deg" viewBox="0 0 16 16">
                <path d="M4.715 6.542 3.343 7.914a3 3 0 1 0 4.243 4.243l1.828-1.829A3 3 0 0 0 8.586 5.5L8 6.086a1.002 1.002 0 0 0-.154.199 2 2 0 0 1 .861 3.337L6.88 11.45a2 2 0 1 1-2.83-2.83l.793-.792a4.018 4.018 0 0 1-.128-1.287z"/>
                <path d="M6.586 4.672A3 3 0 0 0 7.414 9.5l.775-.776a2 2 0 0 1-.896-3.346L9.12 3.55a2 2 0 1 1 2.83 2.83l-.793.792c.112.42.155.855.128 1.287l1.372-1.372a3 3 0 1 0-4.243-4.243L6.586 4.672z"/>
              </svg>
            </div>
          </a>
          <div class="card flex-grow-1">
            <div class="card-header d-flex align-items-center">
              <div class="d-none d-sm-flex me-2" aria-hidden="true">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-person-circle" viewBox="0 0 16 16">
                  <path d="M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
                  <path fill-rule="evenodd" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z"/>
                </svg>
              </div>
              <div class="flex-grow-1 align-items-baseline d-none d-sm-flex">
                <span class="message-author">{{ $message.User }}</span>
                <span class="message-flair ms-1">{{ $message.Flair }}</span>
              </div>
              <div class="flex-grow-1 d-sm-none">
                <div class="message-author">{{ $message.User }}</div>
                <div class="message-flair">{{ $message.Flair }}</div>
              </div>
              <a class="message-link d-inline d-sm-none" href="#{{ $message.ID }}">
                <span class="visually-hidden">Permalink</span>
                <div aria-hidden="true">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-link-45deg" viewBox="0 0 16 16">
                    <path d="M4.715 6.542 3.343 7.914a3 3 0 1 0 4.243 4.243l1.828-1.829A3 3 0 0 0 8.586 5.5L8 6.086a1.002 1.002 0 0 0-.154.199 2 2 0 0 1 .861 3.337L6.88 11.45a2 2 0 1 1-2.83-2.83l.793-.792a4.018 4.018 0 0 1-.128-1.287z"/>
                    <path d="M6.586 4.672A3 3 0 0 0 7.414 9.5l.775-.776a2 2 0 0 1-.896-3.346L9.12 3.55a2 2 0 1 1 2.83 2.83l-.793.792c.112.42.155.855.128 1.287l1.372-1.372a3 3 0 1 0-4.243-4.243L6.586 4.672z"/>
                  </svg>
                </div>
              </a>
            </div>
            <div class="card-body">
			  {{ if $message.Title -}}
              <h2 class="card-title message-title">{{ $message.Title }}</h2>
			  {{- end }}
              <div class="card-text">
				{{ if $message.Parent -}}
                <div class="parent-ref">
                  <a href="#{{ $message.Parent.ID }}" class="parent-name">
					On {{ $message.Parent.FormattedDate }} at {{ $message.Parent.FormattedTime }}, {{ $message.Parent.User }} said:
				  </a>
                  <blockquote class="parent-quote">
					{{ $message.Parent.Body }}
                  </blockquote>
                </div>
				{{ end -}}
                <p class="message-body">
                  {{ $message.Body }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
	  {{ end }}
    </main>
  </body>
</html>
`

var Template = template.Must(template.New("yg-render").Parse(templateString))
