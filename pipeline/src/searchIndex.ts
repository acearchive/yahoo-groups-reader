import fs from "fs/promises";
import path from "path";
import FlexSearch from "flexsearch";

const searchFileName = "search.json";
const indexDirName = "search";

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

export const createSearchIndex = async (
  inputDir: string,
  outputDir: string
) => {
  let searchData;

  try {
    searchData = JSON.parse(
      await fs.readFile(path.join(inputDir, searchFileName), "utf-8")
    );
  } catch {
    return;
  }

  const index: SearchIndex = new FlexSearch.Document({
    preset: "memory",
    document: {
      id: "id",
      store: ["id", "page", "timestamp", "user", "title", "body"],
      index: ["user", "flair", "year", "title", "body"],
    },
  });

  await Promise.all(
    searchData.map((fields: MessageFields) => index.addAsync(fields.id, fields))
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
