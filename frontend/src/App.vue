<template>
  <div id="app">
    <CommentList :product-id="productId" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import CommentList from './components/CommentList.vue'

const productId = ref('')

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  productId.value = params.get('product_id') || ''
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto,
    'Helvetica Neue', Arial, sans-serif;
  background: #f5f5f5;
  color: #333;
  -webkit-font-smoothing: antialiased;
}

#app {
  min-height: 100vh;
}
</style>
