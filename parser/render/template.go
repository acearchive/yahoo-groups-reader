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
    <meta name="description" content="An archive of the Yahoo Groups community &quot;{{ .Title }}&quot;">
    <title>{{ .Title }}</title>
    <link rel="canonical" href="{{ .Pagination.Current }}">
    {{ if .Pagination.Next }}<link rel="next" href="{{ .Pagination.Next }}">{{ end }}
    {{ if .Pagination.Prev }}<link rel="prev" href="{{ .Pagination.Prev }}">{{ end }}
    <link href="/font/noto-sans/300.css" rel="stylesheet">
    <link href="/font/noto-sans/400.css" rel="stylesheet">
    <link href="/font/noto-sans/500.css" rel="stylesheet">
    <link href="/css/bootstrap.min.css" rel="stylesheet">
    <link href="/css/variables.min.css" rel="stylesheet">
    <link href="/css/global.min.css" rel="stylesheet">
    <link href="/css/components.min.css" rel="stylesheet">
    <link href="/css/thread.min.css" rel="stylesheet">
    {{ if .IncludeSearch }}<link href="/css/search.min.css" rel="stylesheet">{{ end }}
	<script src="/js/bootstrap.min.js" async></script>
    {{ if .IncludeSearch -}}<script src="/js/search.min.js" defer></script>{{- end }}
  </head>
  <body>
	<h1 class="thread-title">{{ .Title }}</h1>
    <nav aria-label="Message thread pages">
      <div class="d-flex justify-content-center align-items-center">
        <ul class="pagination">
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.First }}">
              <span aria-hidden="true">«</span>
              <span class="visually-hidden">First</span>
            </a>
          </li>
	  	{{ if .Pagination.Prev -}}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Prev }}">Prev</a>
          </li>
	  	{{- else -}}
          <li class="page-item disabled">
            <a class="page-link" href="#" tabindex="-1" aria-disabled="true">Prev</a>
          </li>
	  	{{- end }}
          {{ range .Pagination.Pages -}}
          <li class="number-page-item page-item{{ if .IsCurrent }} active{{ end }}"{{ if .IsCurrent }} aria-current="page"{{ end }}>
            <a class="page-link" href="{{ .Path }}">{{ .Number }}</a>
          </li>
          {{- end }}
	  	{{ if .Pagination.Next -}}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Next }}">Next</a>
          </li>
	  	{{- else -}}
          <li class="page-item disabled">
            <a class="page-link" href="#" tabindex="-1" aria-disabled="true">Next</a>
          </li>
	  	{{- end }}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Last }}">
              <span aria-hidden="true">»</span>
              <span class="visually-hidden">Last</span>
            </a>
          </li>
        </ul>
      </div>
      {{ if .IncludeSearch -}}
      <div class="d-flex justify-content-center align-items-center my-3">
        <form id="message-search" role="search" class="flex-grow-1 position-relative">
          <input id="search-input" class="form-control" type="search" placeholder="Search messages..." aria-label="Search messages" aria-controls="search-suggestions" aria-haspopup="true" aria-autocomplete="list" aria-keyshortcuts="/" autocomplete="off">
          <span class="keyboard-hint" aria-hidden="true">/</span>
          <div id="search-suggestions" class="shadow rounded"></div>
        </form>
      </div>
      {{- end }}
	</nav>
    <main class="message-thread">
      {{ range $message := .Messages -}}
      <div id="{{ printf "message-%d" $message.Index }}" class="message">
        <div class="message-header">
          <time class="message-date" datetime="{{ $message.Timestamp }}">{{ $message.FormattedDatetime }}</time>
          <span class="message-count">{{ $message.Number }} / {{ $message.TotalCount }}</span>
        </div>
        <div class="d-flex align-items-start">
          <a class="message-link d-none d-sm-inline" href="{{ printf "#message-%d" $message.Index }}">
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
              <div class="flex-grow-1 d-sm-none me-2">
                <div class="message-author">{{ $message.User }}</div>
                <div class="message-flair">{{ $message.Flair }}</div>
              </div>
              <a class="message-link d-inline d-sm-none" href="{{ printf "#message-%d" $message.Index }}">
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
                <div class="parent-message">
                  <div class="parent-banner d-flex text-nowrap">
                    <button class="btn btn-toggle d-inline-block text-wrap text-start parent-name" data-bs-toggle="collapse" data-bs-target="{{ printf "#parent-quote-%d" $message.Index }}" aria-expanded="false" aria-controls="{{ printf "parent-quote-%d" $message.Index }}">
                      <span class="collapse-arrow me-1" aria-hidden="true">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-caret-right-fill" viewBox="0 0 16 16">
                          <path d="m12.14 8.753-5.482 4.796c-.646.566-1.658.106-1.658-.753V3.204a1 1 0 0 1 1.659-.753l5.48 4.796a1 1 0 0 1 0 1.506z"/>
                        </svg>
                      </span>
                      On <time datetime="{{ $message.Parent.Timestamp }}">{{ $message.Parent.FormattedDatetime }}</time>, {{ $message.Parent.User }} said:
                    </button>
                    <a class="parent-link d-inline-block" href="{{ printf "%s#message-%d" $message.Parent.PagePath $message.Parent.Index }}">
                      <span class="visually-hidden">Parent Comment</span>
                      <div class="inline-icon" aria-hidden="true">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-reply-fill" viewBox="0 0 16 16">
                          <path d="M5.921 11.9 1.353 8.62a.719.719 0 0 1 0-1.238L5.921 4.1A.716.716 0 0 1 7 4.719V6c1.5 0 6 0 7 8-2.5-4.5-7-4-7-4v1.281c0 .56-.606.898-1.079.62z"/>
                        </svg>
                      </div>
                    </a>
                  </div>
                  <blockquote id="{{ printf "parent-quote-%d" $message.Index }}" class="collapse parent-quote">
					{{ $message.Parent.Body }}
                  </blockquote>
                </div>
				{{ end -}}
                <div class="message-body">
                  {{ $message.Body }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
	  {{ end }}
    </main>
    <nav aria-label="Message thread pages">
      <div class="d-flex justify-content-center align-items-center">
        <ul class="pagination">
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.First }}">
              <span aria-hidden="true">«</span>
              <span class="visually-hidden">First</span>
            </a>
          </li>
	  	{{ if .Pagination.Prev -}}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Prev }}">Prev</a>
          </li>
	  	{{- else -}}
          <li class="page-item disabled">
            <a class="page-link" href="#" tabindex="-1" aria-disabled="true">Prev</a>
          </li>
	  	{{- end }}
          {{ range .Pagination.Pages -}}
          <li class="number-page-item page-item{{ if .IsCurrent }} active{{ end }}"{{ if .IsCurrent }} aria-current="page"{{ end }}>
            <a class="page-link" href="{{ .Path }}">{{ .Number }}</a>
          </li>
          {{- end }}
	  	{{ if .Pagination.Next -}}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Next }}">Next</a>
          </li>
	  	{{- else -}}
          <li class="page-item disabled">
            <a class="page-link" href="#" tabindex="-1" aria-disabled="true">Next</a>
          </li>
	  	{{- end }}
          <li class="page-item">
            <a class="page-link" href="{{ .Pagination.Last }}">
              <span aria-hidden="true">»</span>
              <span class="visually-hidden">Last</span>
            </a>
          </li>
        </ul>
      </div>
	</nav>
  </body>
</html>
`

var Template = template.Must(template.New("yg-render").Parse(templateString))
