<template>
  <div class="p-4 flex flex-col gap-4">
    <div class="flex justify-between items-center mb-4">
      <h1 class="text-2xl font-bold">Device Management</h1>
      <div class="flex gap-2">
        <Button
          label="Export CSV"
          icon="pi pi-download"
          severity="secondary"
          @click="exportDevices"
        />
        <Button
          label="Refresh"
          icon="pi pi-refresh"
          @click="fetchDevices"
        />
      </div>
    </div>

    <div v-if="loading" class="flex justify-center min-h-[500px]">
      <i class="pi pi-spin pi-spinner text-4xl"></i>
    </div>

    <div v-else-if="error" class="flex justify-center min-h-[500px]">
      <div class="text-center">
        <i class="pi pi-exclamation-triangle text-4xl text-red-500"></i>
        <p class="mt-4 text-red-400">Fehler: {{ error }}</p>
        <Button @click="fetchDevices" label="Erneut versuchen" class="mt-4" />
      </div>
    </div>

    <div v-else class="flex flex-col gap-4">
      <div class="flex gap-4 items-center">
        <InputText
          v-model="searchTerm"
          placeholder="Geräte durchsuchen..."
          class="w-full max-w-md"
        />
        <Dropdown
          v-model="selectedSort"
          :options="sortOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Sortierung"
          class="w-48"
        />
      </div>

      <DataTable
        :value="filteredDevices"
        :paginator="true"
        :rows="25"
        :rowsPerPageOptions="[10, 25, 50, 100]"
        :globalFilterFields="['name', 'ip', 'firstSeen']"
        :filters="filters"
        filterDisplay="row"
        responsiveLayout="scroll"
        stripedRows
        class="p-datatable-sm"
      >
        <Column field="ip" header="IP-Adresse" sortable></Column>
        <Column field="name" header="Name" sortable :filterMatchMode="FilterMatchMode.CONTAINS">
          <template #body="{ data }">
            <InputText
              v-model="data.name"
              @change="updateDeviceName(data)"
              class="w-full"
            />
          </template>
        </Column>
        <Column field="mac" header="MAC-Adresse" sortable></Column>
        <Column field="firstSeen" header="Erstmals gesehen" sortable>
          <template #body="{ data }">
            {{ formatDate(data.firstSeen) }}
          </template>
        </Column>
        <Column field="connectionCount" header="Verbindungen" sortable>
          <template #body="{ data }">
            <Badge :value="data.connectionCount" :severity="getConnectionSeverity(data.connectionCount)" />
          </template>
        </Column>
        <Column header="Aktionen" :exportable="false">
          <template #body="{ data }">
            <div class="flex gap-2">
              <Button
                icon="pi pi-eye"
                size="small"
                text
                @click="showDeviceDetails(data)"
              />
              <Button
                icon="pi pi-history"
                size="small"
                text
                @click="showConnectionHistory(data.ip)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <Toast />
    <Dialog v-model:visible="deviceDetailsVisible" header="Geräte-Details" :style="{ width: '50vw' }" modal>
      <div v-if="selectedDevice" class="flex flex-col gap-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="font-semibold">IP-Adresse:</label>
            <p>{{ selectedDevice.ip }}</p>
          </div>
          <div>
            <label class="font-semibold">MAC-Adresse:</label>
            <p>{{ selectedDevice.mac || 'N/A' }}</p>
          </div>
          <div>
            <label class="font-semibold">Name:</label>
            <p>{{ selectedDevice.name || 'N/A' }}</p>
          </div>
          <div>
            <label class="font-semibold">Erstmals gesehen:</label>
            <p>{{ formatDate(selectedDevice.firstSeen) }}</p>
          </div>
          <div>
            <label class="font-semibold">Letzte Verbindung:</label>
            <p>{{ selectedDevice.lastSeen ? formatDate(selectedDevice.lastSeen) : 'N/A' }}</p>
          </div>
          <div>
            <label class="font-semibold">Gesamtverbindungen:</label>
            <p>{{ selectedDevice.connectionCount }}</p>
          </div>
        </div>
      </div>
    </Dialog>

    <Dialog v-model:visible="historyVisible" header="Verbindungshistorie" :style="{ width: '70vw', maxWidth: '1000px' }" modal>
      <div class="flex flex-col gap-4">
        <div class="flex justify-between items-center">
          <h3>{{ selectedDevice?.ip }}</h3>
          <div class="flex gap-2">
            <Button
              label="Export CSV"
              icon="pi pi-download"
              severity="secondary"
              size="small"
              @click="exportHistoryCSV"
            />
          </div>
        </div>
        <DataTable
          :value="connectionHistory"
          :paginator="true"
          :rows="10"
          :rowsPerPageOptions="[10, 25, 50]"
          stripedRows
          class="p-datatable-sm"
        >
          <Column field="proxyID" header="Proxy ID" sortable></Column>
          <Column field="connectedAt" header="Verbunden am" sortable>
            <template #body="{ data }">
              {{ formatDateTime(data.connectedAt) }}
            </template>
          </Column>
          <Column field="requestCount" header="Anzahl Requests" sortable></Column>
        </DataTable>
      </div>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import axios from 'axios';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';
import Dropdown from 'primevue/dropdown';
import Badge from 'primevue/badge';
import Dialog from 'primevue/dialog';
import Toast from 'primevue/toast';
import { useToast } from 'primevue/usetoast';
import { useAppStore } from '../stores/appStore';

const store = useAppStore();
const toast = useToast();

const devices = ref([]);
const loading = ref(true);
const error = ref(null);
const searchTerm = ref('');
const selectedSort = ref({ label: 'Name (A-Z)', value: 'name_asc' });
const deviceDetailsVisible = ref(false);
const historyVisible = ref(false);
const selectedDevice = ref(null);
const connectionHistory = ref([]);

const sortOptions = [
  { label: 'Name (A-Z)', value: 'name_asc' },
  { label: 'Name (Z-A)', value: 'name_desc' },
  { label: 'IP (A-Z)', value: 'ip_asc' },
  { label: 'Verbindungen (Höchste)', value: 'connections_desc' },
  { label: 'Zuerst gesehen', value: 'firstSeen_desc' }
];

const filters = {
  'ip': { value: null, matchMode: 'contains' },
  'name': { value: null, matchMode: 'contains' },
  'mac': { value: null, matchMode: 'contains' },
  'firstSeen': { value: null, matchMode: 'date' }
};

const filteredDevices = computed(() => {
  let result = [...devices.value];

  if (searchTerm.value) {
    const search = searchTerm.value.toLowerCase();
    result = result.filter(device =>
      device.ip.toLowerCase().includes(search) ||
      device.name?.toLowerCase().includes(search) ||
      device.mac?.toLowerCase().includes(search)
    );
  }

  switch (selectedSort.value.value) {
    case 'name_asc':
      result.sort((a, b) => (a.name || '').localeCompare(b.name || ''));
      break;
    case 'name_desc':
      result.sort((a, b) => (b.name || '').localeCompare(a.name || ''));
      break;
    case 'ip_asc':
      result.sort((a, b) => a.ip.localeCompare(b.ip));
      break;
    case 'connections_desc':
      result.sort((a, b) => b.connectionCount - a.connectionCount);
      break;
    case 'firstSeen_desc':
      result.sort((a, b) => new Date(b.firstSeen) - new Date(a.firstSeen));
      break;
  }

  return result;
});

const fetchDevices = async () => {
  loading.value = true;
  error.value = null;
  try {
    const res = await axios.get('/api/devices');
    devices.value = res.data.map(device => ({
      ...device,
      connectionCount: 1 // Default value if not provided
    }));
    loading.value = false;
  } catch (e) {
    error.value = e.response?.data?.error || e.message;
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to fetch devices', life: 3000 });
    loading.value = false;
  }
};

const updateDeviceName = async (device) => {
  try {
    await axios.put('/api/devices', {
      ip: device.ip,
      name: device.name
    });
    toast.add({ severity: 'success', summary: 'Success', detail: 'Device name updated', life: 3000 });
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to update device name', life: 3000 });
  }
};

const showDeviceDetails = (device) => {
  selectedDevice.value = device;
  deviceDetailsVisible.value = true;
};

const showConnectionHistory = async (ip) => {
  selectedDevice.value = devices.value.find(d => d.ip === ip);
  loading.value = true;
  try {
    const res = await axios.get(`/api/devices/history?device_ip=${ip}`);
    connectionHistory.value = res.data;
    historyVisible.value = true;
    loading.value = false;
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to fetch history', life: 3000 });
    loading.value = false;
  }
};

const exportHistoryCSV = async () => {
  if (selectedDevice.value) {
    try {
      await store.exportDeviceHistory('csv');
      toast.add({ severity: 'success', summary: 'Success', detail: 'History exported', life: 3000 });
    } catch (e) {
      toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to export', life: 3000 });
    }
  }
};

const exportDevices = async () => {
  try {
    await store.exportDeviceHistory('csv');
    toast.add({ severity: 'success', summary: 'Success', detail: 'Devices exported', life: 3000 });
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to export', life: 3000 });
  }
};

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A';
  const date = new Date(dateStr);
  return date.toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  });
};

const formatDateTime = (dateStr) => {
  if (!dateStr) return 'N/A';
  const date = new Date(dateStr);
  return date.toLocaleString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

const getConnectionSeverity = (count) => {
  if (count > 100) return 'danger';
  if (count > 50) return 'warning';
  return 'success';
};

onMounted(() => {
  fetchDevices();
});
</script>
