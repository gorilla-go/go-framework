const gulp = require('gulp');
const autoprefixer = require('gulp-autoprefixer');
const cleanCSS = require('gulp-clean-css');
const uglify = require('gulp-uglify');
const del = require('del');
const path = require('path');
const watch = require('gulp-watch');

// 清理dist目录
function clean() {
  return del(['dist/**', '!dist']);
}

// 编译CSS
function css() {
  return gulp.src('src/css/**/*.css')
    .pipe(autoprefixer())
    .pipe(cleanCSS())
    .pipe(gulp.dest('dist/css'));
}

// 编译所有JS文件
function js() {
  // 使用gulp直接处理JS文件
  return gulp.src('src/js/**/*.js')
    .pipe(uglify())
    .pipe(gulp.dest('dist/js'));
}

// 复制图片
function images() {
  return gulp.src('src/images/**/*')
    .pipe(gulp.dest('dist/images'));
}

// 复制其他文件
function other() {
  return gulp.src(['src/**/*', '!src/css/**', '!src/js/**', '!src/images/**'])
    .pipe(gulp.dest('dist'));
}

// 处理文件删除
function handleDeleted(file) {
  const filePath = file.path;
  const relativePath = path.relative(path.join(__dirname, 'src'), filePath);
  const distPath = path.join(__dirname, 'dist', relativePath);
  console.log(`源文件已删除: ${filePath}`);
  console.log(`正在删除编译文件: ${distPath}`);
  return del(distPath);
}

// 监视文件变化（包括删除）
function watchFiles() {
  // 监视CSS文件
  watch('src/css/**/*.css', { events: ['add', 'change'] }, css);
  watch('src/css/**/*.css', { events: ['unlink'] }, function(file) {
    handleDeleted(file);
  });

  // 监视JS文件
  watch('src/js/**/*.js', { events: ['add', 'change'] }, js);
  watch('src/js/**/*.js', { events: ['unlink'] }, function(file) {
    handleDeleted(file);
  });

  // 监视图片文件
  watch('src/images/**/*', { events: ['add', 'change'] }, images);
  watch('src/images/**/*', { events: ['unlink'] }, function(file) {
    handleDeleted(file);
  });

  // 监视其他文件
  watch(['src/**/*', '!src/css/**', '!src/js/**', '!src/images/**'], { events: ['add', 'change'] }, other);
  watch(['src/**/*', '!src/css/**', '!src/js/**', '!src/images/**'], { events: ['unlink'] }, function(file) {
    handleDeleted(file);
  });
}

// 构建任务
const build = gulp.series(clean, gulp.parallel(css, js, images, other));

// 开发任务
const dev = gulp.series(build, watchFiles);

// 导出任务
exports.clean = clean;
exports.css = css;
exports.js = js;
exports.images = images;
exports.other = other;
exports.watch = watchFiles;
exports.build = build;
exports.dev = dev;
exports.default = build; 