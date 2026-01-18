<template>
    <div class="p-4 flex flex-col gap-4">
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center mb-4 lg:mb-6 gap-4 lg:gap-0">
             <h1 class="text-xl sm:text-2xl lg:text-3xl font-bold">Dashboard</h1>
             <div class="flex flex-col sm:flex-row gap-2 sm:gap-3 w-full lg:w-auto">
                  <Button label="Widget hinzufügen" icon="pi pi-plus" @click="openAddWidget" class="w-full sm:w-auto" />
                  <Button label="Layout zurücksetzen" icon="pi pi-refresh" @click="resetLayout" severity="secondary" class="w-full sm:w-auto" />
             </div>
        </div>

        <div v-if="loading" class="flex justify-center items-center min-h-[500px]">
             <div class="text-center">
                  <i class="pi pi-spin pi-spinner text-4xl text-blue-500"></i>
                  <p class="mt-4 text-gray-400">Lade Dashboard...</p>
             </div>
        </div>

        <div v-else-if="error" class="flex justify-center items-center min-h-[500px]">
             <div class="text-center">
                  <i class="pi pi-exclamation-triangle text-4xl text-red-500"></i>
                  <p class="mt-4 text-red-400">Fehler beim Laden: {{ error }}</p>
                  <Button @click="fetchData" label="Erneut versuchen" class="mt-4" />
             </div>
        </div>

        <div v-else class="grid-stack bg-gray-800/50 rounded-xl min-h-[500px] border border-gray-700/50">
            <Teleport v-for="widget in widgets" :key="widget.id" :to="'#mount_' + widget.id">
                <div class="relative h-full w-full p-3 flex flex-col justify-between animate-fade-in">
                    <DashboardWidget
                        :title="widget.title"
                        :value="getWidgetValue(widget)"
                        :unit="widget.unit"
                        :status="widget.status"
                    />
                     <button
                        type="button"
                        class="absolute top-2 right-2 cursor-pointer text-gray-400 hover:text-red-400 z-10 transition-colors p-1 rounded hover:bg-red-500/20 bg-gray-800 shadow-sm"
                        @click="removeWidget(widget.id)"
                        aria-label="Remove widget"
                        title="Widget entfernen"
                    >
                        <i class="pi pi-times text-sm"></i>
                    </button>
                </div>
            </Teleport>
        </div>

        <Dialog v-model:visible="showAddWidget" header="Widget hinzufügen" :modal="true" :style="{ width: '90vw', maxWidth: '400px' }">
            <div class="flex flex-col gap-4">
                <label class="text-sm font-medium">Proxy wählen</label>
                <Dropdown 
                    v-model="selectedProxy" 
                    :options="proxyOptions" 
                    optionLabel="name" 
                    optionValue="id"
                    placeholder="Wähle einen Proxy" 
                    filter 
                    class="w-full" 
                />
                <Button label="Hinzufügen" @click="confirmAddWidget" :disabled="!selectedProxy" class="w-full" />
            </div>
        </Dialog>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { GridStack } from 'gridstack';
import 'gridstack/dist/gridstack.min.css';
import axios from 'axios';
import Button from 'primevue/button';
import Dialog from 'primevue/dialog';
import Dropdown from 'primevue/dropdown';
import DashboardWidget from '../components/DashboardWidget.vue';
import { useEventSource } from '../utils/eventSource';

const grid = ref(null);
const proxies = ref([]);
const widgets = ref([]);
const showAddWidget = ref(false);
const selectedProxy = ref(null);
const proxyOptions = ref([]);
const loading = ref(true);
const error = ref(null);
let sseDisconnect = null;

onMounted(async () => {
    try {
        await fetchData();

        grid.value = GridStack.init({
            float: true,
            cellHeight: 80,
            minRow: 1,
            margin: 3,
            column: 6,
            disableOneColumnMode: false,
            oneColumnModeDomSort: true,
            oneColumnModeWidth: 640,
            breakpointForNColumn: {
                1: { width: 640, column: 1 },
                2: { width: 768, column: 2 },
                3: { width: 1024, column: 3 },
                4: { width: 1280, column: 4 },
                6: { width: 1536, column: 6 }
            }
        });

        const savedLayout = localStorage.getItem('dashboard_layout');
        let layoutToLoad = [];

        if (savedLayout) {
            try {
                layoutToLoad = JSON.parse(savedLayout);
            } catch (e) {
                console.error(e);
            }
        }

        if (!layoutToLoad || layoutToLoad.length === 0) {
            const defaults = proxies.value.map((p, index) => ({
                x: (index % 3) * 2,
                y: Math.floor(index / 3) * 2,
                w: 2,
                h: 2,
                id: `w_default_${p.id}`,
                proxy_id: p.id,
                title: p.name,
                unit: 'req',
                status: p.status
            }));
            layoutToLoad = defaults;
        }

        loadGrid(layoutToLoad);

        grid.value.on('change', saveLayout);
        grid.value.on('dragstop', saveLayout);
        grid.value.on('resizestop', saveLayout);

        const { data, disconnect } = useEventSource('/api/proxies/stream');
        sseDisconnect = disconnect;

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
                            proxyOptions.value.push({ name: proxyData.name, id: proxyData.id });
                        }
                    }
                    break;
                case 'proxy_removed':
                    if (eventData.proxy_id) {
                        proxies.value = proxies.value.filter(p => p.id !== eventData.proxy_id);
                        proxyOptions.value = proxyOptions.value.filter(p => p.id !== eventData.proxy_id);
                    }
                    break;
            }
        });

    } catch (err) {
        error.value = err.message;
        loading.value = false;
    }
});

onUnmounted(() => {
    if (grid.value) {
        grid.value.destroy();
    }
    if (sseDisconnect) {
        sseDisconnect();
    }
});

const loadGrid = (layout) => {
    grid.value.removeAll();
    widgets.value = [];

    layout.forEach(item => {
        addWidgetToGrid(item);
    });
};

const addWidgetToGrid = (item) => {
    const id = item.id || `w_${Date.now()}`;
    const contentHtml = `<div id="mount_${id}" class="h-full w-full relative"></div>`;

    const el = grid.value.addWidget({
        x: item.x,
        y: item.y,
        w: item.w,
        h: item.h,
        id: id,
        content: ''
    });

    if (el) {
        const contentEl = el.querySelector('.grid-stack-item-content');
        if (contentEl) {
            contentEl.innerHTML = contentHtml;
        }
    }

    nextTick(() => {
        widgets.value.push({
            ...item,
            id: id
        });
    });
};

const saveLayout = () => {
    const items = grid.value.getGridItems();
    const layout = items.map(item => {
        const w = widgets.value.find(x => x.id == item.gridstackNode.id);
        return {
            x: item.gridstackNode.x,
            y: item.gridstackNode.y,
            w: item.gridstackNode.w,
            h: item.gridstackNode.h,
            id: item.gridstackNode.id,
            proxy_id: w ? w.proxy_id : '',
            title: w ? w.title : '',
            unit: w ? w.unit : '',
            status: w ? w.status : ''
        };
    });
    localStorage.setItem('dashboard_layout', JSON.stringify(layout));
};

const fetchData = async () => {
    try {
        loading.value = true;
        error.value = null;
        const res = await axios.get('/api/proxies');
        proxies.value = res.data;
        proxyOptions.value = res.data.map(p => ({ name: p.name, id: p.id }));
        loading.value = false;
    } catch (e) {
        error.value = e.response?.data || e.message || 'Unbekannter Fehler';
        loading.value = false;
    }
};

const getWidgetValue = (widget) => {
    const p = proxies.value.find(x => x.id === widget.proxy_id);
    if (!p) return 'Unknown';
    if (p.status === 'Running') {
        return `${p.requests || 0} Anfragen`;
    }
    return p.status;
};

const openAddWidget = () => {
    showAddWidget.value = true;
};

const confirmAddWidget = () => {
    if (selectedProxy.value) {
        const p = proxies.value.find(x => x.id === selectedProxy.value);
        addWidgetToGrid({
            x: 0, y: 0, w: 2, h: 2,
            proxy_id: selectedProxy.value,
            title: p.name,
            unit: 'req',
            status: p.status
        });
        showAddWidget.value = false;
        selectedProxy.value = null;
        saveLayout();
    }
};

const removeWidget = (id) => {
    const el = document.getElementById(`mount_${id}`);
    if (el) {
        const gridItem = el.closest('.grid-stack-item');
        if (gridItem) {
            grid.value.removeWidget(gridItem);
        }
    }

    widgets.value = widgets.value.filter(w => w.id !== id);
    saveLayout();
};

const resetLayout = () => {
    localStorage.removeItem('dashboard_layout');
    location.reload();
};
</script>

<style>
.grid-stack-item-content {
    background-color: #1f2937;
    background-image: linear-gradient(to bottom right, rgba(255,255,255,0.05), rgba(255,255,255,0));
    color: white;
    border-radius: 16px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid rgba(75, 85, 99, 0.4);
}

.grid-stack-item:hover .grid-stack-item-content {
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    border-color: #3b82f6;
    transform: translateY(-4px);
    z-index: 10 !important;
}

.animate-fade-in {
    animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>
