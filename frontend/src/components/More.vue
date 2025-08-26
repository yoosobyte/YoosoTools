<script setup>
import { ref } from 'vue'
const loading = ref(true)
</script>
<template>
  <div style="width: 100%;height: calc(100vh - 78px);">
    <iframe
        v-loading="loading"
        src="https://element-plus.sxtxhy.com/zh-CN/component/button.html"
        frameborder="0"
        style="width: 100%; height: 100%;"
        @load="loading = false"
    />
  </div>
</template>