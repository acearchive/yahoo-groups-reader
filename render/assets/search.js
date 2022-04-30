function inputFocus(e, search, suggestions) {
    if (e.key === "/") {
        e.preventDefault();
        search.focus();
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

function showResults(index, search, suggestions) {
    const maxResult = 5;

    const value = search.value;
    const results = index.search(value, {limit: maxResult, enrich: true});

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
        entry.innerHTML = `<a href="${href}"><span class="suggestion-title"></span><span class="suggestion-text"></span></a>`;
        entry.querySelector(".suggestion-title").innerHTML = doc.title;
        entry.querySelector(".suggestion-text").innerHTML = doc.description;

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

    if ((e.code === "ArrowUp" || e.code === "KeyK") && (!keyDefault)) {
        e.preventDefault();
        nextIndex= index > 0 ? index-1 : 0;
        focusableSuggestions[nextIndex].focus();
    } else if ((e.code === "ArrowDown" || e.code === "KeyJ") && (!keyDefault)) {
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
    });

    fetch("/search.json")
        .then(response => response.json())
        .then(searchData => {
            for (const searchFields of searchData) {
                index.add(searchFields);
            }
        });

    search.addEventListener("input", () => showResults(index, search, suggestions), true);
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