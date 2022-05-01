const cleanCSS = require("gulp-clean-css");
const purgeCSS = require("gulp-purgecss")
const rename = require("gulp-rename");
const { parallel, src, dest } = require("gulp");

function bootstrap() {
    return src("node_modules/bootstrap/dist/css/bootstrap.css")
        .pipe(purgeCSS({
            content: ["../output/**/*.html"],
        }))
        .pipe(cleanCSS())
        .pipe(rename({ extname: ".min.css" }))
        .pipe(dest("../output/"));
}

exports.default = parallel(bootstrap);
