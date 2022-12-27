import fs from "fs/promises";
import path from "path";
import FlexSearch from "flexsearch";
import { sha256 } from "crypto-hash";

const searchFileName = "search.json";

const indexDirName = (hash) => `search/index/${hash}`;

const flexsearchSettings = {
  preset: "memory",
  document: {
    id: "id",
    store: ["id", "page", "timestamp", "user", "title", "body"],
    index: ["user", "flair", "year", "title", "body"],
  },
};

const hashSearchIndex = async (searchSettings, searchData) =>
  await sha256(JSON.stringify({ settings: searchSettings, data: searchData }), {
    outputFormat: "hex",
  });

export const createSearchIndex = async (inputDir, outputDir) => {
  let searchData;

  try {
    searchData = JSON.parse(
      await fs.readFile(path.join(inputDir, searchFileName), "utf-8")
    );
  } catch {
    return;
  }

  const index = new FlexSearch.Document(flexsearchSettings);

  await Promise.all(
    searchData.map((fields) => index.addAsync(fields.id, fields))
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
