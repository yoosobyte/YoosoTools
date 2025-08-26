<script lang="ts" setup>
import {ref} from 'vue'
import BuildCloud from './BuildCloud.vue'
import KillPort from "./KillPort.vue";
import MyIP from "./MyIP.vue";
import JsonParse from "./JsonParse.vue";
import OssUpload from "./OssUpload.vue";
import Linux from "./Linux.vue";
import More from "./More.vue";

const activeIndex = ref('3')

const handleSelect = (key: string, keyPath: string[]) => {
  activeIndex.value = key
}
</script>
<template>
  <div>
    <el-menu
        :default-active="activeIndex"
        mode="horizontal"
        @select="handleSelect"
    >
      <el-menu-item index="3">IP地址</el-menu-item>
      <el-menu-item index="2">Kill端口</el-menu-item>
      <el-menu-item index="4">JSON格式化</el-menu-item>
      <el-menu-item index="5">OSS上传</el-menu-item>
      <el-menu-item index="6">SSH/SFTP</el-menu-item>
      <el-sub-menu index="1">
        <template #title>一键部署</template>
        <el-menu-item index="1-1">your-project</el-menu-item>
      </el-sub-menu>
      <el-menu-item index="999">More...</el-menu-item>
    </el-menu>
    <div>
      <BuildCloud v-show="activeIndex=='1-1'" />
      <KillPort v-show="activeIndex=='2'" />
      <MyIP v-show="activeIndex=='3'" />
      <JsonParse v-show="activeIndex=='4'" />
      <OssUpload v-show="activeIndex=='5'" />
      <Linux v-show="activeIndex=='6'" />
      <More v-show="activeIndex=='999'" />
    </div>
  </div>
</template>

<style scoped>

</style>