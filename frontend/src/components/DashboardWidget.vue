<script setup>
import { defineProps } from 'vue';

const props = defineProps({
    title: String,
    value: [String, Number],
    unit: String
});
</script>

<template>
    <div class="h-full flex flex-col justify-between p-2">
        <div class="text-gray-400 text-sm font-medium uppercase tracking-wider">{{ title }}</div>
        <div class="text-2xl font-bold text-blue-400 my-2 truncate" :title="value">
            {{ value }} <span v-if="unit" class="text-sm text-gray-500 ml-1">{{ unit }}</span>
        </div>
    </div>
</template>
