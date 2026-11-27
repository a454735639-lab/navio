import assert from 'node:assert/strict'
import fs from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const net = fs.readFileSync(join(root, 'src/components/deskModule/SystemMonitor/AppIconSystemMonitor/Network.vue'), 'utf8')

// Network card's info-mode primary visual must occupy the same 56px height as
// the 56px progress ring used by CPU/RAM/Disk cards, so all four cards align.
assert.match(net, /height:\s*56px/, 'network primary visual needs fixed 56px height to match the ring')
assert.match(net, /flex flex-col items-center justify-center[^>]*height:\s*56px/, 'network speed block must be centered within the 56px area')

console.log('monitor alignment contract tests passed')
