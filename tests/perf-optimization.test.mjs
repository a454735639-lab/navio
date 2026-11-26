import assert from 'node:assert/strict'
import fs from 'node:fs'
const root = new URL('..', import.meta.url).pathname
const read = p => fs.readFileSync(`${root}/${p}`, 'utf8')
const home = read('src/views/home/index.vue')
const api = read('src/api/docker/container.ts')
const chart = read('src/components/apps/DockerManage/components/Container/ResourceChart.vue')
const logs = read('src/components/apps/DockerManage/components/Container/LogViewer.vue')
assert.match(home, /Promise\.all/) // bounded parallel group loading
assert.match(home, /function replaceItemLocally[\s\S]*item\.itemIconGroupId/) // local edit result
assert.match(home, /requestGeneration/) // stale full refresh protection
assert.match(api, /contentType[\s\S]*application\/json/) // 200 business error
assert.match(logs, /onUnmounted/) // lifecycle cleanup
assert.match(logs, /activeTab/) // tab-aware stream
assert.match(chart, /setTimeout/) // no overlapping interval requests
console.log('performance optimization contract tests passed')
