<template>
  <div class="flex items-center space-x-2">
    <div
      class="w-3 h-3 rounded-full"
      :class="{
        'bg-green-500': status === 'Running',
        'bg-red-500': status !== 'Running' && status !== 'Loading...',
        'bg-yellow-500': status === 'Loading...'
      }"
    ></div>
    <span class="font-medium text-gray-700">{{ status }}</span>
  </div>
</template>

<script setup>
defineProps({
  status: String
});
</script>
