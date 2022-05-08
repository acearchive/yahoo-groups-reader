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

const fontWeights = ["300", "400", "500"];

function fontCss() {
    return src(
        fontWeights.map(weight => `node_modules/@fontsource/noto-sans/${weight}.css`)
    ).pipe(dest(path.join(fontDest, "noto-sans")));
}

function fontFiles() {
    return src([
        ...fontWeights.map(weight => `node_modules/@fontsource/noto-sans/files/noto-sans-*-${weight}-normal.woff`),
        ...fontWeights.map(weight => `node_modules/@fontsource/noto-sans/files/noto-sans-*-${weight}-normal.woff2`),
    ]).pipe(dest(path.join(fontDest, "noto-sans", "files")));
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

const font = parallel(fontCss, fontFiles);

exports.clean = parallel(cleanOutput, cleanPublic);

exports.default = series(
    cleanPublic,
    parallel(html, css, js, font, headers, buildSearchIndex),
    cleanOutput,
);