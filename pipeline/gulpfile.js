const cleanCss = require("gulp-clean-css");
const purgeCss = require("gulp-purgecss")
const rename = require("gulp-rename");
const htmlmin = require("gulp-htmlmin");
const del = require("delete");
const createIndex = require("./search.js");
const webpack = require("webpack-stream");
const path = require("path");

const { series, parallel, src, dest } = require("gulp");

const outputDir = process.env.OUTPUT_DIR ?? "../output"
const publicDir = process.env.PUBLIC_DIR ?? "../public"

const jsDest = path.join(publicDir, "js")
const cssDest = path.join(publicDir, "css")
const fontDest = path.join(publicDir, "font")

function bootstrapCss() {
    return src("node_modules/bootstrap/dist/css/bootstrap.css")
        .pipe(purgeCss({
            content: [
                path.join(outputDir, "**/*.html"),
                "./node_modules/bootstrap/js/src/collapse.js",
            ],
        }))
        .pipe(cleanCss())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest(cssDest));
}

function css() {
    return src("src/*.css")
        .pipe(cleanCss())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest(cssDest));
}

function bootstrapJs() {
    return src("src/bootstrap.js")
        .pipe(webpack({
            mode: "production",
            devtool: "source-map",
            output: {
                filename: "bootstrap.min.js"
            },
        }))
        .pipe(dest(jsDest));
}

function searchJs() {
    return src("src/search.js")
        .pipe(webpack({
            mode: "production",
            devtool: "source-map",
            output: {
                filename: "search.min.js"
            }
        }))
        .pipe(dest(jsDest))
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
        .pipe(htmlmin(options))
        .pipe(dest(publicDir));
}

const fontWeights = ["300", "400", "500"];

function fontCss() {
    return src(
        fontWeights.map(weight => `node_modules/@fontsource/noto-sans/${weight}.css`)
    ).pipe(dest(`${fontDest}/noto-sans`));
}

function fontFiles() {
    return src([
        ...fontWeights.map(weight => `node_modules/@fontsource/noto-sans/files/noto-sans-*-${weight}-normal.woff`),
        ...fontWeights.map(weight => `node_modules/@fontsource/noto-sans/files/noto-sans-*-${weight}-normal.woff2`),
    ]).pipe(dest(`${fontDest}/noto-sans/files`));
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

const bootstrap = parallel(bootstrapCss, bootstrapJs);

exports.clean = parallel(cleanOutput, cleanPublic);

exports.default = series(
    cleanPublic,
    parallel(bootstrap, searchJs, css, html, font, buildSearchIndex),
    cleanOutput,
);