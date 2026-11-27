import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const net = readFileSync('src/components/deskModule/SystemMonitor/AppIconSystemMonitor/Network.vue', 'utf8')
const card = readFileSync('src/views/home/components/ItemCards/DockerCard.vue', 'utf8')
const zh = JSON.parse(readFileSync('src/locales/zh-CN.json', 'utf8'))
const en = JSON.parse(readFileSync('src/locales/en-US.json', 'utf8'))

// 1. Network card: cumulative upload AND download side by side
assert.match(net, /totalSent/, 'network card must show cumulative upload (txBytes)')
assert.match(net, /totalRecv/, 'network card must show cumulative download (rxBytes)')
assert.match(net, /txBytes/, 'network card must read cumulative tx counter')
assert.match(net, /rxBytes/, 'network card must read cumulative rx counter')
const detailBlock = net.slice(net.lastIndexOf('<!-- 累积上传/下载并排显示'), net.indexOf('totalRecv') + 200)
assert.match(detailBlock, /flex/, 'cumulative stats must be in a flex row')

// 2. Docker card: cumulative rx/tx in hover description
assert.match(card, /累积/, 'docker card must mention cumulative traffic in description')
assert.match(card, /networkRX|networkTX/, 'docker card must use cumulative network counters')

// 3. Docker card: right-click menu opens monitor/log drawer
assert.match(card, /查看监控|viewMonitor|monitor/, 'docker card context menu must have a monitor entry')
assert.match(card, /LogViewer/, 'docker card must mount the LogViewer drawer')
assert.match(card, /monitorVisible|showMonitor/, 'docker card must have monitor drawer state')

// i18n keys exist in both locales
assert.ok(zh.iconItem.dockerAction.viewMonitor, 'zh viewMonitor key')
assert.ok(en.iconItem.dockerAction.viewMonitor, 'en viewMonitor key')

console.log('docker card enhancement contract tests passed')
