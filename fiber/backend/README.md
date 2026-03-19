# 后台管理

基于 SvelteKit + Tailwind CSS 的后台管理界面，支持站点管理和分类管理。

## 技术栈

- SvelteKit
- Tailwind CSS
- shadcn-svelte 风格组件（Button、Card、Input）

## 功能

- **站点管理**：增删改查站点
- **分类管理**：多级分类树，增删改查

## 布局

左右结构：左侧导航栏，右侧内容区。

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:5173）
npm run dev
```

## 配置

复制 `.env.example` 为 `.env`，配置 Fiber API 地址：

```
VITE_API_URL=http://localhost:3000/api
```

确保 Fiber 服务已启动（默认端口 3000）。

## 构建

```bash
npm run build
npm run preview
```
