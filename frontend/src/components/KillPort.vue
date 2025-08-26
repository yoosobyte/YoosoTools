<script lang="ts" setup>
import {reactive, ref} from 'vue'
import {Close} from '@element-plus/icons-vue'
import {ElMessage} from 'element-plus'

const response = ref('')
const setForm = reactive({
  killPort: '',
})
const respArray = ref([])

const formRef = ref(null)

const saveSet = async () => {
  try {
    const valid = await formRef.value.validate()
    if (valid) {
      response.value = await window.go.main.App.KillPort(setForm.killPort)
      const result = response.value

      // 1. 将新结果添加到数组开头（新在上，旧在下）
      respArray.value.unshift(result)

      // 2. 限制数组最大长度为10
      if (respArray.value.length > 9) {
        respArray.value = respArray.value.slice(0, 9)
      }
    }
  } catch (error) {
    console.error('操作失败:', error)
  }
}

const rules = {
  killPort: [{ required: true, message: '请输入你想Kill的端口号', trigger: 'blur' }],
}

const handleClear = () => {
  ElMessage.info('端口号已清除')
  setForm.killPort = ''
  respArray.value = []
}

const clearHistory = () => {
  respArray.value = []
}

function newType(status){
  if (status.endsWith(';')){
    return "danger";
  }else{
    return "primary";
  }
}

function deleteItem(index){
  respArray.value.splice(index,1)
}

const handleKeyDown = (e) => {
  if (e.key === 'Enter') {
    e.preventDefault() // 阻止默认行为
    saveSet() // 显式调用你的方法
  }
}
</script>
<template>
  <div style="width: 100%;display: flex;align-items: center;justify-content: center;margin-top: 20px;">
    <div>
      <el-form
          ref="formRef"
          :rules="rules"
          :model="setForm"
          label-width="auto"
          label-position="top"
          size="default"
      >
        <el-form-item label="" prop="killPort">
          <h2 style="text-align: center;width: 100%;margin-top: 100px;color: #5c5b5b;">
            请输入您要kill的端口号
          </h2>
          <el-input v-model="setForm.killPort" size="large" placeholder="请输入您要kill的端口号"  @keydown="handleKeyDown">
            <template #append>
              <el-button :icon="Close" @click="handleClear()"/>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <div style="display: flex;align-items: center;width: 100%;">
            <el-button style="width: 100%;" size="large" type="primary" @click="saveSet">kill by port</el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <div class="item-container">
            <el-tag class="item" v-for="(item, index) in respArray" :type="newType(item)" size="large" :key="index" :disable-transitions="false" closable @close="deleteItem(index)">
              {{item}}
            </el-tag>
          </div>
          <el-button
              type="info"
              text
              bg
              v-if="respArray.length > 0"
              style="width: 100%;margin-top: 3px;"
              @click="clearHistory"
          >清空历史记录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>
<style scoped>
.item-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.item {
  /* 确保每个项目占满一行 */
  width: 100%;
  /* 可选样式 */
  padding: 8px;
  margin-bottom: 4px;
}
</style>