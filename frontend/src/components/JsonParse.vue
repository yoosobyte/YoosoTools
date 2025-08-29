<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { MagicStick, DocumentCopy, Fold, Expand } from '@element-plus/icons-vue'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'

const inputJson = ref({
  one:'',
  oneObj:null,
  two:'',
  twoObj:null,
})
const expandOne = ref(true);
const viewOne = ref(true);
function copyResultOne() {
  navigator.clipboard.writeText(inputJson.value.one)
      .then(() => ElMessage.success('已复制'))
      .catch(() => ElMessage.error('复制失败'))
}
function toggleExpandOne() {
  expandOne.value = !expandOne.value;
}
function toggleViewOne() {
  if (viewOne.value) {
    if(inputJson.value.one===""){
      return;
    }
    formatOne();
  }
  viewOne.value = !viewOne.value;
}

const formatOne = () => {
  try {
    inputJson.value.oneObj = JSON.parse(inputJson.value.one);
    inputJson.value.one = JSON.stringify(inputJson.value.oneObj, null, 2)
    // ElMessage.success('格式化成功');
  } catch (err) {
    ElMessage.error('JSON格式错误: ' + err.message)
    inputJson.value.oneObj = null;
  }
}



const expandTwo = ref(true);
const viewTwo = ref(true);
function copyResultTwo() {
  navigator.clipboard.writeText(inputJson.value.two)
      .then(() => ElMessage.success('已复制'))
      .catch(() => ElMessage.error('复制失败'))
}
function toggleExpandTwo() {
  expandTwo.value = !expandTwo.value;
}
function toggleViewTwo() {
  if (viewTwo.value) {
    if(inputJson.value.two===""){
      return;
    }
    formatTwo();
  }
  viewTwo.value = !viewTwo.value;
}

const formatTwo = () => {
  try {
    inputJson.value.twoObj = JSON.parse(inputJson.value.two);
    inputJson.value.two = JSON.stringify(inputJson.value.twoObj, null, 2)
    // ElMessage.success('格式化成功');
  } catch (err) {
    ElMessage.error('JSON格式错误: ' + err.message)
    inputJson.value.twoObj = null;
  }
}



const leftViewer = ref(null);
const rightViewer = ref(null);
let isSyncing = false;

const handleScroll = (source) => {
  if (isSyncing) return;

  isSyncing = true;

  const sourceEl = source === 'left'
      ? leftViewer.value.$el
      : rightViewer.value.$el;

  const targetEl = source === 'left'
      ? rightViewer.value.$el
      : leftViewer.value.$el;

  // 同步滚动位置
  targetEl.scrollTop = sourceEl.scrollTop;

  requestAnimationFrame(() => {
    isSyncing = false;
  });
};
</script>
<template>
  <div style="display: flex;margin-top: 10px;">
    <el-card style="height: calc(100vh - 90px);display: flex;width: 50%;" class="my-card">
      <div class="card-header" style="">
        <div style="width: 50%;font-size: 16px;text-align: start;padding-left: 10px;font-weight: bold;">JSON-A区 </div>
        <div style="width: 50%;text-align: end;padding-right: 2px;">
          <el-button-group>
            <el-button @click="toggleViewOne">
              <el-icon><DocumentCopy /></el-icon>
              {{ viewOne ? '格式化' : '返回编辑' }}
            </el-button>
            <el-button v-if="!viewOne" @click="toggleExpandOne">
              <el-icon v-if="expandOne"><Fold /></el-icon>
              <el-icon v-if="!expandOne"><Expand /></el-icon>
              {{ expandOne ? '折叠' : '展开' }}
            </el-button>
            <el-button v-if="!viewOne" @click="copyResultOne">
              <el-icon><DocumentCopy /></el-icon>
              复制
            </el-button>
          </el-button-group>
        </div>
      </div>
      <div class="textarea-container" v-if="viewOne">
        <el-input
            v-model="inputJson.one"
            :rows="2"
            type="textarea"
            placeholder='例如: {"name":"John","age":30}'
            resize="none"
            class="json-input"
        />
      </div>
      <div class="textarea-container" v-if="!viewOne && inputJson.oneObj">
        <vue-json-pretty
            style="margin-left: 11px;overflow: auto;width: 100%;border-top: 1px solid #eeeeee;padding-top: 5px;"
            :data="inputJson.oneObj"
            :deep="expandOne ? 10 : 3"
            :showLength="true"
            :highlightMouseoverNode="true"
            @scroll="handleScroll('left')"
            ref="leftViewer"
        />
      </div>
    </el-card>
    <div style="width: 10px;height: calc(100vh - 90px);display: flex;align-items: center;justify-content: center;flex-direction: column;">
      <img style="width: 100%;display: block;" src="@/assets/images/vs.png" alt=""/>
    </div>
    <el-card style="height: calc(100vh - 90px);display: flex;width: 50%;" class="my-card">
      <div class="card-header">
        <div style="width: 50%;font-size: 16px;text-align: start;padding-left: 10px;font-weight: bold;">JSON-B区 </div>
        <div style="width: 50%;text-align: end;padding-right: 2px;">
          <el-button-group>
            <el-button @click="toggleViewTwo">
              <el-icon><DocumentCopy /></el-icon>
              {{ viewTwo ? '格式化' : '返回编辑' }}
            </el-button>
            <el-button v-if="!viewTwo" @click="toggleExpandTwo">
              <el-icon v-if="expandTwo"><Fold /></el-icon>
              <el-icon v-if="!expandTwo"><Expand /></el-icon>
              {{ expandTwo ? '折叠' : '展开' }}
            </el-button>
            <el-button v-if="!viewTwo" @click="copyResultTwo">
              <el-icon><DocumentCopy /></el-icon>
              复制
            </el-button>
          </el-button-group>
        </div>
      </div>
      <div class="textarea-container" v-if="viewTwo">
        <el-input
            v-model="inputJson.two"
            :rows="2"
            type="textarea"
            placeholder='例如: {"name":"Ethan","age":25}'
            resize="none"
            class="json-input"
        />
      </div>
      <div class="textarea-container" v-if="!viewTwo && inputJson.twoObj">
        <vue-json-pretty
            style="margin-left: 11px;overflow: auto;width: 100%;border-top: 1px solid #eeeeee;padding-top: 5px;"
            :data="inputJson.twoObj"
            :deep="expandTwo ? 10 : 3"
            :showLength="true"
            :highlightMouseoverNode="true"
            @scroll="handleScroll('right')"
            ref="rightViewer"
        />
      </div>
    </el-card>
  </div>
</template>
<style>
.my-card {
  .el-card__body{
    padding: 0;
    width: 100%;
  }
}
</style>
<style scoped>
.card-header{
  height: 41px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-around;
}
.json-input {
  font-family: 'Fira Code', 'Consolas', monospace;
}
.textarea-container {
  height: calc(100vh - 132px);
  display: flex;
}

:deep(.json-input .el-textarea__inner) {
  min-height: 100% !important;
}
</style>