const cleanCSS = require("gulp-clean-css");
const purgeCSS = require("gulp-purgecss")
const rename = require("gulp-rename");
const uglify = require("gulp-uglify");
const htmlmin = require("gulp-htmlmin");
const del = require("delete");
const createIndex = require("./search.js");
const { series, parallel, src, dest } = require("gulp");

const destPath = "../public/"
const jsDest = "../public/js/"
const cssDest = "../public/css/"
const fontDest = "../public/font/"

function bootstrapCSS() {
    return src("node_modules/bootstrap/dist/css/bootstrap.css")
        .pipe(purgeCSS({
            content: ["../output/**/*.html"],
        }))
        .pipe(cleanCSS())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest(cssDest));
}

function bootstrapJS() {
   return src("node_modules/bootstrap/dist/js/bootstrap.min.js").pipe(dest(jsDest));
}

function css() {
    return src("src/*.css")
        .pipe(cleanCSS())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest(cssDest));
}

function js() {
    return src("src/*.js")
        .pipe(uglify())
        .pipe(rename({ extname: ".min.js" }))
        .pipe(dest(jsDest));
}

function html() {
    const options = {
        collapseBooleanAttributes: true,
        collapseWhitespace: true,
        conservativeCollapse: true,
        removeComments: true,
        removeEmptyAttributes: true,
        removeRedundantAttributes: true,
        sortAttributes: true,
        sortClassName: true,
    };

    return src("../output/**/*.html")
        .pipe(htmlmin(options))
        .pipe(dest(destPath));
}

function font() {
    return src("node_modules/@fontsource/**").pipe(dest(fontDest));
}

function flexsearch() {
    return src("node_modules/flexsearch/dist/flexsearch.bundle.js")
        .pipe(uglify())
        .pipe(rename({ extname: ".min.js" }))
        .pipe(dest(jsDest))
}

function cleanOutput() {
    return del("../output", { force: true });
}

function cleanPublic() {
    return del(destPath, { force: true });
}

function buildSearchIndex() {
    return createIndex("../output", destPath);
}

exports.clean = parallel(cleanOutput, cleanPublic);

exports.default = series(
    cleanPublic,
    parallel(bootstrapCSS, bootstrapJS, flexsearch, css, js, html, font, buildSearchIndex),
    cleanOutput,
);