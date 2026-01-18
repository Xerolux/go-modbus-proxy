<template>
     <div class="p-4 flex flex-col gap-4">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">Proxy Control</h1>
            <div class="flex gap-2">
                <Button
                    icon="pi pi-play"
                    severity="success"
                    label="Start All"
                    @click="controlAllProxies('start_all')"
                    class="text-xs sm:text-sm"
                />
                <Button
                    icon="pi pi-stop"
                    severity="danger"
                    label="Stop All"
                    @click="controlAllProxies('stop_all')"
                    class="text-xs sm:text-sm"
                />
            </div>
        </div>

         <div v-if="loading" class="flex justify-center">
             <i class="pi pi-spin pi-spinner text-4xl"></i>
         </div>

         <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
             <Card v-for="proxy in proxies" :key="proxy.id" class="bg-gray-800 text-white">
                 <template #title>
                     <div class="flex justify-between items-center">
                         <span class="text-lg truncate" :title="proxy.name">{{ proxy.name }}</span>
                         <Tag :severity="getSeverity(proxy.status)" :value="proxy.status" />
                     </div>
                 </template>
                 <template #content>
                     <div class="flex flex-col gap-3">
                         <div class="text-gray-400 text-sm">{{ proxy.description || 'No description' }}</div>
                         <div class="text-sm">Listen: {{ proxy.listen_addr }}</div>
                         <div class="text-sm">Target: {{ proxy.target_addr }}</div>

                         <div class="grid grid-cols-3 gap-2 mt-2">
                              <Button
                                 icon="pi pi-play"
                                 severity="success"
                                 label="Start"
                                 :disabled="proxy.status === 'Running'"
                                 @click="controlProxy(proxy.id, 'start')"
                                 class="text-xs sm:text-sm"
                              />
                              <Button
                                 icon="pi pi-stop"
                                 severity="danger"
                                 label="Stop"
                                 :disabled="proxy.status === 'Stopped' || proxy.status === 'Error'"
                                 @click="controlProxy(proxy.id, 'stop')"
                                 class="text-xs sm:text-sm"
                              />
                               <Button
                                 icon="pi pi-refresh"
                                 severity="info"
                                 label="Restart"
                                 :disabled="proxy.status === 'Stopped'"
                                 @click="controlProxy(proxy.id, 'restart')"
                                 class="text-xs sm:text-sm"
                              />
                         </div>
                     </div>
                 </template>
             </Card>
         </div>
          <Toast />
     </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';
import axios from 'axios';
import Card from 'primevue/card';
import Button from 'primevue/button';
import Tag from 'primevue/tag';
import Toast from 'primevue/toast';
import { useToast } from 'primevue/usetoast';
import { useEventSource } from '../utils/eventSource';

const proxies = ref([]);
const loading = ref(true);
const toast = useToast();
let eventSource = null;

onMounted(async () => {
    try {
        const res = await axios.get('/api/proxies');
        proxies.value = res.data;
    } catch (e) {
        console.error("Failed to fetch proxies");
    } finally {
        loading.value = false;
    }

    const { data, disconnect, isConnected } = useEventSource('/api/proxies/stream');

    watch(isConnected, (connected) => {
        if (!connected) {
            console.warn('SSE connection lost, polling fallback enabled');
        }
    });

    watch(data, (eventData) => {
        if (!eventData) return;

        const eventType = eventData.type;
        const proxyData = eventData.proxy;

        switch (eventType) {
            case 'proxy_added':
            case 'proxy_updated':
            case 'proxy_started':
            case 'proxy_stopped':
                if (proxyData) {
                    const index = proxies.value.findIndex(p => p.id === proxyData.id);
                    if (index !== -1) {
                        proxies.value[index] = proxyData;
                    } else {
                        proxies.value.push(proxyData);
                    }
                }
                break;
            case 'proxy_removed':
                if (eventData.proxy_id) {
                    proxies.value = proxies.value.filter(p => p.id !== eventData.proxy_id);
                }
                break;
        }
    });
});

onUnmounted(() => {
    if (eventSource) {
        eventSource.disconnect();
    }
});

const controlProxy = async (id, action) => {
    try {
        await axios.post('/api/proxies/control', { id, action });
        toast.add({ severity: 'success', summary: 'Success', detail: `Proxy ${action} command sent`, life: 3000 });
        setTimeout(fetchProxies, 500);
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.error || e.message, life: 5000 });
    }
};

const controlAllProxies = async (action) => {
    try {
        await axios.post('/api/proxies/control', { action: action === 'start_all' ? 'start_all' : 'stop_all' });
        toast.add({ severity: 'success', summary: 'Success', detail: `All proxies ${action.replace('_all', '')} command sent`, life: 3000 });
        setTimeout(fetchProxies, 500);
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.error || e.message, life: 5000 });
    }
};

const getSeverity = (status) => {
    switch(status) {
        case 'Running': return 'success';
        case 'Stopped': return 'secondary';
        case 'Error': return 'danger';
        default: return 'info';
    }
};
</script>
