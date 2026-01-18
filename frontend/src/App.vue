<script setup>
import { onMounted } from 'vue';
import { useAppStore } from './stores/appStore';

const store = useAppStore();

onMounted(() => {
  // Apply theme on mount
  if (store.darkMode) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
});

// Watch for dark mode changes
import { watch } from 'vue';
watch(() => store.darkMode, (isDark) => {
  if (isDark) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
});
</script>

<template>
  <div :class="darkMode ? 'dark' : 'light'" class="min-h-screen transition-colors duration-300">
    <router-view></router-view>
  </div>
</template>

<style>
/* Light Theme (default) */
:root {
  --bg-primary: #1f2937;
  --bg-secondary: #374151;
  --bg-tertiary: #4b5563;
  --text-primary: #ffffff;
  --text-secondary: #9ca3af;
  --text-muted: #6b7280;
  --border-color: #374151;
  --accent-color: #3b82f6;
  --success-color: #10b981;
  --warning-color: #f59e0b;
  --danger-color: #ef4444;
}

/* Dark Theme */
.dark {
  --bg-primary: #0f172a;
  --bg-secondary: #1e293b;
  --bg-tertiary: #334155;
  --text-primary: #f1f5f9;
  --text-secondary: #cbd5e1;
  --text-muted: #94a3b8;
  --border-color: #1e293b;
  --accent-color: #60a5fa;
  --success-color: #34d399;
  --warning-color: #fbbf24;
  --danger-color: #f87171;
}

/* Light Theme (explicit) */
.light {
  --bg-primary: #f3f4f6;
  --bg-secondary: #e5e7eb;
  --bg-tertiary: #d1d5db;
  --text-primary: #111827;
  --text-secondary: #374151;
  --text-muted: #6b7280;
  --border-color: #e5e7eb;
  --accent-color: #2563eb;
  --success-color: #059669;
  --warning-color: #d97706;
  --danger-color: #dc2626;
}

body {
  background-color: var(--bg-primary);
  color: var(--text-primary);
  transition: background-color 0.3s ease, color 0.3s ease;
}
</style>
