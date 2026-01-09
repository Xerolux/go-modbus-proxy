<template>
  <div class="min-h-screen bg-gray-100 p-8">
    <div class="max-w-4xl mx-auto space-y-6">

      <header class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Modbus Proxy</h1>
          <p class="text-gray-500">Modern Configuration Interface</p>
        </div>

        <div class="flex space-x-2">
            <button
              @click="restartSystem"
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-yellow-600 hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-yellow-500"
            >
              Restart System
            </button>
        </div>
      </header>

      <div v-if="store.error" class="bg-red-50 border-l-4 border-red-400 p-4 mb-4">
        <div class="flex">
          <div class="ml-3">
            <p class="text-sm text-red-700">
              {{ store.error }}
            </p>
          </div>
        </div>
      </div>

      <VueDraggable
        v-model="widgets"
        :animation="150"
        handle=".handle"
        class="space-y-6"
      >
         <component
            v-for="widget in widgets"
            :key="widget.id"
            :is="widget.component"
            :title="widget.title"
         >
            <template #default>
              <StatusWidget v-if="widget.id === 'status'" :status="store.status.setup_required ? 'Setup Required' : 'Running'" />
              <ConfigForm v-if="widget.id === 'config'" />
            </template>
         </component>
      </VueDraggable>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, markRaw } from 'vue';
import { useAppStore } from './stores/appStore';
import { VueDraggable } from 'vue-draggable-plus';
import Card from './components/Card.vue';
import StatusWidget from './components/StatusWidget.vue';
import ConfigForm from './components/ConfigForm.vue';

const store = useAppStore();

const widgets = ref([
  { id: 'status', title: 'System Status', component: markRaw(Card) },
  { id: 'config', title: 'Configuration', component: markRaw(Card) }
]);

const restartSystem = async () => {
    if (confirm('Are you sure you want to restart the system? connections will be dropped.')) {
        await store.restartSystem();
    }
};

onMounted(() => {
  store.fetchStatus();
  store.fetchProxies();
  store.fetchWebPort();

  // Poll status
  setInterval(() => {
    store.fetchStatus();
  }, 5000);
});
</script>
