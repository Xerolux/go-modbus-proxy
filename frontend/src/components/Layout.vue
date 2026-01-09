<script setup>
import { ref } from "vue";
import { useRouter } from 'vue-router'
import { useAuthStore } from "../stores/auth";
import Menubar from 'primevue/menubar';
import Button from 'primevue/button';

const router = useRouter();
const auth = useAuthStore();

const navigate = (path) => {
    router.push(path);
};

const items = ref([
    {
        label: 'Dashboard',
        icon: 'pi pi-home',
        command: () => router.push('/')
    },
    {
        label: 'Control',
        icon: 'pi pi-sliders-h',
        command: () => router.push('/control')
    },
    {
        label: 'Logs',
        icon: 'pi pi-list',
        command: () => router.push('/logs')
    },
    {
        label: 'Settings',
        icon: 'pi pi-cog',
        command: () => router.push('/config')
    }
]);

const logout = async () => {
    await auth.logout();
    router.push('/login');
}
</script>

<template>
    <div class="min-h-screen flex flex-col">
        <Menubar :model="items" class="rounded-none border-0 border-b border-gray-700 bg-gray-800">
             <template #start>
               <span class="text-xl font-bold px-4 text-white">ModBridge</span>
            </template>
            <template #item="{ item, props }">
                <a v-ripple class="flex items-center gap-2 px-3 py-2 hover:bg-gray-700 rounded cursor-pointer text-gray-200" v-bind="props.action">
                    <i :class="item.icon"></i>
                    <span>{{ item.label }}</span>
                </a>
            </template>
            <template #end>
                <div class="flex items-center gap-2">
                    <Button label="Logout" icon="pi pi-power-off" severity="danger" text @click="logout" />
                </div>
            </template>
        </Menubar>
        <main class="flex-grow bg-gray-900 text-white">
             <router-view></router-view>
        </main>
    </div>
</template>

<style>
.p-menubar {
    padding: 0.5rem;
}
</style>
