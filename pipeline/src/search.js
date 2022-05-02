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
        search.ariaExpanded = "false";
        suggestions.innerHTML = "";
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

function formatTimestamp(date) {
    return new Intl.DateTimeFormat("en-UK", {
        timeZone: "UTC",
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
        timeZoneName: "short",
    }).format(date)
}

async function showResults(index, search, suggestions) {
    const maxResult = 5;

    const value = search.value;
    const results = await index.searchAsync(value, {limit: maxResult, enrich: true});

    search.ariaExpanded = "true";
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
              </span>
              <time class="suggestion-timestamp" datetime="${doc.timestamp}"></time>
              <span class="suggestion-title"></span>
            </span>
            <span class="suggestion-text"></span>
          </a>`;

        entry.querySelector(".suggestion-user").appendChild(document.createTextNode(doc.user));
        entry.querySelector(".suggestion-timestamp").appendChild(document.createTextNode(
            formatTimestamp(new Date(doc.timestamp))
        ));
        entry.querySelector(".suggestion-title").appendChild(document.createTextNode(doc.title));
        entry.querySelector(".suggestion-text").appendChild(document.createTextNode(doc.body));

        suggestions.appendChild(entry);
        if (suggestions.childElementCount === maxResult) break;
    }
}

function suggestionFocus(e, search, suggestions) {
    const focusableSuggestions = suggestions.querySelectorAll("a");
    const focusable = [...focusableSuggestions];
    const index = focusable.indexOf(document.activeElement);

    const hasSuggestions = search.ariaExpanded === "true";

    let nextIndex = 0;

    if (hasSuggestions && e.code === "ArrowUp") {
        e.preventDefault();
        nextIndex = index > 0 ? index-1 : 0;
        focusableSuggestions[nextIndex].focus();
    } else if (hasSuggestions && e.code === "ArrowDown") {
        e.preventDefault();
        nextIndex = index+1 < focusable.length ? index+1 : index;
        focusableSuggestions[nextIndex].focus();
    }
}

const indexFileNames = [
    "reg",
    "store",
    "body.cfg",
    "body.ctx",
    "body.map",
    "flair.cfg",
    "flair.ctx",
    "flair.map",
    "title.cfg",
    "title.ctx",
    "title.map",
    "user.cfg",
    "user.ctx",
    "user.map",
]

function importIndex(index) {
    return Promise.all(indexFileNames.map(fileName => fetch(`/search/${fileName}`)
        .then(response => response.text())
        .then(indexData => index.import(fileName, indexData))));
}

const importIndexOnce = (() => {
    let isIndexed = false;
    return async (index) => {
        if (!isIndexed) {
            isIndexed = true;
            await importIndex(index);
        }
    }
})()

function indexSearch(search, suggestions) {
    const index = new FlexSearch.Document({
        preset: "memory",
        document: {
            id: "id",
            store: ["id", "page", "timestamp", "user", "title", "body"],
            index: ["user", "flair", "title", "body"],
        },
    });

    search.addEventListener("focus", async () => await importIndexOnce(index), true);
    search.addEventListener("input", async () => await showResults(index, search, suggestions), true);
    suggestions.addEventListener("click", () => acceptSuggestion(suggestions), true);
}

const searchForm = document.querySelector("#message-search");
const searchInput = searchForm?.querySelector("#search-input");
const searchSuggestions = searchForm?.querySelector("#search-suggestions");

if (searchInput && searchSuggestions) {
    searchForm.addEventListener("submit", (e) => e.preventDefault());
    document.addEventListener("keydown", (e) => inputFocus(e, searchInput, searchSuggestions));
    document.addEventListener("keydown", (e) => suggestionFocus(e, searchInput, searchSuggestions));
    document.addEventListener("click", (event) => {
        if (!searchSuggestions.contains(event.target)) {
            searchInput.ariaExpanded = "false";
            searchSuggestions.innerHTML = "";
        }
    });

    indexSearch(searchInput, searchSuggestions);
}