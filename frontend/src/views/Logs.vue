<template>
    <div class="p-4 flex flex-col gap-4">
        <h1 class="text-2xl font-bold mb-4">System Logs</h1>

        <div v-if="loading" class="flex justify-center">
             <i class="pi pi-spin pi-spinner text-4xl"></i>
        </div>

        <div v-else class="bg-gray-800 rounded p-4 font-mono text-sm h-[600px] overflow-y-auto flex flex-col-reverse">
             <div v-for="(log, index) in logs" :key="index" class="mb-1 border-b border-gray-700 pb-1">
                 <span class="text-gray-400">[{{ formatDate(log.time) }}]</span>
                 <span :class="getLevelColor(log.level)" class="mx-2 font-bold">{{ log.level }}</span>
                 <span class="text-blue-300">{{ log.component }}:</span>
                 <span class="text-white ml-2">{{ log.message }}</span>
             </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import axios from 'axios';

const logs = ref([]);
const loading = ref(true);
const timer = ref(null);

onMounted(() => {
    fetchLogs();
    timer.value = setInterval(fetchLogs, 5000);
});

onUnmounted(() => {
    if (timer.value) clearInterval(timer.value);
});

const fetchLogs = async () => {
    try {
        const res = await axios.get('/api/logs');
        // Backend returns recent logs first, but we want to append them or replace
        // For simplicity, just replace and let Vue handle DOM updates
        logs.value = res.data;
    } catch (e) {
        console.error("Failed to fetch logs", e);
    } finally {
        loading.value = false;
    }
};

const getLevelColor = (level) => {
    switch(level) {
        case 'INFO': return 'text-green-400';
        case 'WARN': return 'text-yellow-400';
        case 'ERROR': return 'text-red-400';
        case 'FATAL': return 'text-red-600';
        default: return 'text-gray-400';
    }
};

const formatDate = (ts) => {
    return new Date(ts).toLocaleString();
}
</script>
