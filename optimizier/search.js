const SEARCH_DATA_FILE = "search.json";
const SEARCH_INDEX_DIR = "search-index";
const SEARCH_INDEX_KEY_FILE = "index-keys.json";

const fs = require("fs/promises");
const path = require("path");
const FlexSearch = require("flexsearch")

async function appendToJsonFile(outputPath, obj) {
    const currentObj = JSON.parse(await fs.readFile(outputPath, { encoding: "utf-8" }));
    currentObj.push(obj);
    await fs.writeFile(outputPath, JSON.stringify(currentObj));
}

async function createIndex(outputPath) {
    const index = new FlexSearch.Document({
        tokenize: "forward",
        cache: 100,
        document: {
            id: "id",
            store: ["id", "page", "timestamp", "user", "title", "body"],
            index: ["user", "flair", "title", "body"],
        },
    });

    const searchDataPath = path.join(outputPath, SEARCH_DATA_FILE);
    const searchData = JSON.parse(await fs.readFile(searchDataPath, "utf-8"));

    const indexDirPath = path.join(outputPath, SEARCH_INDEX_DIR);
    await Promise.all(searchData.map(fields => index.addAsync(fields)));

    await fs.mkdir(indexDirPath);

    const indexKeysPath = path.join(outputPath, SEARCH_INDEX_KEY_FILE);
    await fs.writeFile(indexKeysPath, JSON.stringify([]));

    await index.export(async (key, data) => {
        const indexPartPath = path.join(indexDirPath, key);
        if (data !== undefined) {
            await Promise.all([
                fs.writeFile(indexPartPath, data),
                appendToJsonFile(indexKeysPath, key),
            ]);
        }
    });

    await fs.rm(searchDataPath);
}

module.exports = { createIndex };