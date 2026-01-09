<template>
    <div class="p-4 flex flex-col gap-4">
        <h1 class="text-2xl font-bold mb-4">Proxy Control</h1>

        <div v-if="loading" class="flex justify-center">
            <i class="pi pi-spin pi-spinner text-4xl"></i>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
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

                        <div class="flex gap-2 mt-2">
                             <Button
                                icon="pi pi-play"
                                severity="success"
                                label="Start"
                                :disabled="proxy.status === 'Running'"
                                @click="controlProxy(proxy.id, 'start')"
                             />
                             <Button
                                icon="pi pi-stop"
                                severity="danger"
                                label="Stop"
                                :disabled="proxy.status === 'Stopped' || proxy.status === 'Error'"
                                @click="controlProxy(proxy.id, 'stop')"
                             />
                              <Button
                                icon="pi pi-refresh"
                                severity="info"
                                label="Restart"
                                :disabled="proxy.status === 'Stopped'"
                                @click="controlProxy(proxy.id, 'restart')"
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
import { ref, onMounted, onUnmounted } from 'vue';
import axios from 'axios';
import Card from 'primevue/card';
import Button from 'primevue/button';
import Tag from 'primevue/tag';
import Toast from 'primevue/toast';
import { useToast } from 'primevue/usetoast';

const proxies = ref([]);
const loading = ref(true);
const toast = useToast();
const timer = ref(null);

onMounted(async () => {
    fetchProxies();
    timer.value = setInterval(fetchProxies, 2000);
});

onUnmounted(() => {
    if (timer.value) clearInterval(timer.value);
});

const fetchProxies = async () => {
    try {
        const res = await axios.get('/api/proxies');
        proxies.value = res.data;
    } catch (e) {
        console.error("Failed to fetch proxies");
    } finally {
        loading.value = false;
    }
};

const controlProxy = async (id, action) => {
    try {
        await axios.post('/api/proxies/control', { id, action });
        toast.add({ severity: 'success', summary: 'Success', detail: `Proxy ${action} command sent`, life: 3000 });
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
