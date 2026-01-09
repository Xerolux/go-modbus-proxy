<template>
    <div class="p-4 flex flex-col h-full">
         <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">Dashboard</h1>
            <div class="flex gap-2">
                 <Button label="Reset Layout" icon="pi pi-refresh" @click="resetLayout" severity="secondary" />
                 <Button label="Add Widget" icon="pi pi-plus" @click="openAddWidget" />
            </div>
        </div>

        <div class="grid-stack bg-gray-800 rounded-lg min-h-[500px]"></div>

        <Teleport v-for="widget in widgets" :key="widget.id" :to="'#mount_' + widget.id">
            <div class="relative h-full w-full flex flex-col justify-between bg-gray-700 rounded shadow-md">
                <DashboardWidget
                    :title="widget.title"
                    :value="getWidgetValue(widget)"
                    :unit="widget.unit"
                />
                 <div class="absolute top-1 right-1 cursor-pointer text-gray-500 hover:text-red-500 z-10" @click="removeWidget(widget.id)">
                    <i class="pi pi-times"></i>
                </div>
            </div>
        </Teleport>

        <Dialog v-model:visible="showAddWidget" header="Add Widget" :modal="true">
            <div class="flex flex-col gap-4 min-w-[300px]">
                <label>Select Proxy</label>
                <Dropdown v-model="selectedProxy" :options="proxyOptions" optionLabel="name" placeholder="Select a Proxy" filter />
                <Button label="Add" @click="confirmAddWidget" :disabled="!selectedProxy" />
            </div>
        </Dialog>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import { GridStack } from 'gridstack';
import 'gridstack/dist/gridstack.min.css';
import axios from 'axios';
import Button from 'primevue/button';
import Dialog from 'primevue/dialog';
import Dropdown from 'primevue/dropdown';
import DashboardWidget from '../components/DashboardWidget.vue';

const grid = ref(null);
const proxies = ref([]);
const widgets = ref([]);
const showAddWidget = ref(false);
const selectedProxy = ref(null);
const proxyOptions = ref([]);
const timer = ref(null);

onMounted(async () => {
    await fetchData();

    grid.value = GridStack.init({
        float: true,
        cellHeight: 100,
        minRow: 1,
        margin: 5,
        column: 6,
        disableOneColumnMode: true
    });

    const savedLayout = localStorage.getItem('dashboard_layout');
    let layoutToLoad = [];

    // Default layout: one widget for each proxy
    const defaults = proxies.value.map((p, index) => ({
        x: (index % 3) * 2,
        y: Math.floor(index / 3) * 2,
        w: 2,
        h: 2,
        id: `w_default_${p.id}`,
        proxy_id: p.id,
        title: p.name,
        unit: 'req'
    }));

    if (savedLayout) {
        try {
            layoutToLoad = JSON.parse(savedLayout);
        } catch (e) { console.error(e) }
    }

    if (!layoutToLoad || layoutToLoad.length === 0) {
        layoutToLoad = defaults;
    }

    loadGrid(layoutToLoad);

    grid.value.on('change', (event, items) => {
        saveLayout();
    });

    grid.value.on('dragstop', (event, element) => {
            saveLayout();
    });

    grid.value.on('resizestop', (event, element) => {
            saveLayout();
    });

    timer.value = setInterval(fetchData, 2000);
});

onUnmounted(() => {
        if (timer.value) clearInterval(timer.value);
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

    // Create widget without content first to avoid escaping issues
    const el = grid.value.addWidget({
        x: item.x,
        y: item.y,
        w: item.w,
        h: item.h,
        id: id,
        content: ''
    });

    // Manually set innerHTML of the content div
    if (el) {
        const contentEl = el.querySelector('.grid-stack-item-content');
        if (contentEl) {
            contentEl.innerHTML = contentHtml;
        }
    }

    // Ensure DOM is ready before Vue tries to Teleport
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
                unit: w ? w.unit : ''
            };
        });
        localStorage.setItem('dashboard_layout', JSON.stringify(layout));
};

const fetchData = async () => {
    try {
        const res = await axios.get('/api/proxies');
        proxies.value = res.data;
        proxyOptions.value = res.data.map(p => ({ name: p.name, id: p.id }));
    } catch (e) {}
};

const getWidgetValue = (widget) => {
    const p = proxies.value.find(x => x.id === widget.proxy_id);
    if (!p) return 'Unknown';
    if (p.status === 'Running') {
        return `Running (${p.requests || 0})`;
    }
    return p.status;
}

const openAddWidget = () => {
    showAddWidget.value = true;
};

const confirmAddWidget = () => {
    if (selectedProxy.value) {
        addWidgetToGrid({
            x: 0, y: 0, w: 2, h: 2,
            proxy_id: selectedProxy.value.id,
            title: selectedProxy.value.name,
            unit: 'req'
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
}
</script>

<style>
.grid-stack-item-content {
    background-color: transparent !important;
    border-radius: 8px;
    overflow: hidden;
}
</style>
