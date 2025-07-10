# Go Framework 静态资源处理系统

这个目录包含了 Go Framework 的静态资源处理系统，使用 Gulp 进行 CSS 和 JavaScript 文件的压缩和优化。

## 目录结构

- `src/`: 源文件目录，包含未压缩的 CSS、JavaScript 和图片文件
  - `css/`: CSS 源文件
  - `js/`: JavaScript 源文件
  - `images/`: 图片文件
- `dist/`: 分发目录，包含压缩和优化后的文件（由 Gulp 自动生成）
- `gulpfile.js`: Gulp 配置文件
- `package.json`: NPM 包配置文件

## 安装依赖

首次使用时，需要安装 Node.js 依赖：

```bash
make install-deps
```

或者直接在 static 目录下运行：

```bash
npm install
```

## 构建静态资源

手动构建静态资源：

```bash
make gulp-build
```

或者在 static 目录下运行：

```bash
npm run build
```

或者在 static 目录下运行：

```bash
npm run watch
```

## 开发流程

1. 在 `src` 目录中编辑源文件
2. Gulp 会自动监视文件变化并重新构建到 `dist` 目录
3. Go 应用程序会从 `dist` 目录提供静态文件

## Gulp 任务

- `gulp clean`: 清理 dist 目录
- `gulp css`: 处理 CSS 文件
- `gulp js`: 处理 JavaScript 文件
- `gulp images`: 复制图片文件
- `gulp other`: 复制其他文件
- `gulp build`: 执行所有构建任务
- `gulp watch`: 监视文件变化并重新构建 