<script setup lang="ts">

import {Files, Setting} from "@element-plus/icons-vue";
import SSHTerminal from "./SSHTerminal.vue";
import SFTPTerminal from "./SFTPTerminal.vue";
import {ref} from "vue";
import Terminal from "./icons/Terminal.vue";
import {Server} from "../types/server";
const handleSelect = (key: string, keyPath: string[]) => {
  activeIndex.value = key
}
const activeIndex = ref('1')

const props = defineProps<{ server: Server }>()

</script>

<template>
  <div class="container">
    <el-menu
        :default-active="activeIndex"
        class="el-menu-vertical-demo"
        :collapse="false"
        @select="handleSelect"
        lazy
    >
      <el-menu-item index="1">
        <el-icon><Terminal /></el-icon>

        <template #title>终端</template>
      </el-menu-item>
      <el-menu-item index="2">
        <el-icon><Files /></el-icon>
        <template #title>目录</template>
      </el-menu-item>
    </el-menu>

    <div class="content">
      <SSHTerminal :server="props.server"  v-show="activeIndex === '1'" />
      <SFTPTerminal :server="props.server"  v-show="activeIndex === '2'" />
    </div>
  </div>
</template>

<style scoped>
.container {
  display: flex; /* 启用 Flex 布局 */
  margin-top: 0
}
.el-menu-vertical-demo {
  flex-shrink: 0; /* 防止菜单被压缩 */
  height: calc(100vh - 127px);
  width: 83px;
}
.el-menu-vertical-demo :deep(.el-menu-tooltip__trigger) {
  align-items: center;
  box-sizing: border-box;
  display: inline-flex;
  height: 100%;
  left: 0;
  padding: 0 var(--el-menu-base-level-padding);
  position: absolute;
  top: 0;
  width: 100%;
  justify-content: center;
}
.el-menu-vertical-demo :deep(.el-menu-item) {
  justify-content: center;
}
.content {
  flex-grow: 1; /* 内容区域填充剩余空间 */
  padding: 0 0 0 0; /* 可选：添加内边距 */
}
</style>