<script lang="ts" setup>
import {onUnmounted, ref} from 'vue'
import {HomeFilled} from '@element-plus/icons-vue'
import ServerList from "./ServerList.vue";
import ServerItem from "./ServerItem.vue";
import {emitter} from "../call/event-bus";
import {Server} from "../types/server";

const activeTab = ref('主页')
const tabList = ref<Server[]>([])
const serverNo: Map<number, number> = new Map();
emitter.on('ServerList.vue.NewTab', (newTab: Server) => {
  // 创建对象的深拷贝
  const tabCopy = { ...newTab };

  const sameServerTabs = tabList.value.filter(
      oldTab => oldTab.serverId === newTab.serverId
  );
  if(sameServerTabs.length > 0){
    let count = serverNo.get(tabCopy.serverId);
    if(count === undefined){
      serverNo.set(tabCopy.serverId,1)
      tabCopy.serverNickName = `${tabCopy.serverName} (${serverNo.get(tabCopy.serverId)})`;
    }else{
      serverNo.set(tabCopy.serverId,count + 1)
      tabCopy.serverNickName = `${tabCopy.serverName} (${serverNo.get(tabCopy.serverId)})`;
    }
  }else{
    tabCopy.serverNickName = `${tabCopy.serverName}`;
    serverNo.delete(tabCopy.serverId)
  }

  //修改Tab
  tabCopy.sessionId = crypto.randomUUID()
  //Push数据
  tabList.value.push(tabCopy);
  //切Tab
  activeTab.value = tabCopy.sessionId;
});

onUnmounted(() => {
  emitter.off('ServerList.vue.NewTab')
})

const handleTabRemove = (sessionId) => {
  const index = tabList.value.findIndex(item => item.sessionId === sessionId)
  if (index !== -1) {
    tabList.value.splice(index, 1)
  }
  if (tabList.value.length == 0) {
    activeTab.value = '主页'
  }
  if (tabList.value.length > 0) {
    activeTab.value = tabList.value[0].sessionId
  }
}
</script>
<template>
  <el-tabs
      type="border-card"
      class="my-linux-tab"
      style="margin-top: 10px;"
      v-model="activeTab"
      @tab-remove="handleTabRemove"
      lazy
  >
    <el-tab-pane name="主页" label="主页">
      <template #label>
        <div class="row-center">
          <el-icon><HomeFilled /></el-icon>
          <div slot="label">主页</div>
        </div>
      </template>
      <ServerList/>
    </el-tab-pane>

    <el-tab-pane v-for="(item,index) in tabList" :name="item.sessionId" :closable="true">
      <template #label>{{item.serverNickName}}</template>
      <ServerItem :server="item"/>
    </el-tab-pane>

  </el-tabs>
</template>
<style scoped>
.row-center {
  display: flex;
  align-items: center;
  justify-content: space-around;
}
.my-linux-tab :deep(.el-tabs__content) {
  padding: 0;
}
</style>