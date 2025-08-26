<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { ElMessage } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from '@xterm/addon-fit'
import 'xterm/css/xterm.css'
import type { Server } from '../types/server'

/* ---------- 状态 ---------- */
const conDis = ref(false)
const terminalRef = ref<HTMLDivElement>()
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
const isConnected = ref(false)

/* ---------- props / 会话 ---------- */
const props = defineProps<{ server: Server }>()
const server = reactive<Server>({ ...props.server })
const sessionId = ref<string>(props.server.sessionId)

/* ---------- 生命周期 ---------- */
onMounted(() => {
  initializeTerminal()
  setupEventListeners()
  connectServer()
})

onUnmounted(() => {
  cleanup()
})

/* ---------- 终端初始化 ---------- */
const initializeTerminal = () => {
  try {
    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", "Microsoft YaHei", monospace',
      theme: {
        background: '#FFFFFF',
        foreground: '#212327',
        cursor: '#212327',
        selectionBackground: '#264F78'
      },
      scrollback: 2000,
      convertEol: true
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)

    if (terminalRef.value) {
      terminal.open(terminalRef.value)
      nextTick(() => fitAddon!.fit())
    }

    /* 键盘透传 */
    terminal.onData(data => {
      if (!isConnected.value) return
      window.go.main.App.SendRaw(data, sessionId.value)
    })

    const ro = new ResizeObserver(() => {
      if (!terminal || !fitAddon) return

      // 检查终端是否可见
      if (terminalRef.value && terminalRef.value.offsetParent === null) {
        return // 终端不可见，不执行调整
      }

      fitAddon.fit()
      if (isConnected.value) {
        window.go.main.App.SendResize(terminal.rows, terminal.cols, sessionId.value)
      }
    })
    ro.observe(terminalRef.value!)
  } catch (e) {
    console.error('终端初始化失败', e)
    ElMessage.error('终端初始化失败')
  }
}

/* ---------- 事件监听 ---------- */
const setupEventListeners = () => {
  EventsOn(`terminal_output_${sessionId.value}`, (data: string) => {
    terminal?.write(data)
  })
  //window.addEventListener('resize', () => fitAddon?.fit())
}

/* ---------- 连接服务器 ---------- */
const connectServer = async () => {
  if (isConnected.value) return
  if (!server.serverUrl || !server.serverUserName || !server.serverPassword) {
    ElMessage.warning('请填写完整的服务器信息')
    return
  }

  conDis.value = true
  try {
    await nextTick(() => fitAddon!.fit())
    const { rows, cols } = terminal!
    const resp = await window.go.main.App.NewCon(
        JSON.stringify(server),
        sessionId.value,
        rows,
        cols
    )
    if (resp === 'success') {
      isConnected.value = true
      ElMessage.success('连接成功')
      terminal!.write('\r\n\x1B[32m✓ 连接成功！\x1B[0m\r\n\r\n')
    } else {
      ElMessage.error(resp)
    }
  } catch (e) {
    ElMessage.error('连接失败: ' + String(e))
  } finally {
    conDis.value = false
  }
}

/* ---------- 断开 / 清理 ---------- */
const closeCon = async () => {
  try {
    await window.go.main.App.CloseCon(sessionId.value)
    isConnected.value = false
    terminal?.write('\r\n\x1B[33m连接已断开\x1B[0m\r\n\r\n')
  } catch {
    /* ignore */
  }
}

const cleanup = () => {
  EventsOff(`terminal_output_${sessionId.value}`)
  window.removeEventListener('resize', () => fitAddon?.fit())
  fitAddon?.dispose()
  terminal?.dispose()
  if (isConnected.value) closeCon()
}
</script>

<template>
  <div class="terminal-wrapper">
    <div ref="terminalRef" class="demo-1"></div>
  </div>
</template>

<style scoped>
/* 父元素：占满可用空间，可随窗口自由缩放 */
.terminal-wrapper {
  flex: 1 1 auto;          /* 弹性布局关键：可伸可缩 */
  min-width: 0;
  min-height: 0;
  height: calc(100vh - 127px);
  padding: 0 0 0 10px;         /* 左右边距 */
  box-sizing: border-box;
}

/* 终端：永远等于父级尺寸，不再自增 */
.demo-1 {
  width: 100%;
  height: 100%;
}

/* 2. xterm 本身设置为块级且封顶 */
.demo-1 :deep(.xterm) {
  display: block;
  width: 100% !important;
  max-width: 100%;       /* 关键：封顶 */
}

/* 3. xterm 内部渲染层强制断行，防止撑开 */
.demo-1 :deep(.xterm-screen) {
  word-break: break-all;
  white-space: pre-wrap;
}

/* 4. 视口只负责垂直滚动，不再横向撑开 */
.demo-1 :deep(.xterm-viewport) {
  overflow-x: hidden !important;
  overflow-y: auto !important;
}
</style>