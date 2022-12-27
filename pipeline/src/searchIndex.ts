import fs from "fs/promises";
import path from "path";
import FlexSearch from "flexsearch";
import hasha from "hasha";

const searchFileName = "search.json";

const indexDirName = (hash: string) => `search/index/${hash}`;

export interface MessageFields {
  id: number;
  page: number;
  timestamp: string;
  user: string;
  flair: string;
  year: string;
  title: string;
  body: string;
}

export type SearchIndex = FlexSearch.Document<MessageFields, string[]>;

type SearchSettings = FlexSearch.IndexOptionsForDocumentSearch<
  MessageFields,
  string[]
>;

const flexsearchSettings: SearchSettings = {
  preset: "memory",
  document: {
    id: "id",
    store: ["id", "page", "timestamp", "user", "title", "body"],
    index: ["user", "flair", "year", "title", "body"],
  },
};

// An object we can hash to generate a cache-buster URL for the search index.
type SearchIndexHashObject = {
  settings: SearchSettings;
  data: ReadonlyArray<MessageFields>;
};

const hashSearchIndex = (obj: SearchIndexHashObject): Promise<string> =>
  hasha.async(JSON.stringify(obj), { encoding: "hex", algorithm: "sha256" });

export const createSearchIndex = async (
  inputDir: string,
  outputDir: string
) => {
  let searchData: ReadonlyArray<MessageFields>;

  try {
    searchData = JSON.parse(
      await fs.readFile(path.join(inputDir, searchFileName), "utf-8")
    );
  } catch {
    return;
  }

  const index: SearchIndex = new FlexSearch.Document(flexsearchSettings);

  await Promise.all(
    searchData.map((fields: MessageFields) => index.addAsync(fields.id, fields))
  );

  const searchIndexHash = await hashSearchIndex({
    settings: flexsearchSettings,
    data: searchData,
  });

  const indexDir = path.join(outputDir, indexDirName(searchIndexHash));
  await fs.mkdir(indexDir, { recursive: true });

  await index.export(async (key, data) => {
    if (data !== undefined) {
      await fs.writeFile(
        path.join(indexDir, key.toString()),
        JSON.stringify(data)
      );
    }
  });
};
