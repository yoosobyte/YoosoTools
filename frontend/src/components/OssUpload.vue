<script setup lang="ts">
import { ref } from 'vue'
import { UploadFilled } from '@element-plus/icons-vue'
import {reactive} from "vue";
import {ElMessage} from "element-plus";
const fileList = ref([]);
const uploadLoading = ref(false);
const form = reactive({
  "willUploadFile": null,
  "rootFileName": null,
  "nameType": "文件名+时分秒+随机数",
  "fileName": "",
  "fileUrl": "等待上传中..."
})

const handleChange = async (file, files) => {
  // 读取文件内容
  form.willUploadFile = await readFileAsArrayBuffer(file.raw)
  form.rootFileName = file.raw.name // 直接获取文件名
}

// 读取文件为ArrayBuffer
function readFileAsArrayBuffer(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = reject
    reader.readAsArrayBuffer(file) // 确保传入的是 Blob 类型
  })
}

async function getSplitFileName() {
  let originalFilename = form.rootFileName;
  // 1. 分离文件名和扩展名
  const lastDotIndex = originalFilename.lastIndexOf('.');
  let nameWithoutExt, extension;

  if (lastDotIndex === -1) {
    // 没有扩展名的情况
    nameWithoutExt = originalFilename;
    extension = '';
  } else {
    nameWithoutExt = originalFilename.substring(0, lastDotIndex);
    extension = originalFilename.substring(lastDotIndex); // 包含点，如 ".txt"
  }
  return [nameWithoutExt,extension]
}

const uploadFile = async () => {
  if(form.rootFileName === null || form.willUploadFile === null){
    ElMessage.info('请拖拽文件至上传区或点击选择文件后再执行上传操作')
    return
  }
  uploadLoading.value = true;
  let fileNameDetail = await getSplitFileName()
  switch (form.nameType){
    case "文件名+时分秒+随机数":
      const now = new Date();
      const timeString = [
        now.getFullYear(),
        String(now.getMonth() + 1).padStart(2, '0'),
        String(now.getDate()).padStart(2, '0'),
        String(now.getHours()).padStart(2, '0'),
        String(now.getMinutes()).padStart(2, '0'),
        String(now.getSeconds()).padStart(2, '0')
      ].join('');
      const randomNum = Math.floor(Math.random() * 10000).toString().padStart(4, '0');
      form.fileName = `${fileNameDetail[0]}_${timeString}_${randomNum}${fileNameDetail[1]}`;
      break;
    case "UUID":
      let uuid1 = crypto.randomUUID();
      form.fileName = `${uuid1}${fileNameDetail[1]}`;
      break;
    case "文件名+UUID":
      let uuid2 = crypto.randomUUID();
      form.fileName = `${fileNameDetail[0]}_${uuid2}${fileNameDetail[1]}`;
      break;
    case "原文件名(可能覆盖)":
      form.fileName = form.rootFileName;
      break;
    default:
      ElMessage.info('请选择文件名命名方式')
      uploadLoading.value = false;
      return
  }
  // 调用Go后端保存文件
  window.go.main.App.SaveFile({
    name: form.fileName,
    data: Array.from(new Uint8Array(form.willUploadFile)) // 转换为普通数组
  }).then((result) => {
    if (result.startsWith('http')) {
      form.fileUrl = result;
      ElMessage.success('上传成功')
    }else{
      ElMessage.warning(result);
    }
    fileList.value = []
    form.willUploadFile = null
    form.rootFileName = null
    uploadLoading.value = false;
  }).catch((err) => {
    fileList.value = []
    form.willUploadFile = null
    form.rootFileName = null
    form.fileUrl = err;
    ElMessage.info('上传失败')
    uploadLoading.value = false;
  })
}

async function copyRemote(url): Promise<void> {
  try {
    if(url && url.startsWith('http')) {
      await navigator.clipboard.writeText(url);
      ElMessage.info('文件URL已复制\n'+url);
    }else{
      ElMessage.info('请先上传')
      return;
    }
  } catch (err) {
    ElMessage.info('复制失败')
  }
}
</script>
<template>
  <div style="width: 100%;display: flex;align-items: center;justify-content: center;margin-top: 20px;">
    <div>
      <el-form
          ref="formRef"
          :model="form"
          label-width="auto"
          label-position="top"
          size="default"
          style="width: 800px;"
      >
        <el-form-item label="">
          <h2 style="text-align: center;width: 100%;margin-top: 50px;color: #5c5b5b;cursor: pointer;user-select: none;">
            上传文件至OSS仓库
          </h2>
          <div style="width: 100%;text-align: center;">
            <el-upload
                class="upload-demo"
                drag
                :auto-upload="false"
                :on-change="handleChange"
                :limit="1"
                :file-list="fileList"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                一、拖拽文件至此处 或 <em>点击选择文件</em>
              </div>
              <template #tip>
                <div class="el-upload__tip" style="text-align: left">
                  <span style="color: red;">
                    *
                  </span>
                  单个文件大小不能超过 1 GB
                </div>
              </template>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item label="请选择URL文件命名方式">
          <div style="display: flex;flex-direction: column;">
            <div>
              <el-radio-group v-model="form.nameType">
                <el-radio-button label="文件名+时分秒+随机数" />
                <el-radio-button label="UUID" />
                <el-radio-button label="文件名+UUID" />
                <el-radio-button label="原文件名(可能覆盖)" />
              </el-radio-group>
            </div>
            <div>
              <el-tag :disable-transitions="false" type="info" v-if="form.nameType==='文件名+时分秒+随机数'">示例: example_20350102150405_1425.txt</el-tag>
              <el-tag :disable-transitions="false" type="info" v-if="form.nameType==='UUID'">示例: 1e87787a-7ca3-11f0-bd6a-5b2b1d13485c.txt</el-tag>
              <el-tag :disable-transitions="false" type="info" v-if="form.nameType==='文件名+UUID'">示例: example_1e87787a-7ca3-11f0-bd6a-5b2b1d13485c.txt</el-tag>
              <el-tag :disable-transitions="false" type="info" v-if="form.nameType==='原文件名(可能覆盖)'">示例: example.txt</el-tag>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="上传结果">
          <el-tag type="primary">{{form.fileUrl}}
           <el-button v-if="form.fileUrl.startsWith('http')" size="small" type="text" @click="copyRemote(form.fileUrl)">复制</el-button>
          </el-tag>
        </el-form-item>
        <el-form-item>
          <div style="display: flex;align-items: center;width: 100%;">
            <el-button :loading="uploadLoading" style="width: 40%;" size="large" type="info" @click="uploadFile" plain>二、上传文件</el-button>
            <el-button :loading="uploadLoading" style="width: 60%;" size="large" type="primary" @click="copyRemote(form.fileUrl)" plain>三、复制上传文件后URL</el-button>
          </div>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>