const cleanCss = require("gulp-clean-css");
const purgeCss = require("gulp-purgecss")
const rename = require("gulp-rename");
const htmlmin = require("gulp-htmlmin");
const del = require("delete");
const createIndex = require("./search.js");
const webpack = require("webpack-stream");
const concat = require("gulp-concat");
const named = require("vinyl-named");
const whitelister = require("purgecss-whitelister");
const path = require("path");

const { series, parallel, src, dest } = require("gulp");

const outputDir = process.env.OUTPUT_DIR ?? "../output"
const publicDir = process.env.PUBLIC_DIR ?? "../public"

const jsDest = path.join(publicDir, "js")
const cssDest = path.join(publicDir, "css")
const fontDest = path.join(publicDir, "font")

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
        .pipe(htmlmin(options))
        .pipe(dest(publicDir));
}

function css() {
    return src([
        "node_modules/bootstrap/dist/css/bootstrap.css",
        "src/font.css",
        "src/variables.css",
        "src/*.css",
    ])
        .pipe(concat("bundle.css"))
        .pipe(purgeCss({
            content: [
                path.join(outputDir, "**/*.html"),
                "node_modules/bootstrap/js/src/collapse.js",
            ],
            safelist: [
                ...whitelister([
                    "./src/search.css",
                ])
            ],
        }))
        .pipe(cleanCss())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest(cssDest));
}

function js() {
    return src("src/*.js")
        .pipe(named())
        .pipe(webpack({
            mode: "production",
            devtool: "source-map",
            output: {
                filename: "[name].min.js"
            },
        }))
        .pipe(dest(jsDest));
}

function headers() {
    return src("src/index.headers")
        .pipe(rename("_headers"))
        .pipe(dest(publicDir));
}

function font() {
    return src([
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-300-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-300-normal.woff",
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-400-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-400-normal.woff",
        "node_modules/@fontsource/noto-sans/files/noto-sans-latin-500-normal.woff2",
        "node_modules/@fontsource/noto-sans/files/noto-sans-all-500-normal.woff",
    ]).pipe(dest(fontDest));
}

function cleanOutput() {
    return del(outputDir, { force: true });
}

function cleanPublic() {
    return del(publicDir, { force: true });
}

function buildSearchIndex() {
    return createIndex(outputDir, publicDir);
}

exports.clean = parallel(cleanOutput, cleanPublic);

exports.default = series(
    cleanPublic,
    parallel(html, css, js, font, headers, buildSearchIndex),
    cleanOutput,
);