// Client-side validation utilities
export const validators = {
  // Validate port (e.g., :5020 or 192.168.1.1:502)
  validatePort: (value) => {
    if (!value || value.trim() === '') {
      return 'Port cannot be empty';
    }

    // Check format
    const portRegex = /^(:[1-9][0-9]{0,4}|[a-zA-Z0-9.-]+:[1-9][0-9]{0,4})$/;
    if (!portRegex.test(value)) {
      return 'Invalid port format. Use ":port" or "host:port"';
    }

    // Extract port number
    let portStr;
    if (value.startsWith(':')) {
      portStr = value.substring(1);
    } else {
      portStr = value.split(':')[1];
    }

    const port = parseInt(portStr, 10);
    if (port < 1 || port > 65535) {
      return 'Port must be between 1 and 65535';
    }

    return null;
  },

  // Validate hostname
  validateHostname: (value) => {
    if (!value || value.trim() === '') {
      return 'Hostname cannot be empty';
    }

    if (value.length > 253) {
      return 'Hostname must be 253 characters or less';
    }

    const hostnameRegex = /^[a-zA-Z0-9.-]+$/;
    if (!hostnameRegex.test(value)) {
      return 'Hostname contains invalid characters';
    }

    return null;
  },

  // Validate name
  validateName: (value) => {
    if (!value || value.trim() === '') {
      return 'Name cannot be empty';
    }

    if (value.length > 100) {
      return 'Name must be 100 characters or less';
    }

    return null;
  },

  // Validate timeout (seconds)
  validateTimeout: (value) => {
    if (!value || value.trim() === '') {
      return 'Timeout cannot be empty';
    }

    const timeout = parseInt(value, 10);
    if (isNaN(timeout)) {
      return 'Timeout must be a number';
    }

    if (timeout < 1 || timeout > 300) {
      return 'Timeout must be between 1 and 300 seconds';
    }

    return null;
  },

  // Validate max retries
  validateMaxRetries: (value) => {
    if (!value || value.trim() === '') {
      return 'Max retries cannot be empty';
    }

    const retries = parseInt(value, 10);
    if (isNaN(retries)) {
      return 'Max retries must be a number';
    }

    if (retries < 0 || retries > 10) {
      return 'Max retries must be between 0 and 10';
    }

    return null;
  },

  // Validate max read size
  validateMaxReadSize: (value) => {
    if (!value || value.trim() === '') {
      return 'Max read size cannot be empty';
    }

    const size = parseInt(value, 10);
    if (isNaN(size)) {
      return 'Max read size must be a number';
    }

    if (size < 0 || size > 65535) {
      return 'Max read size must be between 0 and 65535';
    }

    return null;
  },

  // Validate proxy config
  validateProxyConfig: (config) => {
    const errors = {};

    if (!config.name) {
      errors.name = validators.validateName(config.name);
    }

    if (!config.listen_addr) {
      errors.listen_addr = validators.validatePort(config.listen_addr);
    }

    if (!config.target_addr) {
      errors.target_addr = validators.validatePort(config.target_addr);
    }

    if (config.connection_timeout !== undefined) {
      errors.connection_timeout = validators.validateTimeout(config.connection_timeout.toString());
    }

    if (config.read_timeout !== undefined) {
      errors.read_timeout = validators.validateTimeout(config.read_timeout.toString());
    }

    if (config.max_retries !== undefined) {
      errors.max_retries = validators.validateMaxRetries(config.max_retries.toString());
    }

    if (config.max_read_size !== undefined) {
      errors.max_read_size = validators.validateMaxReadSize(config.max_read_size.toString());
    }

    // Check if any errors
    const hasErrors = Object.values(errors).some(error => error !== null);
    return hasErrors ? errors : null;
  }
};

export default validators;
