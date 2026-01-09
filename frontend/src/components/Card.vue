<template>
  <div class="bg-white rounded-lg shadow-md overflow-hidden border border-gray-200">
    <div v-if="title" class="px-6 py-4 border-b border-gray-100 bg-gray-50 flex justify-between items-center cursor-move handle">
      <h3 class="text-lg font-medium text-gray-900">{{ title }}</h3>
      <slot name="header-actions"></slot>
    </div>
    <div class="p-6">
      <slot></slot>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: String
});
</script>
