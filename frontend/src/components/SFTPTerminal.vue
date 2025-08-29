<script setup lang="ts">
import {
  Document, DocumentAdd, DocumentChecked,FolderAdd,
  FolderOpened, More, Promotion, Refresh,
  RefreshLeft,  Upload
} from "@element-plus/icons-vue";
import {onMounted, reactive, ref} from "vue";
import {Server} from "../types/server";
import {PathItem} from "../types/path-item";
import { ElMessage, ElMessageBox, ElNotification} from "element-plus";
import {EventsOn} from "../../wailsjs/runtime";
import {AddFile, NewAddFile} from "../types/sftp";
const props = defineProps<{ server: Server }>()
const sessionId = ref<string>(props.server.sessionId)
const findIng = ref(false)
const followSSHDir = ref(true)
const nowDir = ref('/');
const nowSSHDir = ref('');
const copyDir = ref('');
const cutDir = ref('');
const nowFileList = ref<PathItem[]>([])
const newFileForm = reactive<AddFile>({...NewAddFile})
const newFileFormOpen = ref(false)
const formRef = ref(null)
const uploadRate = ref(0)
const uploadMsg = ref('')
const downloadRate = ref(0)

if(!followSSHDir.value){
  postNewPath(1);
}

function postNewPath(isInit){
  findIng.value = true;
  window.go.main.App.PostNewPath(nowDir.value,sessionId.value,isInit).then((data) => {
    const result = JSON.parse(data);
    if(result.code === 200){
      nowFileList.value = result.data;
    }else{
      ElMessage.error('获取文件信息失败')
    }
    findIng.value = false;
  })
}

function changeNowDir(dir){
  nowDir.value = dir;
  postNewPath(0);
}

EventsOn(`ssh_dir_${sessionId.value}`, async (sshDir: string) => {
  nowSSHDir.value = sshDir
  if(followSSHDir.value){
    changeNowDir(sshDir)
  }
})

EventsOn(`upload_rate_call_${sessionId.value}`, async (rateAndMsg: string) => {
  if(!rateAndMsg){
    return
  }
  let rateArray = rateAndMsg.split('|');
  uploadRate.value = parseInt(rateArray[0])
  uploadMsg.value = rateArray[1]
})

EventsOn(`download_rate_call_${sessionId.value}`, async (rateAndMsg: string) => {
  if(!rateAndMsg){
    return
  }
  let rateArray = rateAndMsg.split('|');
  downloadRate.value = parseInt(rateArray[0])
})

function changeFollowSwitch(){
  followSSHDir.value = !followSSHDir.value
  if(followSSHDir.value){
    changeNowDir(nowSSHDir.value)
  }
}

function refreshDir(){
  ElMessage.info('目录已刷新')
  copyDir.value = ''
  cutDir.value = ''
  postNewPath(1);
}
function openDir(dir){
  if(dir.isFolder === 1){
    changeNowDir(dir.fullPath)
  }else{
    ElMessageBox.confirm(`是否下载文件${dir.fileName},大小:${dir.fileSize}?`,'提示',{
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info',
    }).then(() => {
      downloadFile(dir.fullPath)
    }).catch(() => {
    })
  }
}

const downloadFile = async (filePath) => {
  let notification  = ElNotification({
    title: '下载中',
    message: "请稍候,下载成功后将弹出提示",
    position: 'bottom-right',
    type:'info'
  })
  try {
    // 1. 先获取文件信息
    const fileInfo = await window.go.main.App.GetFileSize(sessionId.value, filePath)
    const result = JSON.parse(fileInfo)

    if (result.code !== 200) {
      throw new Error('获取文件信息失败')
    }

    // 2. 下载文件内容
    const downloadResult = await window.go.main.App.DownloadFile(sessionId.value, filePath)
    const downloadData = JSON.parse(downloadResult)

    if (downloadData.code !== 200) {
      throw new Error(downloadData.msg || '下载失败')
    }

    // 3. Base64 解码并创建下载链接
    const { fileName, content } = downloadData.data
    const binaryString = atob(content)
    const bytes = new Uint8Array(binaryString.length)

    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }

    const blob = new Blob([bytes], { type: 'application/octet-stream' })
    const url = URL.createObjectURL(blob)

    // 4. 创建下载链接并触发下载
    const link = document.createElement('a')
    link.href = url
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    ElMessage.success('下载成功')

  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error(error.message || '下载失败')
  } finally {
    notification.close();
  }
}
// 返回上一级目录
const goToParent = () => {
  if (nowDir.value === '/') return
  // 去除末尾的斜杠
  const cleanPath = nowDir.value.replace(/\/+$/, '')
  // 分割路径并去除空部分
  const parts = cleanPath.split('/').filter(part => part !== '')
  // 构建父路径
  const parentPath = parts.length > 1
      ? '/' + parts.slice(0, -1).join('/')
      : '/'
  changeNowDir(parentPath)
}

function clickDownload(dir){
  if(dir.isFolder === 1){
    ElMessage.info('下载文件夹建设中,敬请期待')
  }else{
    ElMessageBox.confirm(`是否下载文件${dir.fileName},大小:${dir.fileSize}?`,'提示',{
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info',
    }).then(() => {
      downloadFile(dir.fullPath)
    }).catch(() => {
    })
  }
}

function newFolder(){
  ElMessageBox.prompt('请输入文件夹名', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[\u4e00-\u9fa5a-zA-Z0-9._-]+$/,
    inputErrorMessage: '只允许中文、英文、数字、英文点、下划线和连字符',
    closeOnClickModal: false,
    closeOnPressEscape: false,
  }).then(({ value }) => {
    window.go.main.App.NewFolder(nowDir.value,value,sessionId.value).then((data) => {
      const result = JSON.parse(data);
      if(result.code === 200){
        changeNowDir(result.data);
        ElMessage.success('文件夹已创建')
      }else{
        ElMessage.error(result.msg)
      }
    })
  }).catch(() => {
  });
}

function copyFolderOrFile(dir){
  copyDir.value = dir.fullPath
  ElMessage.success('文件/夹已复制,请转至要粘贴的位置,并点击[粘贴到此目录(复制)]')
}
function cutFolderOrFile(dir){
  cutDir.value = dir.fullPath
  ElMessage.success('文件/夹已剪切,请转至要粘贴的位置,并点击[粘贴到此目录(剪切)]')
}
function renameFolderOrFile(dir){
  const oldName = dir.fileName || '';
  ElMessageBox.prompt('', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[\u4e00-\u9fa5a-zA-Z0-9._-]+$/,
    inputErrorMessage: '只允许中文、英文、数字、英文点、下划线和连字符',
    closeOnClickModal: false,
    closeOnPressEscape: false,
    dangerouslyUseHTMLString: true,
    inputValue: oldName,
    message: `
    <div style="line-height: 1.5;">
      <div>即将重命名: <strong>${dir.fileName}</strong></div>
      <div>请给文件/夹起个名字</div>
    </div>
  `
  }).then(({ value }) => {
    window.go.main.App.RenameFolderOrFile(dir.fullPath,value,sessionId.value).then((data) => {
      const result = JSON.parse(data);
      if(result.code === 200){
        changeNowDir(nowDir.value);
        ElMessage.success('文件/夹已重命名')
      }else{
        ElMessage.error(result.msg)
      }
    })
  }).catch(() => {
  });
}

function pasteCopy(){
  // 提取旧文件名
  const oldName = copyDir.value.split(/[/\\]/).pop() || '';

  ElMessageBox.prompt('', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[\u4e00-\u9fa5a-zA-Z0-9._-]+$/,
    inputErrorMessage: '只允许中文、英文、数字、英文点、下划线和连字符',
    closeOnClickModal: false,
    closeOnPressEscape: false,
    dangerouslyUseHTMLString: true,
    // 关键：把旧文件名作为默认值
    inputValue: oldName,
    message: `
    <div style="line-height: 1.5;">
      <div>您复制了: <strong>${copyDir.value}</strong></div>
      <div>将复制到目录: <strong>${nowDir.value}</strong></div>
      <div>请给文件/夹起个名字</div>
    </div>
  `
  }).then(({ value }) => {
    window.go.main.App.MoveOrCopyFolderAndFile(
        nowDir.value,
        copyDir.value,
        value,
        sessionId.value,
        1
    ).then((data) => {
      const result = JSON.parse(data);
      if (result.code === 200) {
        changeNowDir(nowDir.value);
        copyDir.value = '';
        ElMessage.success('文件/夹已粘贴');
      } else {
        ElMessage.error(result.msg);
      }
    });
  }).catch(() => {});
}
function pasteCut(){
  const oldName = cutDir.value.split(/[/\\]/).pop() || '';

  ElMessageBox.prompt('', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[\u4e00-\u9fa5a-zA-Z0-9._-]+$/,
    inputErrorMessage: '只允许中文、英文、数字、英文点、下划线和连字符',
    closeOnClickModal: false,
    closeOnPressEscape: false,
    inputValue: oldName,
    dangerouslyUseHTMLString: true,
    message: `
    <div style="line-height:1.6;">
      <div>您剪切了：<strong>${cutDir.value}</strong></div>
      <div>将剪切到目录：<strong>${nowDir.value}</strong></div>
      <div>请给文件/夹起个名字：</div>
    </div>
  `
  }).then(({ value }) => {
    window.go.main.App
        .MoveOrCopyFolderAndFile(
            nowDir.value,
            cutDir.value,
            value,
            sessionId.value,
            0
        )
        .then((data) => {
          const result = JSON.parse(data);
          if (result.code === 200) {
            changeNowDir(nowDir.value);
            cutDir.value = '';
            ElMessage.success('文件/夹已粘贴');
          } else {
            ElMessage.error(result.msg);
          }
        });
  }).catch(() => {});
}
function removeFolderOrFile(dir){
  ElMessageBox.confirm(`是否删除文件${dir.isFolder?'夹':''}:${dir.fileName},大小:${dir.fileSize}?`,'提示',{
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info',
  }).then(() => {
    window.go.main.App.DeleteItem(dir.fullPath,sessionId.value).then((data) => {
      const result = JSON.parse(data);
      if(result.code === 200){
        changeNowDir(nowDir.value);
        ElMessage.success(`文件${dir.isFolder?'夹':''}已删除`)
      }else{
        ElMessage.error(result.msg)
      }
    })
  }).catch(() => {
  })
}
const rules = {
  fileName: [{required: true, message: '请输入新文件文件名', trigger: 'blur'}],
  fileContent: [{required: true, message: '请输入新文件内容', trigger: 'blur'}],
}
function newFile(){
  newFileFormOpen.value = true;
}

const closeAddFile = async () => {
  newFileFormOpen.value = false;

  // 创建全新的 reactive 对象
  const newForm = reactive<AddFile>({
    fileName: '',
    fileContent: ''
  });

  // 如果需要替换原引用，可能需要重新赋值
  // 但注意：直接替换 reactive 的引用可能有问题
  Object.assign(newFileForm, newForm);
}

const saveNewFile = async () => {
  const valid = await formRef.value.validate()
  if (valid) {
    let resultStr = await window.go.main.App.SaveNewFile(newFileForm.fileName,newFileForm.fileContent,nowDir.value,sessionId.value)
    const result = JSON.parse(resultStr)
    if (result.code === 200) {
      changeNowDir(nowDir.value);
      ElMessage.success(result.msg)
      await closeAddFile();
    }
  }
}
const UploadDirOrFile = async () => {
  ElMessageBox.confirm(`请选择你要上传的文件类型?`,'提示',{
    confirmButtonText: '文件夹',
    cancelButtonText: '文件',
    distinguishCancelAndClose: true,
    type: 'info',
  }).then(() => {
    window.go.main.App.SelectFolder('您要上传的文件夹').then((path)=>{
      if (path) {
        window.go.main.App.UploadDirOrFile(nowDir.value,path,sessionId.value).then((resultStr)=>{
          const result = JSON.parse(resultStr)
          if (result.code === 200) {
            changeNowDir(nowDir.value);
            ElMessage.success('上传成功')
          }else{
            console.error(result.msg);
          }
        });
      }
    });
  }).catch((error) => {
    console.log(error);
    if (error === 'cancel') {
      window.go.main.App.SelectFile('您要上传的文件').then((path)=>{
        if (path) {
          window.go.main.App.UploadDirOrFile(nowDir.value,path,sessionId.value).then((resultStr)=>{
            const result = JSON.parse(resultStr)
            if (result.code === 200) {
              changeNowDir(nowDir.value);
              ElMessage.success('上传成功')
            }else{
              console.error(result.msg);
            }
          });
        }
      });
    }
  })
}
onMounted(() => {
  setupFileDropListener();
})
const setupFileDropListener = () => {
  //v1,v2API冲突,这个不能删
  window.runtime.OnFileDrop((x, y, paths) => {
    // ElMessage.success("文件拖放事件1:"+paths)
    // console.log('文件拖放事件:', paths, '位置:', x, y);
  }, true);
  window.runtime.EventsOn("wails:file-drop", (x, y, paths) => {
    ElMessage.success("开始上传,文件/夹:"+paths.length +"个")
    for (const [index, path] of paths.entries()) {
      window.go.main.App.UploadDirOrFile(nowDir.value,path,sessionId.value).then((resultStr)=>{
        const result = JSON.parse(resultStr)
        if (result.code === 200) {
          changeNowDir(nowDir.value);
        }else{
          console.error(result.msg);
        }
      });
    }
  });
};
</script>
<template>
  <div class="sftp-container">
    <el-input v-model="nowDir" readonly placeholder="在此输出路径"  :input-style="{fontSize:'20px',color:'#969696',fontFamily:'Times'}" size="large">
      <template #append>
        <el-button :icon="Promotion" @click="changeFollowSwitch">跟随终端目录:{{followSSHDir===true?'开':'关'}}</el-button>
      </template>
    </el-input>
    <div style="margin-top: 10px;display: flex;">
      <el-button plain type="" :icon="FolderAdd" @click="newFolder">新建文件夹</el-button>
      <el-button plain type="" :icon="DocumentAdd" @click="newFile">新建文件</el-button>
      <el-button plain type="" :icon="Refresh" @click="refreshDir">刷新</el-button>
      <el-button plain type="" :icon="Upload" @click="UploadDirOrFile">上传到此目录(支持拖动)</el-button>
      <el-button plain type="" @click="pasteCopy" v-if="copyDir!==''" :icon="DocumentChecked">粘贴到此目录(复制)</el-button>
      <el-button plain type="" @click="pasteCut" v-if="cutDir!==''" :icon="DocumentChecked">粘贴到此目录(剪切)</el-button>
      <div style="margin-left: 10px;" v-show="uploadRate > 0 && uploadRate < 100">
        <el-tag closable type="info" size="large">
          <div>
            上传进度 :&nbsp;&nbsp;&nbsp;{{uploadRate}}% | {{uploadMsg}}}
          </div>
        </el-tag>
      </div>
      <div style="margin-left: 10px;" v-show="downloadRate > 0 && downloadRate < 100">
        <el-tag type="info" size="large">
          <div>
            下载进度 :&nbsp;&nbsp;&nbsp;{{downloadRate}}%
          </div>
        </el-tag>
      </div>
    </div>
    <div class="custom-grid" style="margin-top: 10px;" v-loading="findIng">
      <el-card shadow="hover" class="a-center card-height" @dblclick="goToParent">
        <div class="a-center a-row">
          <el-icon class="unchose-color" size="40"><RefreshLeft /></el-icon>
          <div class="title-text y-font unchose-color" style="margin-top: 10px;">
            返回上一层
          </div>
        </div>
      </el-card>
      <div v-for="(item, index) in nowFileList" :key="index" style="position: relative;">
        <el-card shadow="hover" class="a-center card-height" :class="item.isFolder?'folderBg':'fileBg'" @dblclick="openDir(item)">
          <div>
            <div class="a-center a-row">
              <el-icon class="" size="50" v-if="item.isFolder === 1"><FolderOpened /></el-icon>
              <el-icon class="" size="50" v-if="item.isFolder === 0"><Document /></el-icon>
              <el-tooltip
                  ref="tooltipRef"
                  class="box-item"
                  transition=""
                  effect="dark"
                  :content="item.fileName"
                  placement="bottom"
                  :disabled="item.fileName.length <=13"
              >
                <div class="title-text ellipsis" style="font-weight: normal;font-size: 15px;">
                  {{item.fileName}}
                </div>
              </el-tooltip>
              <div class="title-text unchose-color" style="font-weight: normal;font-size: 13px;margin-top: 1px;">
                {{item.editTime}}
              </div>
            </div>
            <div style="position: absolute;bottom: -7px;right: 0;z-index: 10">
              <el-dropdown placement="bottom-start" trigger="click">
                <span class="el-dropdown-link">
                  <el-icon class="unchose-color" style="padding: 1px 8px;">
                    <More />
                  </el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="copyFolderOrFile(item)">复制</el-dropdown-item>
                    <el-dropdown-item @click="cutFolderOrFile(item)">剪切</el-dropdown-item>
                    <el-dropdown-item @click="renameFolderOrFile(item)">重命名</el-dropdown-item>
                    <el-dropdown-item @click="clickDownload(item)">下载</el-dropdown-item>
                    <el-dropdown-item :disabled="item.isDanger === 1" @click="removeFolderOrFile(item)">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <div class="unchose-color" style="position: absolute;bottom: 0;right: 0;left:0;font-size: 13px;user-select:none;background-color: #dddddd;text-align: center">
              {{item.fileSize}}
            </div>
          </div>
        </el-card>
      </div>
    </div>


    <el-dialog v-model="newFileFormOpen" title="新建文件" width="40%" draggable :destroy-on-close="true" :close-on-click-modal="false" :close-on-press-escape="false">
      <el-form
          ref="formRef"
          :model="newFileForm"
          label-width="auto"
          label-position="top"
          size="default"
          :rules="rules"
      >
        <el-form-item label="文件名" prop="fileName">
          <el-input v-model="newFileForm.fileName" placeholder="请输入新文件文件名"/>
        </el-form-item>
        <el-form-item label="内容" prop="fileContent">
          <el-input v-model="newFileForm.fileContent"  type="textarea" :rows="8" placeholder="请输入新文件内容"/>
        </el-form-item>
      </el-form>
      <template #footer>
      <span class="dialog-footer">
        <el-button @click="closeAddFile">关 闭</el-button>
        <el-button type="primary" @click="saveNewFile">
          保 存
        </el-button>
      </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.sftp-container {
  height: calc(100vh - 147px);
  padding: 10px;
  overflow: auto;
}

.custom-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr); /* 4列 */
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

.a-center {
  display: flex;
  align-items: center;
  justify-content: center;
}
.a-row{
  flex-direction: column;
  height: 121px;
}
.title-text{
  font-size: 13px;
  font-weight: bold;
  user-select: none;
  margin-top: 5px;
}
.unchose-color{
  color: #969696;
}
.y-font{
  font-family: 幼圆;
}
.card-height{
  cursor: pointer;
  height: 150px;
}
.ellipsis {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 115px; /* 指定最大宽度 */
}
.folderBg{
  background-color: #ffffff;
}
.fileBg{
  background-color: #f3f3f3;
}
</style>