<script lang="ts" setup>
import ToolMain from './components/ToolMain.vue'
window.addEventListener('dragover',  e => e.preventDefault(), false);
window.addEventListener('drop',      e => e.preventDefault(), false);
</script>
<template>
  <ToolMain />
</template>
