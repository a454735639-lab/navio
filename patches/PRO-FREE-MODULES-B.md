# Sun-Panel OSS：B 方案改动说明

本次改动基于用户提供的 `sun-panel-oss` 源码包，目标是移除 PRO 管理入口，并开放源码中已经具备实现、仅受前端 PRO 状态限制的功能。

## 已修改

1. **移除 PRO 管理入口**
   - 从系统内置应用列表移除 `ProAuth`。
   - 从应用统一导出入口移除 `ProAuth` 导出。
   - PRO 授权相关后端代码和 API 暂时保留，避免影响其他历史数据或兼容逻辑。

2. **开放自定义搜索引擎数量**
   - 移除未授权时最多添加两个自定义搜索引擎的前端限制。
   - 移除对应 PRO 提示标识。

3. **开放多账号切换数量**
   - 移除未授权时限制添加/切换账号数量的前端判断。
   - 保留原有登录、切换、退出和账号数据逻辑。

4. **开放管理用户数量**
   - 移除前端“最多两个用户”的 PRO 数量判断。
   - 移除对应 PRO 提示标识。
   - 用户创建仍由原有用户管理 API 处理。

5. **开放全局站点设置**
   - 去掉站点标题、Favicon、登录验证码设置的 PRO 遮罩和标签。
   - 原有站点设置读取/保存接口本身已有 OSS 免授权调用路径。

6. **开放自定义 JS/CSS**
   - 去掉自定义 JS/CSS 编辑器的 PRO 遮罩和标签。
   - 前端原本就调用后端 `saveCustomStyleCodeOss` 免授权接口，本次保留该接口，不修改授权服务。

7. **开放 Docker 卡片启动容器选项**
   - 去掉“打开页面时启动容器”设置项的 PRO 遮罩。
   - 去掉 Docker 管理操作列的 PRO 标签。
   - 保留原有 OSS `switchState` 接口；未改动仍受授权保护的旧版接口。

## 保留项

- 没有删除 PRO 授权服务、授权数据、账号认证或 API 定义。
- 没有伪造授权状态，也没有把所有 `isExpired` 判断统一改成已授权。
- 未在当前改动中拆除 Docker 以外的其他授权功能或未知用途的 PRO 代码。

## 验证说明

- 已检查前端 PRO 遮罩/标签引用，目标功能上的遮罩和标签已移除。
- 当前运行环境没有安装项目依赖（`node_modules` 不存在，且未检测到 pnpm），因此尚未执行 `vue-tsc`、ESLint 或前端生产构建。
- 这是一份源码修改包，不是已验证可运行的 Docker 镜像。建议先在测试容器构建并验证，再替换现有服务。

## 建议验证顺序

1. 登录后台，确认系统内置应用中不再显示 PRO 授权入口。
2. 搜索框设置中连续添加三个以上自定义搜索引擎，保存并刷新确认持久化。
3. 添加第三个及更多账号，确认账号列表可正常切换。
4. 添加第三个及更多普通用户，确认 API 保存成功。
5. 修改站点标题、Favicon、登录验证码配置并保存。
6. 修改自定义 JS/CSS 并保存，确认文件内容能够重新读取。
7. 添加 Docker 卡片并启用“打开页面时启动容器”，在测试容器上验证启动行为。

> 注意：自定义 JS 可执行脚本，启用前应确认代码来源可信；Docker 启停会直接影响容器服务，建议先在非生产容器验证。

## V1.1 follow-up: legacy PRO gates and background requests

- `service/biz/globalSetting.go`: `GetSiteStting` now reads the core site settings without a PRO-expiration gate. This keeps login captcha and site branding settings available to normal OSS installations.
- `service/api/api_v1/docker/container.go`: removed the legacy Docker start/stop endpoint's PRO gate so it matches the existing OSS `switchState` endpoint.
- `service/api/api_v1/panel/globalSetting.go`: removed the legacy custom JS/CSS save endpoint's PRO gate so it matches the existing OSS save endpoint.
- `service/biz/proAuth.go`: startup still loads any stored authorization session for backward compatibility, but no longer starts remote registration/heartbeat/renewal background routines.
- Removed the unused GlobalSetting-specific frontend PRO-expiration store; the dedicated PRO management page/store remains in the source for compatibility.
- Ordinary panel account authentication and login CAPTCHA verification are preserved.
