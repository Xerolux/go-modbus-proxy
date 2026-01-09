<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import Card from 'primevue/card';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import Message from 'primevue/message';

const password = ref('');
const error = ref('');
const auth = useAuthStore();
const router = useRouter();
const loading = ref(false);

const handleLogin = async () => {
    loading.value = true;
    error.value = '';
    const success = await auth.login(password.value);
    loading.value = false;
    if (success) {
        router.push('/');
    } else {
        error.value = 'Invalid password';
    }
};
</script>

<template>
    <div class="flex items-center justify-center min-h-screen bg-gray-900">
        <Card class="w-full max-w-md bg-gray-800 border-gray-700 text-white">
            <template #title>Login</template>
            <template #content>
                <div class="flex flex-col gap-4">
                    <div class="flex flex-col gap-2">
                        <label for="password">Password</label>
                        <InputText id="password" v-model="password" type="password" @keyup.enter="handleLogin" />
                    </div>
                    <Message v-if="error" severity="error">{{ error }}</Message>
                    <Button label="Login" @click="handleLogin" :loading="loading" />
                </div>
            </template>
        </Card>
    </div>
</template>

<style scoped>
:deep(.p-card) {
    background: #1f2937;
    color: white;
}
:deep(.p-inputtext) {
    background: #374151;
    border-color: #4b5563;
    color: white;
}
</style>
