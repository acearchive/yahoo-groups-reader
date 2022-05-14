import cleanCss from "gulp-clean-css";
import purgeCss from "gulp-purgecss";
import rename from "gulp-rename";
import htmlmin from "gulp-htmlmin";
import del from "delete";
import buildSearchIndex from "./search.js";
import webpack from "webpack-stream";
import concat from "gulp-concat";
import named from "vinyl-named";
import whitelister from "purgecss-whitelister";
import hash from "gulp-hash-filename";
import path from "path";
import captureWebsite from "capture-website";
import inject from "gulp-inject";
import lazypipe from "lazypipe";
import tap from "gulp-tap";
import gulp from "gulp";

const { series, parallel, src, dest } = gulp;

const outputDir = process.env.OUTPUT_DIR ?? "../output"
const publicDir = process.env.PUBLIC_DIR ?? "../public"
const disallowRobots = process.env.DISALLOW_ROBOTS

const outputCss = [];
const outputJs = [];

const cssSources = [
    "node_modules/bootstrap/dist/css/bootstrap.css",
    "src/font.css",
    "src/variables.css",
    "src/global.css",
    "src/components.css",
    "src/thread.css",
    "src/search.css",
]
const cssPipeline = lazypipe()
    .pipe(concat, "bundle.css")
    .pipe(purgeCss, {
        content: [
            path.join(outputDir, "**/*.html"),
            "node_modules/bootstrap/js/src/collapse.js",
        ],
        safelist: [
            ...whitelister([
                "./src/search.css",
            ])
        ],
    })
    .pipe(cleanCss)
    .pipe(hash, { format: "{name}-{hash}.min.css" })
    .pipe(dest, "css", { cwd: publicDir })
    .pipe(tap, file => outputCss.push(file.path));

const jsSources = ["src/*.js"]

const jsPipeline = lazypipe()
    .pipe(named)
    .pipe(webpack, {
        mode: "production",
        devtool: "source-map",
        output: {
            filename: "[name]-[contenthash].min.js"
        },
    })
    .pipe(dest, "js", { cwd: publicDir })
    .pipe(tap, file => file.extname === ".js" && outputJs.push(file.path));

const injectTag = (name) => {
    return `<!-- inject:${name}:{{ext}} -->`
}

const injectJsTransform = (filename) => {
    return `<script src="${filename}" defer></script>`
}

function html() {
    const options = {
        collapseBooleanAttributes: true,
        collapseWhitespace: true,
        removeComments: true,
        removeEmptyAttributes: true,
        removeRedundantAttributes: true,
        sortAttributes: true,
        sortClassName: true,
    };

    return src(path.join(outputDir, "**/*.html"))
        .pipe(inject(src(cssSources).pipe(cssPipeline())))
        .pipe(inject(src([...jsSources, "!src/search.js", "!src/feather.js"]).pipe(jsPipeline()), {
            transform: injectJsTransform,
            removeTags: true,
        }))
        .pipe(inject(src("src/search.js").pipe(jsPipeline()), {
            starttag: injectTag("search"),
            removeTags: true,
            transform: injectJsTransform,
        }))
        .pipe(inject(src("src/feather.js").pipe(jsPipeline()), {
            starttag: injectTag("feather"),
            removeTags: true,
            transform: injectJsTransform,
        }))
        .pipe(htmlmin(options))
        .pipe(dest(publicDir));
}

function headers() {
    return src("src/index.headers")
        .pipe(rename("_headers"))
        .pipe(dest(publicDir));
}

function robots() {
    if (disallowRobots === undefined) {
        return src("src/allow.robots.txt")
            .pipe(rename("robots.txt"))
            .pipe(dest(publicDir));
    } else {
        return src("src/deny.robots.txt")
            .pipe(rename("robots.txt"))
            .pipe(dest(publicDir));
    }
}

function font() {
    return src([
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-300-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-300-normal.woff",
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-400-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-400-normal.woff",
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-500-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-500-normal.woff",
    ]).pipe(dest("font", { cwd: publicDir }));
}

function cleanOutput() {
    return del(outputDir, { force: true });
}

function cleanPublic() {
    return del(publicDir, { force: true });
}

function searchIndex() {
    return buildSearchIndex(outputDir, publicDir);
}

function captureScreenshot() {
    return captureWebsite.file(path.join(publicDir, "index.html"), path.join(publicDir, "screenshot.png"), {
        delay: 1,
        scripts: outputJs,
        styles: outputCss,
    })
}

export const clean = parallel(cleanOutput, cleanPublic);

const main = series(
    cleanPublic,
    parallel(html, font, headers, robots, searchIndex),
    captureScreenshot,
    cleanOutput,
);

export default main;