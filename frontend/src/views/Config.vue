<template>
    <div class="p-4 flex flex-col gap-4">
        <h1 class="text-2xl font-bold mb-4">Configuration</h1>

        <div v-if="loading" class="flex justify-center">
            <i class="pi pi-spin pi-spinner text-4xl"></i>
        </div>

        <div v-else class="flex flex-col gap-6">
            <!-- New ConfigForm Component -->
            <Card class="bg-gray-800 text-white">
                <template #title>System Configuration</template>
                <template #content>
                     <ConfigForm />
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

        <Toast />
        <ConfirmDialog />
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import Card from 'primevue/card';
import Button from 'primevue/button';
import Toast from 'primevue/toast';
import ConfirmDialog from 'primevue/confirmdialog';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import ConfigForm from '../components/ConfigForm.vue';
import { useAppStore } from '../stores/appStore';

const loading = ref(true);
const toast = useToast();
const confirm = useConfirm();
const store = useAppStore();

onMounted(async () => {
    try {
        await store.fetchWebPort();
        await store.fetchProxies();
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to load config', life: 3000 });
    } finally {
        loading.value = false;
    }
});

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
