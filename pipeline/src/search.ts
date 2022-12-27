import FlexSearch from "flexsearch";

interface MessageFields {
  id: number;
  page: number;
  timestamp: string;
  user: string;
  flair: string;
  year: string;
  title: string;
  body: string;
}

type SearchIndex = FlexSearch.Document<MessageFields, string[]>;

const userIcon = `
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-person-circle" viewBox="0 0 16 16">
    <path d="M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
    <path fill-rule="evenodd" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z"/>
  </svg>
`;

interface NetworkInformation {
  saveData?: boolean;
}

// Whether the client prefers reducing data usage. This uses an experimental
// browser API that is currently only implemented in Chromium browsers, so we
// cannot rely on it always being implemented.
//
// The Network Information API as a whole seems to have been largely abandoned,
// but the `saveData` property in particular seems to be seeing more recent
// support as part of the Save Data API.
//
// Network Information API: https://wicg.github.io/netinfo/
// Save Data API: https://wicg.github.io/savedata/
const preferSaveData = () => {
  const netInfo: NetworkInformation | undefined = (navigator as any).connection;

  // To be extra safe, we should default to `true`, when the API isn't
  // implemented.
  return netInfo?.saveData ?? true;
};

const inputFocus = (
  e: KeyboardEvent,
  search: HTMLInputElement,
  suggestions: HTMLElement
) => {
  if (e.key === "/" && search !== document.activeElement) {
    e.preventDefault();
    search.focus();
    search.scrollIntoView();
  }

  if (e.key === "Escape") {
    search.blur();
    suggestions.innerHTML = "";
  }
};

const acceptSuggestion = (suggestions: HTMLElement) => {
  while (suggestions.lastChild) {
    suggestions.removeChild(suggestions.lastChild);
  }

  return false;
};

const hrefForMessage = (id: string, page: number) => {
  if (page === 1) {
    return `/#message-${id}`;
  } else {
    return `/${page}/#message-${id}`;
  }
};

const formatTimestamp = (date: Date) => {
  return new Intl.DateTimeFormat("en-UK", {
    timeZone: "UTC",
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
    timeZoneName: "short",
  }).format(date);
};

const showResults = async (
  index: SearchIndex,
  search: HTMLInputElement,
  suggestions: HTMLElement
) => {
  const maxResult = 10;

  await importIndexOnce(index);

  const value = search.value;
  const results = await index.searchAsync(value, {
    limit: maxResult,
    enrich: true,
  });

  suggestions.innerHTML = "";

  const flatResults: Record<string, MessageFields> = {};

  results.forEach((result) => {
    result.result.forEach((r) => {
      flatResults[hrefForMessage(r.doc.id.toString(), r.doc.page)] = r.doc;
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

    entry
      .querySelector(".suggestion-user")
      ?.appendChild(document.createTextNode(doc.user));

    entry
      .querySelector(".suggestion-timestamp")
      ?.appendChild(
        document.createTextNode(formatTimestamp(new Date(doc.timestamp)))
      );

    entry
      .querySelector(".suggestion-title")
      ?.appendChild(document.createTextNode(doc.title));

    entry
      .querySelector(".suggestion-text")
      ?.appendChild(document.createTextNode(doc.body));

    suggestions.appendChild(entry);
    if (suggestions.childElementCount === maxResult) break;
  }
};

const suggestionFocus = (e: KeyboardEvent, suggestions: HTMLElement) => {
  const focusableSuggestions = suggestions.querySelectorAll("a");
  const focusable = [...focusableSuggestions];
  const index = focusable.indexOf(document.activeElement as HTMLAnchorElement);

  const hasSuggestions = suggestions.childElementCount > 0;

  let nextIndex = 0;

  if (hasSuggestions && e.code === "ArrowUp") {
    e.preventDefault();
    nextIndex = index > 0 ? index - 1 : 0;
    focusableSuggestions[nextIndex].focus();
  } else if (hasSuggestions && e.code === "ArrowDown") {
    e.preventDefault();
    nextIndex = index + 1 < focusable.length ? index + 1 : index;
    focusableSuggestions[nextIndex].focus();
  }
};

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
];

const importIndex = async (index: SearchIndex) => {
  await Promise.all(
    indexFileNames.map((fileName) =>
      fetch(`/search/${fileName}`)
        .then((response) => response.text())
        .then((indexData) => index.import(fileName, JSON.parse(indexData)))
    )
  );
};

const importIndexOnce = (() => {
  let importIndexPromise: Promise<void>;

  return async (index: SearchIndex): Promise<void> => {
    if (importIndexPromise === undefined) {
      importIndexPromise = importIndex(index);
    }
    await importIndexPromise;
  };
})();

const indexSearch = async (
  search: HTMLInputElement,
  suggestions: HTMLElement
) => {
  const index: SearchIndex = new FlexSearch.Document({
    preset: "memory",
    document: {
      id: "id",
      store: ["id", "page", "timestamp", "user", "title", "body"],
      index: ["user", "flair", "year", "title", "body"],
    },
  });

  // Download and import the search index eagerly instead of lazily for a
  // better experience when the user isn't trying to save data.
  if (!preferSaveData()) {
    await importIndexOnce(index);
  }

  search.addEventListener(
    "focus",
    async () => await importIndexOnce(index),
    true
  );
  search.addEventListener(
    "input",
    async () => await showResults(index, search, suggestions),
    true
  );
  suggestions.addEventListener(
    "click",
    () => acceptSuggestion(suggestions),
    true
  );
};

const searchForm: Element | undefined =
  document.querySelector("#message-search") ?? undefined;

const searchInput: HTMLInputElement | undefined =
  searchForm?.querySelector("#search-input") ?? undefined;

const searchSuggestions: HTMLElement | undefined =
  searchForm?.querySelector("#search-suggestions") ?? undefined;

if (searchInput && searchSuggestions) {
  searchForm?.addEventListener("submit", (e) => e.preventDefault());

  document.addEventListener("keydown", (e) =>
    inputFocus(e, searchInput, searchSuggestions)
  );

  document.addEventListener("keydown", (e) =>
    suggestionFocus(e, searchSuggestions)
  );

  document.addEventListener("click", (event) => {
    if (!searchSuggestions.contains(event.target as Node | null)) {
      searchSuggestions.innerHTML = "";
    }
  });

  indexSearch(searchInput, searchSuggestions);
}
