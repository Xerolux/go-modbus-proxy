<template>
    <div class="p-4 flex flex-col gap-4">
        <h1 class="text-2xl font-bold mb-4">Configuration</h1>

        <div v-if="loading" class="flex justify-center">
            <i class="pi pi-spin pi-spinner text-4xl"></i>
        </div>

        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card class="bg-gray-800 text-white">
                <template #title>Web Interface</template>
                <template #content>
                     <div class="flex flex-col gap-4">
                         <div class="flex flex-col gap-2">
                            <label>Web Server Port</label>
                            <InputText v-model="webPort" />
                            <small class="text-gray-400">Change requires restart</small>
                        </div>
                     </div>
                </template>
            </Card>

            <Card class="bg-gray-800 text-white">
                <template #title>System Actions</template>
                <template #content>
                    <div class="flex flex-col gap-4">
                        <Button label="Restart System" icon="pi pi-refresh" severity="danger" @click="confirmRestart" />
                    </div>
                </template>
            </Card>
        </div>

        <div class="flex gap-4 mt-4">
            <Button label="Save Configuration" icon="pi pi-save" @click="saveConfig" :loading="saving" />
        </div>

        <Toast />
        <ConfirmDialog />
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import Card from 'primevue/card';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import Toast from 'primevue/toast';
import ConfirmDialog from 'primevue/confirmdialog';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';

const webPort = ref('');
const loading = ref(true);
const saving = ref(false);
const toast = useToast();
const confirm = useConfirm();

onMounted(async () => {
    try {
        const res = await axios.get('/api/config/webport');
        webPort.value = res.data.web_port;
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to load config', life: 3000 });
    } finally {
        loading.value = false;
    }
});

const saveConfig = async () => {
    saving.value = true;
    try {
        await axios.put('/api/config/webport', { web_port: webPort.value });
        toast.add({ severity: 'success', summary: 'Success', detail: 'Configuration saved. Please restart.', life: 3000 });
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.error || e.message, life: 5000 });
    } finally {
        saving.value = false;
    }
};

const confirmRestart = () => {
    confirm.require({
        message: 'Are you sure you want to restart the service?',
        header: 'Confirmation',
        icon: 'pi pi-exclamation-triangle',
        accept: async () => {
            try {
                await axios.post('/api/system/restart');
                toast.add({ severity: 'info', summary: 'Restarting', detail: 'System is restarting...', life: 3000 });
            } catch (e) {
                toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to restart', life: 3000 });
            }
        }
    });
};
</script>
