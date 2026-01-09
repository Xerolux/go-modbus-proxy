import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useAppStore = defineStore('app', () => {
  const proxies = ref([]);
  const webPort = ref('');
  const status = ref({
    setup_required: false,
    proxies: []
  });

  const isLoading = ref(false);
  const error = ref(null);

  const fetchProxies = async () => {
    try {
      const res = await fetch('/api/proxies');
      if (!res.ok) throw new Error('Failed to fetch proxies');
      proxies.value = await res.json();
    } catch (e) {
      error.value = e.message;
    }
  };

  const addProxy = async (proxyData) => {
    isLoading.value = true;
    try {
      const res = await fetch('/api/proxies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(proxyData)
      });
      if (!res.ok) throw new Error('Failed to add proxy');
      await fetchProxies();
      return true;
    } catch (e) {
      error.value = e.message;
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const updateProxy = async (proxyData) => {
    isLoading.value = true;
    try {
      const res = await fetch('/api/proxies', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(proxyData)
      });
      if (!res.ok) throw new Error('Failed to update proxy');
      await fetchProxies();
      return true;
    } catch (e) {
      error.value = e.message;
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteProxy = async (id) => {
    isLoading.value = true;
    try {
      const res = await fetch(`/api/proxies?id=${id}`, {
        method: 'DELETE'
      });
      if (!res.ok) throw new Error('Failed to delete proxy');
      await fetchProxies();
      return true;
    } catch (e) {
      error.value = e.message;
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const fetchWebPort = async () => {
    try {
      const res = await fetch('/api/config/webport');
      if (res.ok) {
        const data = await res.json();
        webPort.value = data.web_port;
      }
    } catch (e) {
      console.error('Failed to fetch web port', e);
    }
  };

  const saveWebPort = async (port) => {
    isLoading.value = true;
    try {
      const res = await fetch('/api/config/webport', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ web_port: port })
      });
      if (!res.ok) throw new Error('Failed to save web port');
      webPort.value = port;
      return true;
    } catch (e) {
      error.value = e.message;
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const fetchStatus = async () => {
    try {
      const res = await fetch('/api/status');
      if (res.ok) {
        const data = await res.json();
        status.value = data;
        // Optionally update proxies from status if detailed info is there,
        // but explicit fetchProxies is safer for lists.
      } else {
        status.value = { setup_required: false, proxies: [] };
      }
    } catch (e) {
      console.error('Status fetch failed', e);
    }
  };

  const restartSystem = async () => {
      try {
          await fetch('/api/system/restart', { method: 'POST' });
          return true;
      } catch (e) {
          error.value = e.message;
          return false;
      }
  };

  return {
    proxies,
    webPort,
    status,
    isLoading,
    error,
    fetchProxies,
    addProxy,
    updateProxy,
    deleteProxy,
    fetchWebPort,
    saveWebPort,
    fetchStatus,
    restartSystem
  };
});
