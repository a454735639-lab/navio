# Network probe

详见 [docs/NETWORK-PROBE.md](docs/NETWORK-PROBE.md)。

# Sun-Panel OSS 完整版（基于 v1-1 · local edition）

基线：`sun-panel-oss-b-free-v1-1.zip`（B 方案修订版：PRO 遮罩移除、三处后端校验拆掉、启动不再心跳）
本版在其上补齐商业镜像独有的 4 个端点，并修掉全新实例备份失败的 bug。

## 1. 相对 v1-1 的改动（14 个文件）

### 后端

| 端点 | 方法 | 位置 | 说明 |
|---|---|---|---|
| `/api/docker/container/restartContainer` | POST `{id}` | `api/api_v1/docker/container.go` | 重启容器，不做 PRO 校验；60s 超时（docker stop 默认 10s 宽限） |
| `/api/docker/container/getContainerStates` | POST `{containerIds:[], ts}` | 同上 | 响应 `{"containerStates":{id:stats},"ts":<原样回传>}`，与商业版一致；复用 5s 缓存 + 并发上限 8 |
| `/api/about/version/runningInfo` | POST | `api/api_v1/system/about.go` | 返回 `configPath/webPath/customStylePath/sourcePath/sourceTempPath/binaryFilePath`，字段与商业版一致 |
| `/api/panel/itemIcon/getListAllGroup` | POST | `api/api_v1/panel/itemIcon.go` | 返回当前账号所有分组 + 组内图标（`itemInfos`，空组为 `null`，与商业版一致）；`count=len(list)` |

`getContainersStats` 的采集逻辑抽成 `collectContainersState()` 供两个端点共用，行为不变。

### 备份修复（`service/biz/backup.go`）

全新实例从未上传过文件时 `uploads` 目录不存在，打包阶段仍把它塞进 zip 条目 → `stat .../uploads: no such file or directory`。
改为目录存在才加入，与 `custom` 的处理方式一致。修复后空实例备份正常（实测 zip 含 `sqlite/database.db` + `backup.json`）。

### 前端

- `src/api/docker/container.ts`：+`restartContainer`、`getContainerStates`
- `src/api/system/about.ts`：+`RunningInfo` 类型、`runningInfo()`
- `src/api/panel/itemIcon.ts`：+`getListAllGroup()`
- `DockerManage/.../table/RowAction.vue`：启动/停止开关旁增加「重启容器」按钮（NPopconfirm 二次确认，仅 running/paused/exited/created 显示）
- `About/index.vue`：管理员可见「运行信息」区块，展示 6 项路径
- 两套语言包：`iconItem.dockerAction.restartContainer/restartConfirm`、`apps.about.runningInfo`

## 2. 实测结果（iSH aarch64 编译 + 实际运行）

二进制 `/tmp/sp_v11c` → 本目录 `sun-panel-linux-arm64`（26,575,728 字节）

| 检查 | 结果 |
|---|---|
| `go build` | EXIT=0 |
| `POST /api/about/version/runningInfo` | `{"configPath":"./conf","webPath":"./web","customStylePath":"./custom","sourcePath":"./uploads","sourceTempPath":"./runtime/temp","binaryFilePath":"..."}` |
| `POST /api/panel/itemIcon/getListAllGroup` | `count=1`，分组含 `itemInfos`（先 addMultiple 一条再查） |
| `POST /api/panel/backup/backup` | `{"fileName":"sun-panel-backup-20261124234101.zip"}` |
| 下载备份 zip | http=200，2336 字节，含 `sqlite/database.db`、`backup.json` |
| `POST /api/docker/container/getContainerStates` | `{"containerStates":{},"ts":7}`（ts 原样回传） |
| `POST /api/docker/container/restartContainer` | 无 docker socket 时返回 docker 错误（不是 1008，证明 PRO 校验已不在） |
| 站点设置保存/读回 | `{"title":"T-OF-V11",...}` |
| `/api/openness/proIsExpired` | `true`（v1-1 保留真实授权判断；前端遮罩已移除，不影响功能） |
| 三个 TS 文件 + 两个 Vue script 块 | `node --experimental-strip-types --check` 通过 |

未执行（NOT RUN）：
- 前端完整构建（`pnpm install` + `vue-tsc` + `vite build`）——本机内存不够，需在 NAS 上跑 Dockerfile；
- 真实 docker daemon 下的容器启停/重启/批量状态（本机无 docker socket，只验证了接口可达与错误路径）。

## 3. 构建

```bash
# 源码已打补丁，直接 Dockerfile 构建即可（会自动 pnpm build + go-bindata + go build）
docker build -t sun-panel:1.8.1-oss-complete .
```

或在 NAS 上用补丁：

```bash
unzip sun-panel-oss-b-free-v1-1.zip && cd sun-panel-oss-main
patch -p1 < complete-endpoints-and-fixes.patch
docker build -t sun-panel:1.8.1-oss-complete .
```

只跑后端（已有前端 dist 时）：

```bash
cd service
go-bindata-assetfs -o=assets/bindata.go -pkg=assets assets/...
CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o sun-panel main.go
```

## 4. 回滚

- 源码：`patch -p1 -R < complete-endpoints-and-fixes.patch`，或直接用 `sun-panel-oss-b-free-v1-1.zip` 重新解包。
- 运行：本版只新增端点，数据库结构无变化（`itemInfos` 只是查询拼接），可直接换回 v1-1 镜像；已存在的数据与 conf 均可沿用。

## 5. 注意

- 保留 v1-1 的 PRO 授权代码（未空实现）。启动不 Register/Ping，但 `/panel/proAuth/*` 等旧端点仍在；若以后有人从前端调用，仍会访问远端授权服。
- `ProIsExpired()` 仍返回真实判断（无 token → true）。本版通过「移除前端遮罩 + 拆掉 3 处后端校验」达成解锁，不依赖它。
- 上游升级后如果新增了别的 PRO 判断点，需要单独确认（v2 那种全局 `return false` 没有这个问题）。
