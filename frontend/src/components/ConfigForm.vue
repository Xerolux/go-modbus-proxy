<template>
  <div class="space-y-6">
    <!-- Global Config -->
    <div class="grid grid-cols-1 gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-300 dark:text-gray-200">Web Interface Address</label>
        <div class="flex space-x-2 mt-1">
             <input
               v-model="store.webPort"
               type="text"
               class="block w-full rounded-md border-gray-600 bg-gray-700 text-white shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
               placeholder=":8080"
             >
            <button
                @click="savePort"
                class="px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
            >
                Save
            </button>
        </div>
      </div>
    </div>

    <div class="border-t border-gray-700 pt-4">
      <div class="flex justify-between items-center mb-4">
        <h4 class="text-md font-medium text-gray-200">Proxies</h4>
        <button
          @click="addProxy"
          class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
        >
          Add Proxy
        </button>
      </div>

      <VueDraggable
        v-model="store.proxies"
        :animation="150"
        handle=".proxy-handle"
        class="space-y-4"
        @end="onReorder"
      >
        <div
          v-for="(proxy, index) in store.proxies"
          :key="proxy.id || index"
          class="border border-gray-600 rounded-md p-4 bg-gray-700 relative group"
          :class="{'border-blue-500 ring-1 ring-blue-500': proxy._isNew || proxy._isDirty}"
        >
          <div class="absolute right-2 top-2 flex space-x-2">
            <button
               @click="removeProxy(proxy.id)"
               class="text-red-400 hover:text-red-300 p-1"
               title="Remove Proxy"
            >
              <TrashIcon class="w-4 h-4" />
            </button>
             <div class="cursor-move proxy-handle text-gray-400 hover:text-gray-300 p-1">
              <GripVerticalIcon class="w-4 h-4" />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pr-8">
            <div>
              <label class="block text-xs font-medium text-gray-400">Name</label>
              <input v-model="proxy.name" @input="markDirty(proxy)" type="text" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
            </div>
             <div>
              <label class="block text-xs font-medium text-gray-400">Listen Addr</label>
              <input v-model="proxy.listen_addr" @input="markDirty(proxy)" type="text" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border" placeholder=":502">
            </div>
             <div>
              <label class="block text-xs font-medium text-gray-400">Target Addr</label>
              <input v-model="proxy.target_addr" @input="markDirty(proxy)" type="text" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border" placeholder="192.168.1.100:502">
            </div>
             <div>
                <label class="block text-xs font-medium text-gray-400">Description</label>
                <input v-model="proxy.description" @input="markDirty(proxy)" type="text" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
             </div>
              <div class="flex items-center space-x-4 mt-4">
                 <label class="inline-flex items-center">
                   <input type="checkbox" v-model="proxy.enabled" @change="markDirty(proxy)" class="rounded border-gray-600 text-blue-600 bg-gray-700 shadow-sm focus:border-blue-500 focus:ring focus:ring-blue-200 focus:ring-opacity-50">
                   <span class="ml-2 text-sm text-gray-300">Enabled</span>
                </label>
                 <label class="inline-flex items-center">
                  <input type="checkbox" v-model="proxy.paused" @change="markDirty(proxy)" class="rounded border-gray-600 text-blue-600 bg-gray-700 shadow-sm focus:border-blue-500 focus:ring focus:ring-blue-200 focus:ring-opacity-50">
                  <span class="ml-2 text-sm text-gray-300">Paused</span>
                </label>
             </div>
             <div class="mt-2">
                 <label class="block text-xs font-medium text-gray-400">Tags</label>
                 <input v-model="proxy.tags" @input="markDirty(proxy)" type="text" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border" placeholder="comma,separated,tags">
             </div>
          </div>

           <div class="mt-2 flex justify-between items-center">
                <div class="text-xs text-gray-400 cursor-pointer hover:text-gray-300" @click="proxy._showAdvanced = !proxy._showAdvanced">
                    {{ proxy._showAdvanced ? 'Hide Advanced' : 'Show Advanced' }}
                </div>
                <button
                    v-if="proxy._isDirty || proxy._isNew"
                    @click="saveProxy(proxy)"
                    class="px-3 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
                >
                    {{ proxy._isNew ? 'Create' : 'Update' }}
                </button>
           </div>

           <div v-if="proxy._showAdvanced" class="mt-2 grid grid-cols-1 md:grid-cols-3 gap-4 border-t border-gray-600 pt-2">
              <div>
                <label class="block text-xs font-medium text-gray-400">Conn Timeout (s)</label>
                <input v-model.number="proxy.connection_timeout" @input="markDirty(proxy)" type="number" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-400">Read Timeout (s)</label>
                <input v-model.number="proxy.read_timeout" @input="markDirty(proxy)" type="number" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-400">Max Retries</label>
                <input v-model.number="proxy.max_retries" @input="markDirty(proxy)" type="number" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-400">Max Read Size (0=unlimited)</label>
                <input v-model.number="proxy.max_read_size" @input="markDirty(proxy)" type="number" class="mt-1 block w-full rounded border-gray-600 bg-gray-600 text-white sm:text-sm p-1 border">
              </div>
           </div>

        </div>
      </VueDraggable>
    </div>
  </div>
</template>

<script setup>
 import { useAppStore } from '../stores/appStore';
 import { VueDraggable } from 'vue-draggable-plus';
 import { TrashIcon, GripVerticalIcon } from 'lucide-vue-next';
 import validators from '../utils/validators';

 const store = useAppStore();

 const validationErrors = ref({});

 const addProxy = () => {
   store.proxies.push({
     id: '',
     name: 'New Proxy',
     listen_addr: ':5020',
     target_addr: '127.0.0.1:502',
     enabled: true,
     paused: false,
     connection_timeout: 10,
     read_timeout: 30,
     max_retries: 3,
     max_read_size: 0,
     description: '',
     tags: [],
     _isNew: true,
     _showAdvanced: false
   });
 };

 const removeProxy = async (id) => {
   if (!id) {
     // Removing an unsaved new proxy
     const idx = store.proxies.findIndex(p => !p.id);
     if (idx !== -1) store.proxies.splice(idx, 1);
     return;
     }
   if (confirm('Are you sure you want to remove this proxy?')) {
     await store.deleteProxy(id);
     }
 };

 const saveProxy = async (proxy) => {
   // Validate before saving
   const errors = validators.validateProxyConfig(proxy);

   if (errors) {
     validationErrors.value = errors;
     return;
   }

   validationErrors.value = {};

   let success = false;
   if (proxy._isNew) {
     success = await store.addProxy(proxy);
   } else {
     success = await store.updateProxy(proxy);
   }
   if (success) {
     // Refresh handled by store, but we might want to clear dirty flags if we kept the object
     // Store re-fetch replaces the list, so flags are gone.
   }
 };

 const savePort = async () => {
   if (confirm('Changing the port requires a system restart. Continue?')) {
     await store.saveWebPort(store.webPort);
     }
 };

 const markDirty = (proxy) => {
   if (!proxy._isNew) {
     proxy._isDirty = true;
     // Clear validation error on input
     if (validationErrors.value[proxy.id]) {
       delete validationErrors.value[proxy.id];
     }
   }
 };

 const validateField = (proxy, field, value) => {
   const errors = validators.validateProxyConfig({ ...proxy, [field]: value });

   if (errors && errors[field]) {
     validationErrors.value[proxy.id || proxy._tempId] = {
       ...validationErrors.value[proxy.id || proxy._tempId],
       [field]: errors[field]
     };
   } else if (validationErrors.value[proxy.id || proxy._tempId]) {
     delete validationErrors.value[proxy.id || proxy._tempId][field];
   }
 };

 const onReorder = () => {
   // Backend doesn't support reordering yet (ID based list), but visually it works.
 };
</script>
