<script setup>
  import { ref, onMounted, onUnmounted } from "vue";
  import { useRouter } from 'vue-router'
  import { useAuthStore } from "../stores/auth";
  import { useAppStore } from "../stores/appStore";
  import Menubar from 'primevue/menubar';
  import Button from 'primevue/button';
  import Sidebar from 'primevue/sidebar';
  import InputSwitch from 'primevue/inputswitch';

  const router = useRouter();
  const auth = useAuthStore();
  const appStore = useAppStore();

  const mobileMenuVisible = ref(false);
  const isMobile = ref(false);

  const navigate = (path) => {
      router.push(path);
      mobileMenuVisible.value = false;
  };

  const items = ref([
      {
          label: 'Dashboard',
          icon: 'pi pi-home',
          command: () => navigate('/')
      },
      {
          label: 'Control',
          icon: 'pi pi-sliders-h',
          command: () => navigate('/control')
      },
      {
          label: 'Logs',
          icon: 'pi pi-list',
          command: () => navigate('/logs')
      },
      {
          label: 'Settings',
          icon: 'pi pi-cog',
          command: () => navigate('/config')
      }
  ]);

  const logout = async () => {
      await auth.logout();
      router.push('/login');
  }

  const checkMobile = () => {
      isMobile.value = window.innerWidth < 768;
  };

  onMounted(() => {
      checkMobile();
      window.addEventListener('resize', checkMobile);
  });

  onUnmounted(() => {
      window.removeEventListener('resize', checkMobile);
  });
</script>

<template>
    <div class="min-h-screen flex flex-col">
        <Menubar :model="isMobile ? [] : items" class="rounded-none border-0 border-b border-gray-700 bg-gray-800">
             <template #start>
               <div class="flex items-center gap-4">
                   <Button v-if="isMobile" icon="pi pi-bars" text rounded @click="mobileMenuVisible = true" class="text-white" />
                   <span class="text-xl font-bold px-4 text-white">ModBridge</span>
               </div>
            </template>
            <template #item="{ item, props }">
                <a v-ripple class="flex items-center gap-2 px-3 py-2 hover:bg-gray-700 rounded cursor-pointer text-gray-200" v-bind="props.action">
                    <i :class="item.icon"></i>
                    <span>{{ item.label }}</span>
                </a>
            </template>
             <template #end>
                 <div class="flex items-center gap-2">
                     <div class="flex items-center gap-2 px-3">
                         <i :class="appStore.darkMode ? 'pi pi-moon' : 'pi pi-sun'"></i>
                         <InputSwitch v-model="appStore.darkMode" @change="appStore.toggleDarkMode" />
                     </div>
                     <Button label="Logout" icon="pi pi-power-off" severity="danger" text @click="logout" />
                 </div>
             </template>
        </Menubar>

        <Sidebar v-model:visible="mobileMenuVisible" :baseZIndex="10000">
            <div class="flex flex-col gap-2">
                <div v-for="item in items" :key="item.label">
                    <Button
                        @click="item.command"
                        :label="item.label"
                        :icon="item.icon"
                        text
                        class="w-full text-left"
                        size="large"
                    />
                </div>
                <div class="mt-4 pt-4 border-t border-gray-700">
                    <Button
                        @click="logout"
                        label="Logout"
                        icon="pi pi-power-off"
                        severity="danger"
                        text
                        class="w-full text-left"
                        size="large"
                    />
                </div>
            </div>
        </Sidebar>

        <main class="flex-grow bg-gray-900 text-white">
             <router-view></router-view>
        </main>
    </div>
</template>

<style>
.p-menubar {
    padding: 0.5rem;
}

.p-sidebar {
    background-color: #1f2937;
    color: white;
}

@media (max-width: 768px) {
    .p-menubar {
        padding: 0.5rem;
    }
}
</style>
