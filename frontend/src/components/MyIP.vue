<script lang="ts" setup>
import {reactive, ref} from 'vue'
import {CopyDocument, DocumentCopy, Guide, MagicStick, Refresh, StarFilled} from '@element-plus/icons-vue'
import {ElMessage} from 'element-plus'
import {EventsOn} from "../../wailsjs/runtime";

const form = reactive({
  "ret": "ok",
  "err": "",
  "data": {
    "ip": "x.x.x.x",
    "localIp": "x.x.x.x",
    "location": ["国家", "省份", "市区", "", "运营商"],
  },
})
const ipList = ref([])
const reqLoading = ref(false)

async function copyIp(data): Promise<void> {
  try {
    await navigator.clipboard.writeText(data);
    ElMessage({
      message: data + ' 已复制',
      grouping: true,
      type: 'success',
    })
  } catch (err) {
    ElMessage.info('复制失败')
  }
}

getInitData('0');

function getInitData (code) {
  reqLoading.value = true
  setTimeout(() => {
    window.go.main.App.GetIpInfo().then((resp)=>{
      const newData = JSON.parse(resp);
      Object.assign(form,newData)
      reqLoading.value = false
      if(code==='1'){
        ElMessage({
          message: '已重新获取IP',
          grouping: true,
          type: 'info',
        })
      }
    }).catch(err=>{
      reqLoading.value = false
      if(code==='1'){
        ElMessage.warning('获取URL请求异常,请刷新')
      }
      getInitData('0');
    })
  },100)
}

function getAddr(type){
  let addrList = form.data.location
  if(type==='addr'){
    return addrList[0] +'-'+ addrList[1] +'-'+ addrList[2];
  }
  if(type==='agent'){
    return addrList[4];
  }
}

getRadioIpList()

function getRadioIpList () {
  window.go.main.App.GetRadioIpList().then((resp)=>{
    const ipListData = JSON.parse(resp);
    ipListData.data.sort((a, b) => {
      return a.localIp.localeCompare(b.localIp);
    });
    ipList.value = ipListData.data;
  }).catch(err=>{
  })
}

EventsOn(`new_peer_data`, async () => {
  getRadioIpList();
})

function postRadio (){
  window.go.main.App.PostRadio().then((resp)=>{
    ElMessage({
      message: '已重新发送广播,快告诉你的小伙伴吧',
      grouping: true,
      type: 'info',
    })
  }).catch(err=>{
  })
}
</script>
<template>
  <div style="width: 100%;display: flex;align-items: center;justify-content: center;">
    <div>
      <el-form
          ref="formRef"
          :model="form"
          label-width="auto"
          label-position="top"
          size="default"
          style="width: 500px;"
      >
        <el-form-item label="" prop="killPort">
          <h2 style="text-align: center;width: 100%;margin-top: 90px;color: #5c5b5b;cursor: pointer;user-select: none;padding-left: 16px;" @click="getInitData('1')">
            您的IP地址
            <el-button style="padding: 4px 1px 7px 1px" :icon="Refresh" text :loading="reqLoading"/>
          </h2>
          <el-descriptions
              label-width="80"
              title=""
              direction="horizontal"
              :column="1"
              size="large"
              border
              style="width: 100%;"
          >
            <el-descriptions-item label="远程IP">{{form.data.ip}}</el-descriptions-item>
            <el-descriptions-item label="本地IP">{{form.data.localIp}}</el-descriptions-item>
            <el-descriptions-item label="地区" >{{getAddr('addr')}}</el-descriptions-item>
            <el-descriptions-item label="运营商" >{{getAddr('agent')}}</el-descriptions-item>
            <el-descriptions-item label="IP组" v-if="ipList.length > 0">
              <div
                  v-for="(item, index) in ipList"
                  :key="index"
                  style="margin: 3px 0;"
              >
                <div style="display: flex;align-items: center;width: 100%;height: 28px;background-color: #F4F4F5;border-radius: 3px;border: 1px solid #E9E9EB;">
                  <div style="text-align: start;width: 30%;padding-left: 10px;">{{item.localIp}}</div>
                  <div style="text-align: start;width: 10%;">
                    <el-button type="info" text size="small" @click="copyIp(item.localIp)">复制</el-button>
                  </div>
                  <div class="new-space" style="width: 20%;text-align: center;display:flex;justify-content: center;">
                    <el-divider style="width: 55px;" class="new-divider">
                      <el-icon style="color:#73767A;background-color: #f4f4f5;"><star-filled /></el-icon>
                    </el-divider>
                  </div>
                  <div style="text-align: start;color:#409EFF;width: 30%;">{{item.remoteIp}}</div>
                  <div style="text-align: start;width: 10%;padding-right: 10px;">
                    <el-button type="primary" text size="small" @click="copyIp(item.remoteIp)">复制</el-button>
                  </div>
                </div>
              </div>
            </el-descriptions-item>
          </el-descriptions>
          <div v-if="form.err!==''">
            <el-tag type="danger" size="large">{{form.err}}</el-tag>
          </div>
        </el-form-item>
        <el-form-item>
          <div style="display: flex;align-items: center;width: 100%;">
            <el-button style="width: 15%;" size="large" type="info" :icon="MagicStick" @click="postRadio" plain>广播</el-button>
            <el-button style="width: 35%;" size="large" type="info" :icon="DocumentCopy" @click="copyIp(form.data.localIp)" plain>复制本地IP</el-button>
            <el-button style="width: 50%;" size="large" type="primary" :icon="CopyDocument" @click="copyIp(form.data.ip)" plain>复制远程IP->剪贴板</el-button>
          </div>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>
<style scoped>
.new-space :deep(.el-divider__text) {
  padding: 0 0;
}
.new-divider :deep(.el-divider__text) {
  background-color: #f4f4f5;
}
</style>