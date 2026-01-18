import axios from 'axios';

axios.defaults.withCredentials = true;

// Helper to get cookie by name
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

axios.interceptors.request.use(config => {
    const csrfToken = getCookie('csrf_token');
    if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
    }
    return config;
}, error => {
    return Promise.reject(error);
});

export default axios;
