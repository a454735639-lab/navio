# sun-panel-oss 源码包（main @ 82c5ab14）

## 包内容
- 仓库：https://github.com/a454735639-lab/sun-panel-oss （public，MIT，default branch = main）
- 下载时间：2026-11-25 05:36 (+08:00)
- 上游 commit：`82c5ab14`（pushed 2026-11-24T21:13:06Z，仓库体积 8.0 MB）
- 文件数：515 个文件 / 11.5 MB 解包后（含 public 静态资源、package-lock、pnpm-lock）

| 文件 | 大小 | sha256 |
|---|---|---|
| sun-panel-oss-main.tar.gz | 8.1 MB | bf64738589facb77d7947dac634d18922adb9a32c4527dc07661ae4faef39ab5 |
| sun-panel-oss-main.zip | 8.4 MB | 2b889a51284c423affbe573ea6436a0ae0080dfad92f47f8e92d1571d9011690 |

完整性：`gzip -t` 通过、`unzip -t` 无错误。

## 目录速览
```
sun-panel-oss-main/
├── src/           前端（Vue3 + Vite + TS）：components/apps/{Gallery,PublicGallery,BackupMigration}
├── service/       后端（Go + Gin）：api/api_v1/panel/{gallery,publicGallery,backup}.go
│                  biz/backup.go、lib/file/{base,zip}.go、models/file.go(+IsPublicGallery)
├── patches/       gallery-backup-modules.patch（21 文件 +1513/-8）、BUILD-ON-NAS.md、docker-compose.yml
├── docs-modules/  SPEC-gallery-backup.md（接口规格）
├── .github/       workflows/docker-build.yml（buildx 双架构推 GHCR）
├── Dockerfile     pnpm@9 + go1.21-alpine3.18，容器内自动 go-bindata + 前端 build
└── package.json / pnpm-lock.yaml / tsconfig.json / vite.config.ts
```

## 构建（已含三模块，无需再打补丁）
```bash
cd sun-panel-oss-main
docker build -t sun-panel:1.8.1-oss .          # 首次 8-15 分钟
docker compose up -d                            # 端口 / 挂载见 docker-compose.yml
```

或直接用现成镜像：`docker pull ghcr.io/a454735639-lab/sun-panel-oss:latest`

## 备注
- 该仓库已是「开源 master + 图库/公开图库/备份」的合并树，`patches/gallery-backup-modules.patch` 仅作留档/复现用，**不要**在仓库内再 apply（会冲突）。
- 数据迁移：把老实例 `conf/` 整个拷过来即可，sqlite 会自动补 `is_public_gallery` 列。
