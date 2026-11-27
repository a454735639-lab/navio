import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'

const mon = readFileSync('service/lib/monitor/monitor.go', 'utf8')
const biz = readFileSync('service/biz/docker.go', 'utf8')
const rank = readFileSync('src/components/apps/DockerManage/components/Container/TrafficRank.vue', 'utf8')
const page = readFileSync('src/components/apps/DockerManage/components/Container/index.vue', 'utf8')
const zh = JSON.parse(readFileSync('src/locales/zh-CN.json', 'utf8'))
const en = JSON.parse(readFileSync('src/locales/en-US.json', 'utf8'))

// 1. 后端网络卡：优先读宿主机 /proc/net/dev，回退容器 netns
assert.match(mon, /IOCountersByFile/, 'must read host net counters by file when mounted')
assert.match(mon, /HostNetDevPath|host-net\/dev/, 'must define host net dev bind-mount path')
assert.match(mon, /os\.Stat/, 'must stat for the mounted file with fallback')
assert.match(mon, /isVirtualNetIF/, 'must filter virtual interfaces')
for (const p of ['veth', 'br-', 'docker', 'virbr'])
  assert.ok(mon.includes(`"${p}"`), `virtual prefix ${p} must be excluded`)
assert.match(mon, /IOCounters\(true\)/, 'must keep container-netns fallback')

// 2. 后端 block IO：除法移出循环
assert.match(biz, /return blkRead \/ 1024 \/ 1024/, 'block IO must divide after summation')
assert.ok(!/blkRead = \(blkRead \+ /.test(biz), 'block IO must not divide inside loop')

// 3. 前端流量排行面板
assert.match(rank, /getContainersStats/, 'rank panel must use batch stats API')
assert.match(rank, /networkRX|networkTX/, 'rank panel must read cumulative rx/tx')
assert.match(rank, /kbToBytes|\* 1024/, 'rank panel must convert backend KB to bytes')
assert.match(rank, /totals/, 'rank panel must show aggregate totals')
assert.match(rank, /NDataTable/, 'rank panel must render a table')

// 4. 入口按钮 + 抽屉已接线
assert.match(page, /TrafficRank/, 'docker page must mount TrafficRank')
assert.match(page, /trafficRankVisible/, 'docker page must control rank drawer')

// 5. i18n key 双语齐全
for (const k of ['trafficRank', 'totalDownload', 'totalUpload', 'totalAll', 'containerCount', 'trafficTip', 'loadFailed'])
  assert.ok(zh.apps.dockerManage[k], `zh missing ${k}`)
for (const k of ['trafficRank', 'totalDownload', 'totalUpload', 'totalAll', 'containerCount', 'trafficTip', 'loadFailed'])
  assert.ok(en.apps.dockerManage[k], `en missing ${k}`)

// 6. bind-mount 必须落到 NAS compose（部署侧由人工/脚本加，仓库内不强校验文件存在）
console.log('network card + traffic rank contract tests passed')
