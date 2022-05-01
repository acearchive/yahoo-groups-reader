const userIcon = `
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-person-circle" viewBox="0 0 16 16">
    <path d="M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
    <path fill-rule="evenodd" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z"/>
  </svg>
`;

function inputFocus(e, search, suggestions) {
    if (e.key === "/" && search !== document.activeElement) {
        e.preventDefault();
        search.focus();
        search.scrollIntoView();
    }

    if (e.key === "Escape") {
        search.blur();
        suggestions.classList.add("d-none");
    }
}

function acceptSuggestion(suggestions) {
    while (suggestions.lastChild) {
        suggestions.removeChild(suggestions.lastChild);
    }

    return false;
}

function hrefForMessage(id, page) {
    if (page === 1) {
        return `/#message-${id}`
    } else {
        return `/${page}/#message-${id}`
    }
}

async function showResults(index, search, suggestions) {
    const maxResult = 5;

    const value = search.value;
    const results = await index.searchAsync(value, {limit: maxResult, enrich: true});

    suggestions.classList.remove("d-none");
    suggestions.innerHTML = "";

    const flatResults = {};
    results.forEach(result => {
        result.result.forEach(r => {
            flatResults[hrefForMessage(r.doc.id, r.doc.page)] = r.doc;
        });
    });

    for (const href in flatResults) {
        const doc = flatResults[href];

        const entry = document.createElement("div");
        entry.innerHTML = `
          <a href="${href}">
            <span class="suggestion-header">
              <span class="suggestion-user">
                <span class="inline-icon me-1" aria-hidden="true">${userIcon}</span>
                ${doc.user}
              </span>
              <time class="suggestion-timestamp" datetime="${doc.timestamp}">
                ${new Intl.DateTimeFormat([], { dateStyle: "medium", timeStyle: "short" }).format(new Date(doc.timestamp))}
              </time>
              <span class="suggestion-title">${doc.title}</span>
            </span>
            <span class="suggestion-text">${doc.body}</span>
          </a>`;

        suggestions.appendChild(entry);
        if (suggestions.childElementCount === maxResult) break;
    }
}

function suggestionFocus(e, suggestions) {
    const focusableSuggestions = suggestions.querySelectorAll("a");
    const focusable = [...focusableSuggestions];
    const index = focusable.indexOf(document.activeElement);

    const keyDefault = suggestions.classList.contains("d-none");

    let nextIndex = 0;

    if (e.code === "ArrowUp" && !keyDefault) {
        e.preventDefault();
        nextIndex= index > 0 ? index-1 : 0;
        focusableSuggestions[nextIndex].focus();
    } else if (e.code === "ArrowDown" && !keyDefault) {
        e.preventDefault();
        nextIndex= index+1 < focusable.length ? index+1 : index;
        focusableSuggestions[nextIndex].focus();
    }
}

function indexSearch(search, suggestions) {
    const index = new FlexSearch.Document({
        tokenize: "forward",
        cache: 100,
        document: {
            id: "id",
            store: ["id", "page", "timestamp", "user", "title", "body"],
            index: ["user", "flair", "title", "body"],
        },
        worker: true,
    });

    fetch("/search.json")
        .then(response => response.json())
        .then(searchData => Promise.all(searchData.map(fields => index.addAsync(fields))));

    search.addEventListener("input", async () => await showResults(index, search, suggestions), true);
    suggestions.addEventListener("click", () => acceptSuggestion(suggestions), true);
}

const searchInput = document.querySelector("#message-search > .search-input");
const searchSuggestions = document.querySelector("#message-search > .search-suggestions");

if (searchInput && searchSuggestions) {
    document.addEventListener("keydown", (e) => inputFocus(e, searchInput, searchSuggestions));
    document.addEventListener("keydown", (e) => suggestionFocus(e, searchSuggestions));
    document.addEventListener("click", function(event) {
        if (!searchSuggestions.contains(event.target)) {
            searchSuggestions.classList.add("d-none");
        }
    });

    indexSearch(searchInput, searchSuggestions);
}