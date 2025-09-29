const gulp = require('gulp');
const autoprefixer = require('gulp-autoprefixer');
const cleanCSS = require('gulp-clean-css');
const uglify = require('gulp-uglify');
const del = require('del');
const path = require('path');
const fs = require('fs');

// ==================== 配置选项 ====================
const config = {
  // 路径配置
  paths: {
    src: {
      base: 'src',
      css: 'src/css/**/*.css',
      js: 'src/js/**/*.js',
      images: 'src/images/**/*',
      other: ['src/**/*', '!src/css/**', '!src/js/**', '!src/images/**']
    },
    dest: {
      base: 'dist',
      css: 'dist/css',
      js: 'dist/js',
      images: 'dist/images'
    },
    watch: {
      css: 'src/css/**/*.css',
      js: 'src/js/**/*.js',
      images: 'src/images/**/*',
      other: ['src/**/*', '!src/css/**', '!src/js/**', '!src/images/**']
    }
  },

  // 插件选项
  options: {
    autoprefixer: {
      cascade: false,
      grid: true
    },
    cleanCSS: {
      compatibility: 'ie10',
      level: 2,
      inline: ['none']
    },
    uglify: {
      mangle: {
        toplevel: true
      },
      compress: {
        drop_console: false, // 保留 console，开发时有用
        drop_debugger: true,
        pure_funcs: ['console.log']
      },
      output: {
        comments: false
      }
    }
  },

  // 环境配置
  isDevelopment: process.env.NODE_ENV !== 'production',
  isProduction: process.env.NODE_ENV === 'production'
};

// ==================== 工具函数 ====================
/**
 * 错误处理器
 */
function handleError(error) {
  console.error(`❌ 构建错误: ${error.message}`);
  if (config.isDevelopment) {
    console.error(error.stack);
  }
  this.emit('end');
}

/**
 * 记录任务开始
 */
function logTaskStart(taskName) {
  console.log(`🚀 开始执行任务: ${taskName}`);
}

/**
 * 记录任务完成
 */
function logTaskEnd(taskName) {
  console.log(`✅ 任务完成: ${taskName}`);
}

/**
 * 确保目录存在
 */
function ensureDir(dirPath) {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
  }
}

/**
 * 处理文件删除
 */
function handleFileDeleted(filePath) {
  const relativePath = path.relative(path.join(__dirname, config.paths.src.base), filePath);
  const distPath = path.join(__dirname, config.paths.dest.base, relativePath);

  console.log(`🗑️  源文件已删除: ${relativePath}`);

  return del(distPath).then((deleted) => {
    if (deleted.length > 0) {
      console.log(`🗑️  已删除编译文件: ${path.relative(__dirname, deleted[0])}`);
    }
  }).catch((error) => {
    console.error(`❌ 删除文件失败: ${error.message}`);
  });
}

/**
 * 获取文件大小信息
 */
function getFileSizeInfo(filePath) {
  try {
    const stats = fs.statSync(filePath);
    return `${(stats.size / 1024).toFixed(2)} KB`;
  } catch (error) {
    return 'Unknown';
  }
}

// ==================== 构建任务 ====================
/**
 * 清理 dist 目录
 */
function clean() {
  logTaskStart('清理目录');
  return del([`${config.paths.dest.base}/**`, `!${config.paths.dest.base}`])
    .then(() => {
      logTaskEnd('清理目录');
    });
}

/**
 * 编译 CSS
 */
function buildCSS() {
  logTaskStart('编译 CSS');

  // 确保目标目录存在
  ensureDir(config.paths.dest.css);

  let stream = gulp.src(config.paths.src.css, { since: gulp.lastRun(buildCSS) })
    .on('error', handleError);

  // 添加 autoprefixer
  stream = stream.pipe(autoprefixer(config.options.autoprefixer))
    .on('error', handleError);

  // 生产环境压缩 CSS
  if (config.isProduction) {
    stream = stream.pipe(cleanCSS(config.options.cleanCSS))
      .on('error', handleError);
  }

  return stream
    .pipe(gulp.dest(config.paths.dest.css))
    .on('end', () => {
      logTaskEnd('编译 CSS');
    });
}

/**
 * 编译 JavaScript
 */
function buildJS() {
  logTaskStart('编译 JavaScript');

  // 确保目标目录存在
  ensureDir(config.paths.dest.js);

  let stream = gulp.src(config.paths.src.js, { since: gulp.lastRun(buildJS) })
    .on('error', handleError);

  // 生产环境压缩 JS
  if (config.isProduction) {
    stream = stream.pipe(uglify(config.options.uglify))
      .on('error', handleError);
  }

  return stream
    .pipe(gulp.dest(config.paths.dest.js))
    .on('end', () => {
      logTaskEnd('编译 JavaScript');
    });
}

/**
 * 处理图片
 */
function buildImages() {
  logTaskStart('处理图片');

  // 确保目标目录存在
  ensureDir(config.paths.dest.images);

  return gulp.src(config.paths.src.images, { since: gulp.lastRun(buildImages) })
    .on('error', handleError)
    .pipe(gulp.dest(config.paths.dest.images))
    .on('end', () => {
      logTaskEnd('处理图片');
    });
}

/**
 * 处理其他文件
 */
function buildOther() {
  logTaskStart('处理其他文件');

  return gulp.src(config.paths.src.other, { since: gulp.lastRun(buildOther) })
    .on('error', handleError)
    .pipe(gulp.dest(config.paths.dest.base))
    .on('end', () => {
      logTaskEnd('处理其他文件');
    });
}

// ==================== 监视任务 ====================
/**
 * 监视文件变化
 */
function watchFiles() {
  console.log('👀 开始监视文件变化...');

  // 监视 CSS 文件
  const cssWatcher = gulp.watch(config.paths.watch.css, buildCSS);
  cssWatcher.on('unlink', handleFileDeleted);
  cssWatcher.on('change', (filePath) => {
    console.log(`📝 CSS 文件变化: ${path.relative(__dirname, filePath)}`);
  });

  // 监视 JS 文件
  const jsWatcher = gulp.watch(config.paths.watch.js, buildJS);
  jsWatcher.on('unlink', handleFileDeleted);
  jsWatcher.on('change', (filePath) => {
    console.log(`📝 JS 文件变化: ${path.relative(__dirname, filePath)}`);
  });

  // 监视图片文件
  const imagesWatcher = gulp.watch(config.paths.watch.images, buildImages);
  imagesWatcher.on('unlink', handleFileDeleted);
  imagesWatcher.on('change', (filePath) => {
    console.log(`📝 图片文件变化: ${path.relative(__dirname, filePath)}`);
  });

  // 监视其他文件
  const otherWatcher = gulp.watch(config.paths.watch.other, buildOther);
  otherWatcher.on('unlink', handleFileDeleted);
  otherWatcher.on('change', (filePath) => {
    console.log(`📝 其他文件变化: ${path.relative(__dirname, filePath)}`);
  });

  console.log('👀 文件监视已启动');
}

// ==================== 复合任务 ====================
/**
 * 完整构建任务
 */
const build = gulp.series(
  clean,
  gulp.parallel(buildCSS, buildJS, buildImages, buildOther)
);

/**
 * 开发任务
 */
const dev = gulp.series(build, watchFiles);

/**
 * 生产构建任务
 */
function setBuildEnv(done) {
  process.env.NODE_ENV = 'production';
  config.isProduction = true;
  config.isDevelopment = false;
  console.log('🏭 设置为生产环境');
  done();
}

const buildProd = gulp.series(setBuildEnv, build);

// ==================== 信息任务 ====================
/**
 * 显示构建信息
 */
function info() {
  console.log('📋 构建配置信息:');
  console.log(`   环境: ${config.isDevelopment ? '开发环境' : '生产环境'}`);
  console.log(`   源目录: ${config.paths.src.base}`);
  console.log(`   输出目录: ${config.paths.dest.base}`);
  console.log(`   CSS 压缩: ${config.isProduction ? '启用' : '禁用'}`);
  console.log(`   JS 压缩: ${config.isProduction ? '启用' : '禁用'}`);
  return Promise.resolve();
}

/**
 * 统计构建结果
 */
function stats() {
  console.log('📊 构建统计:');

  const distFiles = {
    css: path.join(config.paths.dest.css, '*.css'),
    js: path.join(config.paths.dest.js, '*.js')
  };

  try {
    const cssFiles = require('glob').sync(distFiles.css);
    const jsFiles = require('glob').sync(distFiles.js);

    console.log(`   CSS 文件: ${cssFiles.length} 个`);
    cssFiles.forEach(file => {
      console.log(`     ${path.basename(file)}: ${getFileSizeInfo(file)}`);
    });

    console.log(`   JS 文件: ${jsFiles.length} 个`);
    jsFiles.forEach(file => {
      console.log(`     ${path.basename(file)}: ${getFileSizeInfo(file)}`);
    });

  } catch (error) {
    console.log('   无法获取统计信息（需要安装 glob 依赖）');
  }

  return Promise.resolve();
}

// ==================== 任务导出 ====================
// 基础任务
exports.clean = clean;
exports.css = buildCSS;
exports.js = buildJS;
exports.images = buildImages;
exports.other = buildOther;

// 监视任务
exports.watch = watchFiles;

// 复合任务
exports.build = build;
exports.dev = dev;
exports.prod = buildProd;

// 信息任务
exports.info = info;
exports.stats = stats;

// 默认任务
exports.default = build;

// ==================== 任务帮助 ====================
function help() {
  console.log(`
📖 可用的 Gulp 任务:

🏗️  构建任务:
   gulp build     - 完整构建项目
   gulp prod      - 生产环境构建（启用压缩）
   gulp clean     - 清理 dist 目录

🔧 单独任务:
   gulp css       - 只编译 CSS
   gulp js        - 只编译 JavaScript
   gulp images    - 只处理图片
   gulp other     - 只处理其他文件

👀 开发任务:
   gulp dev       - 开发模式（构建 + 监视）
   gulp watch     - 只启动文件监视

📊 信息任务:
   gulp info      - 显示构建配置
   gulp stats     - 显示构建统计
   gulp help      - 显示此帮助信息

🚀 快速开始:
   npm run build  - 构建项目
   npm run watch  - 开发模式
  `);
  return Promise.resolve();
}

exports.help = help;