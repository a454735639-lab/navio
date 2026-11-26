import assert from 'node:assert/strict'
import fs from 'node:fs'
const root = new URL('..', import.meta.url).pathname
const read = p => fs.readFileSync(`${root}/${p}`, 'utf8')
const hold = read('src/components/common/HoldConfirmButton/index.vue')
const edit = read('src/components/deskModule/SystemMonitor/Edit/index.vue')
const row = read('src/components/apps/DockerManage/components/Container/table/RowAction.vue')
const zh = read('src/locales/zh-CN.json')
const en = read('src/locales/en-US.json')

// HoldConfirmButton: real progress, pointer cancel resets, a11y, reduced motion
assert.match(hold, /pointerdown/)
assert.match(hold, /pointerup/)
assert.match(hold, /pointercancel/)
assert.match(hold, /requestAnimationFrame/)
assert.match(hold, /prefers-reduced-motion/)
assert.match(hold, /role="button"/)
assert.match(hold, /tabindex="0"/)
assert.match(hold, /duration/)
assert.match(hold, /stroke-dashoffset/) // SVG progress ring, no layout thrash

// Restart uses hold-to-confirm instead of a single click popconfirm
assert.match(row, /HoldConfirmButton/)
assert.doesNotMatch(row, /NPopconfirm[\s\S]*doRestart/)

// System monitor save: submitLoading driven by real request lifecycle
assert.match(edit, /submitLoading\.value = true/)
assert.match(edit, /submitLoading\.value = true[\s\S]*try \{[\s\S]*finally \{[\s\S]*submitLoading\.value = false/)

// i18n keys exist in both languages
assert.match(zh, /"holdToRestart"/)
assert.match(en, /"holdToRestart"/)
assert.match(zh, /"releaseToCancel"/)
assert.match(en, /"releaseToCancel"/)

console.log('interaction feedback contract tests passed')
