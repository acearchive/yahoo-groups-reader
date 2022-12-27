import fs from "fs/promises";
import path from "path";
import { sha256 } from "crypto-hash";
import FlexSearch from "flexsearch";

const searchFileName = "search.json";

const indexDirName = "search/index/";

const searchSettings = {
  preset: "memory",
  document: {
    id: "id",
    store: ["id", "page", "timestamp", "user", "title", "body"],
    index: ["user", "flair", "year", "title", "body"],
  },
};

const readSearchData = async (inputDir) => {
  try {
    return JSON.parse(
      await fs.readFile(path.join(inputDir, searchFileName), "utf-8")
    );
  } catch {
    return undefined;
  }
};

export const calculateSearchIndexEtag = async (inputDir) => {
  const searchData = await readSearchData(inputDir);

  if (searchData === undefined) return undefined;

  // The etag should be generated from both the actual data and the settings
  // used to generate the index.
  const hashObj = {
    settings: searchSettings,
    data: searchData,
  };

  const hash = await sha256(JSON.stringify(hashObj), { outputFormat: "hex" });

  // This is a weak etag because the combination of the search settings and the
  // data represents a sort of semantic equality, but is not a byte-for-byte
  // guarantee.
  return `W/"${hash}"`;
};

export const createSearchIndex = async (inputDir, outputDir) => {
  const searchData = await readSearchData(inputDir);

  if (searchData === undefined) return;

  const index = new FlexSearch.Document(searchSettings);

  await Promise.all(
    searchData.map((fields) => index.addAsync(fields.id, fields))
  );

  const indexDir = path.join(outputDir, indexDirName);
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
