<template>
  <div
    v-if="visible"
    :class="['proxy-config-panel', minimized ? 'minimized' : '']"
    :style="{ left: position.x + 'px', top: position.y + 'px' }"
    @mousedown="startDrag"
  >
    <!-- Header with drag handle -->
    <div class="panel-header" @mousedown="startDrag">
      <div class="header-content">
        <i class="pi pi-cog text-lg animate-spin-slow"></i>
        <span class="font-semibold">Proxy-Konfiguration</span>
      </div>
      <div class="header-actions">
        <button
          @click="toggleMinimize"
          class="p-1 hover:bg-white/10 rounded transition-colors"
          :title="minimized ? 'Erweitern' : 'Minimieren'"
        >
          <i :class="minimized ? 'pi pi-chevron-down' : 'pi pi-chevron-up'"></i>
        </button>
        <button
          @click="close"
          class="p-1 hover:bg-red-500/30 rounded transition-colors text-red-400"
          title="Schließen"
        >
          <i class="pi pi-times"></i>
        </button>
      </div>
    </div>

    <!-- Content -->
    <div v-if="!minimized" class="panel-content">
      <!-- Proxy Selection -->
      <div class="section">
        <h3 class="section-title">Proxies auswählen</h3>
        <div class="proxy-list">
          <div
            v-for="proxy in proxies"
            :key="proxy.id"
            :class="['proxy-item', { selected: selectedProxies.includes(proxy.id) }]"
            @click="toggleProxySelection(proxy.id)"
          >
            <div class="proxy-info">
              <i class="pi" :class="getStatusIcon(proxy.status)"></i>
              <span class="proxy-name">{{ proxy.name }}</span>
            </div>
            <div class="proxy-status">
              <Tag :severity="getSeverity(proxy.status)" :value="proxy.status" class="text-xs" />
            </div>
          </div>
        </div>
      </div>

      <!-- Batch Configuration -->
      <div v-if="selectedProxies.length > 0" class="section">
        <h3 class="section-title">
          Batch-Konfiguration
          <span class="text-sm font-normal text-gray-400">({{ selectedProxies.length }} ausgewählt)</span>
        </h3>

        <div class="config-form">
          <!-- Connection Settings -->
          <div class="config-group">
            <label class="config-label">Verbindungstimeout (s)</label>
            <InputNumber
              v-model="batchConfig.connection_timeout"
              :min="1"
              :max="300"
              class="w-full"
              showButtons
            />
          </div>

          <div class="config-group">
            <label class="config-label">Read Timeout (s)</label>
            <InputNumber
              v-model="batchConfig.read_timeout"
              :min="1"
              :max="600"
              class="w-full"
              showButtons
            />
          </div>

          <div class="config-group">
            <label class="config-label">Max Retries</label>
            <InputNumber
              v-model="batchConfig.max_retries"
              :min="0"
              :max="10"
              class="w-full"
              showButtons
            />
          </div>

          <div class="config-group">
            <label class="config-label">Max Read Size (0=unbegrenzt)</label>
            <InputNumber
              v-model="batchConfig.max_read_size"
              :min="0"
              :max="1000"
              class="w-full"
              showButtons
            />
          </div>

          <div class="config-group checkbox-group">
            <div class="flex items-center gap-2">
              <Checkbox v-model="batchConfig.enabled" binary />
              <span class="text-sm">Aktiviert</span>
            </div>
          </div>

          <div class="config-group checkbox-group">
            <div class="flex items-center gap-2">
              <Checkbox v-model="batchConfig.paused" binary />
              <span class="text-sm">Pausiert</span>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="actions">
          <Button
            label="Konfiguration anwenden"
            icon="pi pi-check"
            @click="applyBatchConfig"
            :loading="applying"
            class="flex-1"
          />
          <Button
            label="Alle starten"
            icon="pi pi-play"
            severity="success"
            @click="batchAction('start_all')"
            :loading="applying"
          />
          <Button
            label="Alle stoppen"
            icon="pi pi-stop"
            severity="danger"
            @click="batchAction('stop_all')"
            :loading="applying"
          />
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <i class="pi pi-arrow-left text-4xl text-gray-500 mb-3"></i>
        <p class="text-gray-400">Wähle Proxies aus, um sie zu konfigurieren</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import InputNumber from 'primevue/inputnumber';
import Checkbox from 'primevue/checkbox';
import Button from 'primevue/button';
import Tag from 'primevue/tag';
import axios from 'axios';

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  proxies: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['close', 'refresh']);

const minimized = ref(false);
const position = reactive({ x: window.innerWidth - 420, y: 100 });
const dragging = ref(false);
const dragOffset = reactive({ x: 0, y: 0 });
const selectedProxies = ref([]);
const applying = ref(false);

const batchConfig = reactive({
  connection_timeout: 10,
  read_timeout: 30,
  max_retries: 3,
  max_read_size: 0,
  enabled: true,
  paused: false
});

const startDrag = (e) => {
  if (e.target.closest('button') || e.target.closest('input')) return;
  dragging.value = true;
  dragOffset.x = e.clientX - position.x;
  dragOffset.y = e.clientY - position.y;
  document.addEventListener('mousemove', onDrag);
  document.addEventListener('mouseup', stopDrag);
};

const onDrag = (e) => {
  if (!dragging.value) return;

  let newX = e.clientX - dragOffset.x;
  let newY = e.clientY - dragOffset.y;

  // Boundary constraints
  newX = Math.max(0, Math.min(newX, window.innerWidth - 400));
  newY = Math.max(0, Math.min(newY, window.innerHeight - 200));

  position.x = newX;
  position.y = newY;
};

const stopDrag = () => {
  dragging.value = false;
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
};

const toggleMinimize = () => {
  minimized.value = !minimized.value;
};

const close = () => {
  emit('close');
};

const toggleProxySelection = (proxyId) => {
  const index = selectedProxies.value.indexOf(proxyId);
  if (index > -1) {
    selectedProxies.value.splice(index, 1);
  } else {
    selectedProxies.value.push(proxyId);
  }
};

const getStatusIcon = (status) => {
  switch (status) {
    case 'Running': return 'pi-circle-fill text-green-400';
    case 'Stopped': return 'pi-circle text-gray-400';
    case 'Error': return 'pi-circle-fill text-red-400';
    default: return 'pi-circle text-yellow-400';
  }
};

const getSeverity = (status) => {
  switch (status) {
    case 'Running': return 'success';
    case 'Stopped': return 'secondary';
    case 'Error': return 'danger';
    default: return 'info';
  }
};

const applyBatchConfig = async () => {
  if (selectedProxies.value.length === 0) return;

  applying.value = true;
  try {
    const promises = selectedProxies.value.map(proxyId => {
      const proxy = props.proxies.find(p => p.id === proxyId);
      if (!proxy) return Promise.resolve();

      const updatedProxy = {
        ...proxy,
        connection_timeout: batchConfig.connection_timeout,
        read_timeout: batchConfig.read_timeout,
        max_retries: batchConfig.max_retries,
        max_read_size: batchConfig.max_read_size,
        enabled: batchConfig.enabled,
        paused: batchConfig.paused
      };

      return axios.put('/api/proxies', updatedProxy);
    });

    await Promise.all(promises);
    emit('refresh');
    selectedProxies.value = [];
  } catch (error) {
    console.error('Failed to apply batch config:', error);
  } finally {
    applying.value = false;
  }
};

const batchAction = async (action) => {
  applying.value = true;
  try {
    await axios.post('/api/proxies/control', { action });
    emit('refresh');
  } catch (error) {
    console.error('Failed to execute batch action:', error);
  } finally {
    applying.value = false;
  }
};

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
});
</script>

<style scoped>
.proxy-config-panel {
  position: fixed;
  width: 400px;
  max-height: 80vh;
  background: rgba(31, 41, 55, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(75, 85, 99, 0.5);
  border-radius: 16px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 1000;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.proxy-config-panel.minimized {
  height: auto;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: rgba(59, 130, 246, 0.1);
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
  cursor: move;
  user-select: none;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.panel-content {
  padding: 16px;
  overflow-y: auto;
  max-height: calc(80vh - 60px);
}

.panel-content::-webkit-scrollbar {
  width: 6px;
}

.panel-content::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.panel-content::-webkit-scrollbar-thumb {
  background: rgba(75, 85, 99, 0.5);
  border-radius: 3px;
}

.panel-content::-webkit-scrollbar-thumb:hover {
  background: rgba(75, 85, 99, 0.7);
}

.section {
  margin-bottom: 20px;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
  margin-bottom: 12px;
}

.proxy-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.proxy-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: rgba(17, 24, 39, 0.5);
  border: 1px solid rgba(75, 85, 99, 0.3);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.proxy-item:hover {
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(59, 130, 246, 0.3);
  transform: translateX(4px);
}

.proxy-item.selected {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.5);
}

.proxy-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.proxy-name {
  font-size: 14px;
  font-weight: 500;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.config-group.checkbox-group {
  flex-direction: row;
}

.config-label {
  font-size: 12px;
  font-weight: 500;
  color: #9ca3af;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.animate-spin-slow {
  animation: spin-slow 3s linear infinite;
}
</style>
