<template>
    <div class="p-4 flex flex-col gap-4">
        <h1 class="text-2xl font-bold mb-4 text-gray-200">Configuration</h1>

        <div v-if="loading" class="flex justify-center">
            <i class="pi pi-spin pi-spinner text-4xl text-blue-500"></i>
        </div>

        <div v-else class="flex flex-col gap-6">
            <Tabs value="0">
                <TabList class="bg-gray-800 text-gray-200">
                    <Tab value="0">Proxies</Tab>
                    <Tab value="1">Logging</Tab>
                    <Tab value="2">Security</Tab>
                    <Tab value="3">Email</Tab>
                    <Tab value="4">Backup</Tab>
                    <Tab value="5">Advanced</Tab>
                </TabList>

                <TabPanels class="bg-gray-800 text-white p-4 rounded">
                    <TabPanel value="0">
                        <ConfigForm />
                    </TabPanel>

                    <TabPanel value="1">
                        <div class="space-y-4">
                            <h3 class="text-lg font-semibold">Logging Configuration</h3>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Log Level</label>
                                    <Dropdown v-model="config.log_level" :options="logLevels" optionLabel="label" optionValue="value" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Max File Size (MB)</label>
                                    <InputNumber v-model="config.log_max_size" :min="1" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Max Files</label>
                                    <InputNumber v-model="config.log_max_files" :min="1" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Max Age (Days)</label>
                                    <InputNumber v-model="config.log_max_age_days" :min="1" class="w-full" />
                                </div>
                            </div>

                            <Button @click="saveConfig" label="Save Logging Configuration" icon="pi pi-save" />
                        </div>
                    </TabPanel>

                    <TabPanel value="2">
                        <div class="space-y-6">
                            <div>
                                <h3 class="text-lg font-semibold mb-4">SSL/TLS</h3>
                                <div class="grid grid-cols-1 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Enable SSL/TLS</label>
                                        <ToggleSwitch v-model="config.tls_enabled" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Certificate File</label>
                                        <InputText v-model="config.tls_cert_file" class="w-full" placeholder="/path/to/cert.pem" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Key File</label>
                                        <InputText v-model="config.tls_key_file" class="w-full" placeholder="/path/to/key.pem" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Session Timeout (Hours)</label>
                                        <InputNumber v-model="config.session_timeout" :min="1" class="w-full" />
                                    </div>
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">CORS</h3>
                                <div class="grid grid-cols-1 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Allowed Origins</label>
                                        <Chips v-model="config.cors_allowed_origins" class="w-full" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Allowed Methods</label>
                                        <Chips v-model="config.cors_allowed_methods" class="w-full" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Allowed Headers</label>
                                        <Chips v-model="config.cors_allowed_headers" class="w-full" />
                                    </div>
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">Rate Limiting</h3>
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Enable Rate Limiting</label>
                                        <ToggleSwitch v-model="config.rate_limit_enabled" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Requests per Minute</label>
                                        <InputNumber v-model="config.rate_limit_requests" :min="1" class="w-full" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Burst Size</label>
                                        <InputNumber v-model="config.rate_limit_burst" :min="1" class="w-full" />
                                    </div>
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">IP Filtering</h3>
                                <div class="grid grid-cols-1 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Enable IP Whitelist</label>
                                        <ToggleSwitch v-model="config.ip_whitelist_enabled" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">IP Whitelist</label>
                                        <Chips v-model="config.ip_whitelist" class="w-full" placeholder="192.168.1.0/24" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Enable IP Blacklist</label>
                                        <ToggleSwitch v-model="config.ip_blacklist_enabled" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">IP Blacklist</label>
                                        <Chips v-model="config.ip_blacklist" class="w-full" placeholder="10.0.0.0/8" />
                                    </div>
                                </div>
                            </div>

                            <Button @click="saveConfig" label="Save Security Configuration" icon="pi pi-shield" />
                        </div>
                    </TabPanel>

                    <TabPanel value="3">
                        <div class="space-y-4">
                            <h3 class="text-lg font-semibold">Email Configuration</h3>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Enable Email Alerts</label>
                                    <ToggleSwitch v-model="config.email_enabled" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">SMTP Server</label>
                                    <InputText v-model="config.email_smtp_server" class="w-full" placeholder="smtp.gmail.com" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">SMTP Port</label>
                                    <InputNumber v-model="config.email_smtp_port" :min="1" :max="65535" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">From Email</label>
                                    <InputText v-model="config.email_from" class="w-full" placeholder="noreply@example.com" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">To Email</label>
                                    <InputText v-model="config.email_to" class="w-full" placeholder="admin@example.com" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Username</label>
                                    <InputText v-model="config.email_username" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Password</label>
                                    <Password v-model="config.email_password" :feedback="false" toggleMask class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Alert on Error</label>
                                    <ToggleSwitch v-model="config.email_alert_on_error" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Alert on Warning</label>
                                    <ToggleSwitch v-model="config.email_alert_on_warning" />
                                </div>
                            </div>

                            <Button @click="saveConfig" label="Save Email Configuration" icon="pi pi-envelope" />
                        </div>
                    </TabPanel>

                    <TabPanel value="4">
                        <div class="space-y-4">
                            <h3 class="text-lg font-semibold">Backup Configuration</h3>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Enable Backups</label>
                                    <ToggleSwitch v-model="config.backup_enabled" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Backup Interval</label>
                                    <Dropdown v-model="config.backup_interval" :options="backupIntervals" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Retention (Count)</label>
                                    <InputNumber v-model="config.backup_retention" :min="1" class="w-full" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Backup Path</label>
                                    <InputText v-model="config.backup_path" class="w-full" placeholder="./backups" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Backup Database</label>
                                    <ToggleSwitch v-model="config.backup_database" />
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-300 mb-1">Backup Configuration</label>
                                    <ToggleSwitch v-model="config.backup_config" />
                                </div>
                            </div>

                            <Button @click="saveConfig" label="Save Backup Configuration" icon="pi pi-download" />
                        </div>
                    </TabPanel>

                    <TabPanel value="5">
                        <div class="space-y-6">
                            <div>
                                <h3 class="text-lg font-semibold mb-4">Advanced Configuration</h3>
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Debug Mode</label>
                                        <ToggleSwitch v-model="config.debug_mode" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Max Connections</label>
                                        <InputNumber v-model="config.max_connections" :min="1" class="w-full" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Enable Metrics</label>
                                        <ToggleSwitch v-model="config.metrics_enabled" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Metrics Port</label>
                                        <InputText v-model="config.metrics_port" class="w-full" placeholder=":9090" />
                                    </div>
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">Password</h3>
                                <div class="grid grid-cols-1 gap-4">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">Current Password</label>
                                        <Password v-model="passwordForm.currentPassword" :feedback="false" toggleMask class="w-full" />
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-300 mb-1">New Password</label>
                                        <Password v-model="passwordForm.newPassword" toggleMask class="w-full" />
                                        <div class="mt-2 p-3 bg-blue-500/10 border border-blue-500/30 rounded-lg">
                                            <p class="text-xs text-blue-300 font-medium mb-1">Passwort-Anforderungen:</p>
                                            <ul class="text-xs text-gray-400 space-y-1 ml-4 list-disc">
                                                <li>Mindestens 8 Zeichen lang</li>
                                                <li>Mindestens 3 von: Großbuchstaben, Kleinbuchstaben, Zahlen, Sonderzeichen</li>
                                                <li>Nicht zu einfach oder häufig verwendet</li>
                                            </ul>
                                        </div>
                                    </div>
                                    <Button label="Change Password" icon="pi pi-key" @click="changePassword" />
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">Configuration Backup</h3>
                                <div class="flex gap-4">
                                    <Button label="Export Configuration" icon="pi pi-download" @click="exportConfig" />
                                    <Button label="Import Configuration" icon="pi pi-upload" severity="secondary" @click="triggerImport" />
                                    <input type="file" ref="importFile" accept=".json" @change="importConfig" style="display: none" />
                                </div>
                            </div>

                            <div>
                                <h3 class="text-lg font-semibold mb-4">System Actions</h3>
                                <div class="flex gap-4">
                                    <Button label="Restart System" icon="pi pi-refresh" severity="danger" @click="confirmRestart" />
                                </div>
                            </div>

                            <Button @click="saveConfig" label="Save Advanced Configuration" icon="pi pi-cog" />
                        </div>
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>

        <Toast />
        <ConfirmDialog />
    </div>
</template>

<script setup>
 import { ref, onMounted } from 'vue';
 import axios from 'axios';
 import Button from 'primevue/button';
 import Password from 'primevue/password';
 import Tabs from 'primevue/tabs';
 import TabList from 'primevue/tablist';
 import Tab from 'primevue/tab';
 import TabPanels from 'primevue/tabpanels';
 import TabPanel from 'primevue/tabpanel';
 import Dropdown from 'primevue/dropdown';
 import InputNumber from 'primevue/inputnumber';
 import InputText from 'primevue/inputtext';
 import ToggleSwitch from 'primevue/toggleswitch';
 import Chips from 'primevue/chips';
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

 const passwordForm = ref({
     currentPassword: '',
     newPassword: ''
 });

 const config = ref({
     log_level: 'INFO',
     log_max_size: 100,
     log_max_files: 10,
     log_max_age_days: 30,
     tls_enabled: false,
     tls_cert_file: '',
     tls_key_file: '',
     session_timeout: 24,
     cors_allowed_origins: ['*'],
     cors_allowed_methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
     cors_allowed_headers: ['Content-Type', 'Authorization'],
     rate_limit_enabled: true,
     rate_limit_requests: 60,
     rate_limit_burst: 100,
     ip_whitelist_enabled: false,
     ip_whitelist: [],
     ip_blacklist_enabled: false,
     ip_blacklist: [],
     email_enabled: false,
     email_smtp_server: '',
     email_smtp_port: 587,
     email_from: '',
     email_to: '',
     email_username: '',
     email_password: '',
     email_alert_on_error: true,
     email_alert_on_warning: false,
     backup_enabled: true,
     backup_interval: 'daily',
     backup_retention: 7,
     backup_path: './backups',
     backup_database: true,
     backup_config: true,
     metrics_enabled: true,
     metrics_port: ':9090',
     debug_mode: false,
     max_connections: 1000
 });

 const logLevels = [
     { label: 'DEBUG', value: 'DEBUG' },
     { label: 'INFO', value: 'INFO' },
     { label: 'WARN', value: 'WARN' },
     { label: 'ERROR', value: 'ERROR' }
 ];

 const backupIntervals = [
     { label: 'Hourly', value: 'hourly' },
     { label: 'Daily', value: 'daily' },
     { label: 'Weekly', value: 'weekly' }
 ];

 const importFile = ref(null);

 const fetchConfig = async () => {
     try {
         const res = await axios.get('/api/config/system');
         config.value = { ...config.value, ...res.data };
     } catch (e) {
         console.error('Failed to fetch config', e);
     }
 };

 onMounted(async () => {
     await Promise.all([
         store.fetchWebPort(),
         store.fetchProxies(),
         fetchConfig()
     ]);
     loading.value = false;
 });

 const saveConfig = async () => {
     try {
         await axios.put('/api/config/system', config.value);
         toast.add({ severity: 'success', summary: 'Success', detail: 'Configuration saved', life: 3000 });
     } catch (e) {
         toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data || e.message, life: 5000 });
     }
 };

 const changePassword = async () => {
     try {
         await axios.post('/api/config/password', passwordForm.value);
         toast.add({ severity: 'success', summary: 'Success', detail: 'Password changed successfully', life: 3000 });
         passwordForm.value = { currentPassword: '', newPassword: '' };
     } catch (e) {
         let errorMsg = e.response?.data || e.message;
         // Provide user-friendly error messages for common password validation errors
         if (typeof errorMsg === 'string') {
             if (errorMsg.includes('at least 8 characters')) {
                 errorMsg = 'Das Passwort muss mindestens 8 Zeichen lang sein';
             } else if (errorMsg.includes('at least 3 of')) {
                 errorMsg = 'Das Passwort muss mindestens 3 dieser Zeichenarten enthalten: Großbuchstaben, Kleinbuchstaben, Zahlen, Sonderzeichen';
             } else if (errorMsg.includes('too common')) {
                 errorMsg = 'Das Passwort ist zu einfach oder häufig verwendet';
             }
         }
         toast.add({ severity: 'error', summary: 'Error', detail: errorMsg, life: 5000 });
     }
 };

 const exportConfig = async () => {
     try {
         const res = await axios.get('/api/config/export', { responseType: 'blob' });
         const url = window.URL.createObjectURL(new Blob([res.data]));
         const link = document.createElement('a');
         link.href = url;
         link.setAttribute('download', 'config.json');
         document.body.appendChild(link);
         link.click();
         link.remove();
         toast.add({ severity: 'success', summary: 'Success', detail: 'Configuration exported', life: 3000 });
     } catch (e) {
         toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to export configuration', life: 5000 });
     }
 };

 const triggerImport = () => {
     importFile.value.click();
 };

 const importConfig = async (event) => {
     const file = event.target.files[0];
     if (!file) return;

     try {
         const formData = new FormData();
         formData.append('file', file);
         await axios.post('/api/config/import', formData, {
             headers: { 'Content-Type': 'multipart/form-data' }
         });
         toast.add({ severity: 'success', summary: 'Success', detail: 'Configuration imported', life: 3000 });
         await store.fetchProxies();
     } catch (e) {
         toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to import configuration', life: 5000 });
     }
     event.target.value = '';
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
