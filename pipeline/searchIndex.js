import fs from "fs/promises";
import path from "path";
import FlexSearch from "flexsearch";

const searchFileName = "search.json";

const indexDirName = "search/index/";

const flexsearchSettings = {
  preset: "memory",
  document: {
    id: "id",
    store: ["id", "page", "timestamp", "user", "title", "body"],
    index: ["user", "flair", "year", "title", "body"],
  },
};

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
