<template>
     <div class="p-4 flex flex-col gap-4">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">Proxy Control</h1>
            <div class="flex gap-2">
                <Button
                    icon="pi pi-plus"
                    severity="info"
                    label="Add Proxy"
                    @click="openAddProxyDialog"
                    class="text-xs sm:text-sm"
                />
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

                         <div class="grid grid-cols-4 gap-2 mt-2">
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
                              <Button
                                 icon="pi pi-pencil"
                                 severity="secondary"
                                 label="Edit"
                                 @click="openEditProxyDialog(proxy)"
                                 class="text-xs sm:text-sm"
                              />
                         </div>
                         <Button
                             icon="pi pi-eye"
                             severity="secondary"
                             label="View Logs"
                             @click="openProxyLogs(proxy.id)"
                             class="text-xs sm:text-sm mt-2"
                         />
                         <Button
                             icon="pi pi-trash"
                             severity="danger"
                             label="Delete"
                             @click="confirmDeleteProxy(proxy.id)"
                             class="text-xs sm:text-sm mt-2"
                         />
                     </div>
                 </template>
             </Card>
         </div>

         <Dialog v-model:visible="showProxyDialog" :header="isEditMode ? 'Edit Proxy' : 'Add Proxy'" modal class="w-full max-w-lg">
             <div class="flex flex-col gap-4">
                 <div>
                     <label class="block text-sm font-medium mb-1">Name</label>
                     <InputText v-model="proxyForm.name" class="w-full" />
                 </div>
                 <div>
                     <label class="block text-sm font-medium mb-1">Listen Address</label>
                     <InputText v-model="proxyForm.listen_addr" placeholder=":5020" class="w-full" />
                 </div>
                 <div>
                     <label class="block text-sm font-medium mb-1">Target Address</label>
                     <InputText v-model="proxyForm.target_addr" placeholder="192.168.1.100:502" class="w-full" />
                 </div>
                 <div>
                     <label class="block text-sm font-medium mb-1">Description</label>
                     <InputText v-model="proxyForm.description" class="w-full" />
                 </div>
                 <div class="flex gap-4">
                     <div class="flex-1">
                         <label class="block text-sm font-medium mb-1">Connection Timeout (s)</label>
                         <InputNumber v-model="proxyForm.connection_timeout" :min="1" class="w-full" />
                     </div>
                     <div class="flex-1">
                         <label class="block text-sm font-medium mb-1">Read Timeout (s)</label>
                         <InputNumber v-model="proxyForm.read_timeout" :min="1" class="w-full" />
                     </div>
                 </div>
                 <div class="flex gap-4">
                     <div class="flex-1">
                         <label class="block text-sm font-medium mb-1">Max Retries</label>
                         <InputNumber v-model="proxyForm.max_retries" :min="0" class="w-full" />
                     </div>
                     <div class="flex-1">
                         <label class="block text-sm font-medium mb-1">Max Read Size (0=unlimited)</label>
                         <InputNumber v-model="proxyForm.max_read_size" :min="0" class="w-full" />
                     </div>
                 </div>
                 <div class="flex items-center gap-4">
                     <div class="flex items-center gap-2">
                         <Checkbox v-model="proxyForm.enabled" binary />
                         <span class="text-sm">Enabled</span>
                     </div>
                     <div class="flex items-center gap-2">
                         <Checkbox v-model="proxyForm.paused" binary />
                         <span class="text-sm">Paused</span>
                     </div>
                 </div>
                 <div>
                     <label class="block text-sm font-medium mb-1">Tags</label>
                     <Chips v-model="proxyForm.tags" class="w-full" placeholder="Add tags" />
                 </div>
             </div>
             <template #footer>
                 <Button label="Cancel" severity="secondary" @click="showProxyDialog = false" />
                 <Button :label="isEditMode ? 'Update' : 'Create'" @click="saveProxy" />
             </template>
         </Dialog>

         <Dialog v-model:visible="showLogsDialog" :header="`Logs - ${currentProxy?.name}`" modal class="w-full max-w-4xl">
             <div class="bg-gray-900 rounded p-4 font-mono text-sm h-[500px] overflow-y-auto">
                 <div v-if="proxyLogs.length === 0" class="text-gray-500">No logs available</div>
                 <div v-else class="space-y-1">
                     <div v-for="(log, index) in proxyLogs" :key="index" class="border-b border-gray-700 pb-1">
                         <span class="text-gray-400">[{{ formatLogTime(log.time) }}]</span>
                         <span :class="getLogLevelColor(log.level)" class="mx-2 font-bold">{{ log.level }}</span>
                         <span class="text-white">{{ log.message }}</span>
                     </div>
                 </div>
             </div>
         </Dialog>
         <Toast />
         <ConfirmDialog />
     </div>
</template>

<script setup>
 import { ref, onMounted, onUnmounted, watch } from 'vue';
 import axios from 'axios';
 import Card from 'primevue/card';
 import Button from 'primevue/button';
 import Tag from 'primevue/tag';
 import Dialog from 'primevue/dialog';
 import InputText from 'primevue/inputtext';
 import InputNumber from 'primevue/inputnumber';
 import Checkbox from 'primevue/checkbox';
 import Chips from 'primevue/chips';
 import Toast from 'primevue/toast';
 import ConfirmDialog from 'primevue/confirmdialog';
 import { useToast } from 'primevue/usetoast';
 import { useConfirm } from 'primevue/useconfirm';
 import { useEventSource } from '../utils/eventSource';

 const proxies = ref([]);
 const loading = ref(true);
 const toast = useToast();
 const confirm = useConfirm();
 let eventSource = null;

 const showProxyDialog = ref(false);
 const isEditMode = ref(false);
 const proxyForm = ref({
     id: '',
     name: 'New Proxy',
     listen_addr: ':5020',
     target_addr: '127.0.0.1:502',
     description: '',
     connection_timeout: 10,
     read_timeout: 30,
     max_retries: 3,
     max_read_size: 0,
     enabled: true,
     paused: false,
     tags: []
 });

 const showLogsDialog = ref(false);
 const currentProxy = ref(null);
 const proxyLogs = ref([]);

 const fetchProxies = async () => {
     try {
         const res = await axios.get('/api/proxies');
         proxies.value = res.data;
     } catch (e) {
         console.error("Failed to fetch proxies");
     }
 };

 onMounted(async () => {
     await fetchProxies();
     loading.value = false;

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

 const openAddProxyDialog = () => {
     isEditMode.value = false;
     proxyForm.value = {
         id: '',
         name: 'New Proxy',
         listen_addr: ':5020',
         target_addr: '127.0.0.1:502',
         description: '',
         connection_timeout: 10,
         read_timeout: 30,
         max_retries: 3,
         max_read_size: 0,
         enabled: true,
         paused: false,
         tags: []
     };
     showProxyDialog.value = true;
 };

 const openEditProxyDialog = (proxy) => {
     isEditMode.value = true;
     proxyForm.value = { ...proxy };
     showProxyDialog.value = true;
 };

 const saveProxy = async () => {
     try {
         if (isEditMode.value) {
             await axios.put('/api/proxies', proxyForm.value);
             toast.add({ severity: 'success', summary: 'Success', detail: 'Proxy updated', life: 3000 });
         } else {
             await axios.post('/api/proxies', proxyForm.value);
             toast.add({ severity: 'success', summary: 'Success', detail: 'Proxy created', life: 3000 });
         }
         showProxyDialog.value = false;
         await fetchProxies();
     } catch (e) {
         toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.error || e.message, life: 5000 });
     }
 };

 const confirmDeleteProxy = (id) => {
     confirm.require({
         message: 'Are you sure you want to delete this proxy?',
         header: 'Confirm Delete',
         icon: 'pi pi-exclamation-triangle',
         accept: async () => {
             try {
                 await axios.delete(`/api/proxies?id=${id}`);
                 toast.add({ severity: 'success', summary: 'Success', detail: 'Proxy deleted', life: 3000 });
                 await fetchProxies();
             } catch (e) {
                 toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.error || e.message, life: 5000 });
             }
         }
     });
 };

 const openProxyLogs = async (id) => {
     currentProxy.value = proxies.value.find(p => p.id === id);
     try {
         const res = await axios.get('/api/logs');
         const allLogs = res.data;
         proxyLogs.value = allLogs.filter(log => log.proxy_id === id || (log.component && log.component.includes(id)));
     } catch (e) {
         console.error("Failed to fetch logs", e);
         proxyLogs.value = [];
     }
     showLogsDialog.value = true;
 };

 const formatLogTime = (dateStr) => {
     const date = new Date(dateStr);
     return date.toLocaleString('de-DE', {
         hour: '2-digit',
         minute: '2-digit',
         second: '2-digit',
     });
 };

 const getLogLevelColor = (level) => {
     switch(level) {
         case 'INFO': return 'text-green-400';
         case 'WARN': return 'text-yellow-400';
         case 'ERROR': return 'text-red-400';
         case 'FATAL': return 'text-red-600';
         default: return 'text-gray-400';
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
