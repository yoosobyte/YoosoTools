<script setup lang="ts">
import {CircleClose, CirclePlus, Edit} from "@element-plus/icons-vue";
import {reactive, ref} from "vue";
import {ElMessage, ElMessageBox} from "element-plus";

const open = ref(false)
const serverList = ref<Server[]>([])
const serverForm = reactive<Server>({...NewServer})
const formRef = ref(null)

function newServer() {
  closeServer();
  open.value = true;
}

function editServer(serverId) {
  window.go.main.App.GetOneServer(serverId).then((resp) => {
    const result = JSON.parse(resp);
    if (result.code === 200) {
      open.value = !open.value;
      Object.assign(serverForm,result.data);
    }
  })
}

const rules = {
  serverName: [{required: true, message: '请输入服务器名称', trigger: 'blur'}],
  serverUrl: [{required: true, message: '请输入服务器IP', trigger: 'blur'}],
  serverPort: [{required: true, message: '请输入服务器端口', trigger: 'blur'}],
  serverUserName: [{required: true, message: '请输入服务器账号', trigger: 'blur'}],
  serverPassword: [{required: true, message: '请输入服务器密码', trigger: 'blur'}],
}
const closeServer = async () => {
  open.value = false;
  Object.assign(serverForm, NewServer);
}

const saveServer = async () => {
  const valid = await formRef.value.validate()
  if (valid) {
    let resultStr = '';
    console.log(serverForm.serverId)
    if (serverForm.serverId === undefined || serverForm.serverId === null){
      resultStr = await window.go.main.App.SaveServer(JSON.stringify(serverForm))
    }else{
      resultStr = await window.go.main.App.EditServer(JSON.stringify(serverForm))
    }
    const result = JSON.parse(resultStr)
    if (result.code === 200) {
      getServerList();
      Object.assign(serverForm, NewServer);
      ElMessage.success(result.msg)
      await closeServer();
    }
  }
}

getServerList();

function getServerList() {
  window.go.main.App.GetListServer().then((resp) => {
    const result = JSON.parse(resp);
    if (result.code === 200) {
      serverList.value = result.data;
      closeServer();
    }
  })
}

function removeServer(serverId) {
  ElMessageBox.confirm('确认删除此项?','提示',{
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info',
  }).then(() => {
    window.go.main.App.RemoveServer(serverId).then((resp) => {
      const result = JSON.parse(resp);
      if (result.code === 200) {
        serverList.value = result.data;
        getServerList();
      }
    })
  }).catch(() => {
  })
}

import {emitter} from '../call/event-bus'
import {NewServer, Server} from "../types/server";

function choseServer(item){
  emitter.emit('ServerList.vue.NewTab', item as Server);
}
</script>
<template>
  <div class="container">
    <div class="custom-grid">
      <el-card shadow="hover" class="al-center" style="cursor: pointer;" @click="newServer">
        <div class="al-center" style="flex-direction: column;height: 121px;">
          <el-icon style="color: #969696;" size="50">
            <CirclePlus/>
          </el-icon>
          <div style="font-family: 幼圆;font-size: 16px;font-weight: bold;color: #969696;user-select: none;">
            新建连接
          </div>
        </div>
      </el-card>
      <div v-for="(item, index) in serverList" :key="index">
        <el-card shadow="hover" class="new-card2" @dblclick="choseServer(item)">
          <el-descriptions
              class="new-border"
              :column="2"
              border
          >
            <el-descriptions-item :span="2">
              <template #label>
                <div class="cell-item">
                  名称
                </div>
              </template>
              <div class="al-center" style="justify-content: space-between;">
                <div>{{ item.serverName }}</div>
                <div>
                  <el-icon style="color: #969696;cursor: pointer;margin-right: 10px;" @click.stop="editServer(item.serverId)">
                    <Edit/>
                  </el-icon>
                  <el-icon style="color: #3b3a3a;cursor: pointer;" @click.stop="removeServer(item.serverId)">
                    <CircleClose/>
                  </el-icon>
                </div>
              </div>
            </el-descriptions-item>
            <el-descriptions-item :span="2">
              <template #label>
                <div class="cell-item">主机</div>
              </template>
              {{ item.serverUrl }}
            </el-descriptions-item>
            <el-descriptions-item :span="2">
              <template #label>
                <div class="cell-item">端口</div>
              </template>
              {{ item.serverPort }}
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">账号</div>
              </template>
              {{ item.serverUserName }}
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">密码</div>
              </template>
              ******
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </div>
    </div>

    <el-dialog v-model="open" title="编辑服务器配置" width="40%" draggable :destroy-on-close="true" :close-on-click-modal="false" :close-on-press-escape="false">
      <el-form
          ref="formRef"
          :model="serverForm"
          label-width="auto"
          label-position="top"
          size="default"
          :rules="rules"
      >
        <el-form-item label="名称" prop="serverName">
          <el-input v-model="serverForm.serverName" placeholder="请输入服务器名称"/>
        </el-form-item>
        <el-form-item label="主机" prop="serverUrl">
          <el-input v-model="serverForm.serverUrl" placeholder="请输入服务器IP"/>
        </el-form-item>
        <el-form-item label="端口" prop="serverPort">
          <el-input v-model="serverForm.serverPort" placeholder="请输入服务器端口"/>
        </el-form-item>
        <el-form-item label="账号" prop="serverUserName">
          <el-input v-model="serverForm.serverUserName" placeholder="请输入服务器账号"/>
        </el-form-item>
        <el-form-item label="密码" prop="serverPassword">
          <el-input type="password" v-model="serverForm.serverPassword" placeholder="请输入服务器密码"/>
        </el-form-item>
      </el-form>
      <template #footer>
      <span class="dialog-footer">
        <el-button @click="closeServer">关 闭</el-button>
        <el-button type="primary" @click="saveServer">
          保 存
        </el-button>
      </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.container {
  height: calc(100vh - 147px);
  padding: 10px;
  overflow: auto;
}

.custom-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr); /* 4列 */
  gap: 10px; /* 统一间距 */
}

.new-card2 {
  user-select: none;
}
.new-card2 :deep(.el-card__body) {
  padding: 0;

}

.new-border {
  padding: 0 !important;
  margin: 0 !important;
}

.al-center {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>