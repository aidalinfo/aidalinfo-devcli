<template>
  <div class="mongo-server-selector">
    <div v-if="servers.length === 0" class="no-servers">
      <p>No MongoDB servers configured</p>
      <router-link to="/settings" class="btn btn-primary">
        Configure Servers
      </router-link>
    </div>

    <div v-else>
      <div class="selector-header">
        <label for="server-select">Select MongoDB Server:</label>
        <router-link to="/settings" class="config-link" title="Configure servers">
          <i class="fas fa-cog"></i>
        </router-link>
      </div>

      <select
        id="server-select"
        v-model="selectedServerId"
        @change="onServerChange"
        class="server-select"
      >
        <option value="" disabled>Choose a server...</option>
        <option
          v-for="server in servers"
          :key="server.id"
          :value="server.id"
        >
          {{ server.name }}
          {{ server.isDefault ? '(Default)' : '' }}
          - {{ server.host }}:{{ server.port }}
        </option>
      </select>

      <div v-if="selectedServer" class="server-details">
        <div class="detail-row">
          <span class="label">Host:</span>
          <span class="value">{{ selectedServer.host }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Port:</span>
          <span class="value">{{ selectedServer.port }}</span>
        </div>
        <div class="detail-row" v-if="selectedServer.user">
          <span class="label">User:</span>
          <span class="value">{{ selectedServer.user }}</span>
        </div>
        <div v-if="showConnectionStatus" class="connection-status">
          <span v-if="connectionTested === null" class="testing">
            <i class="fas fa-spinner fa-spin"></i> Testing connection...
          </span>
          <span v-else-if="connectionTested" class="success">
            <i class="fas fa-check-circle"></i> Connection successful
          </span>
          <span v-else class="error">
            <i class="fas fa-times-circle"></i> Connection failed
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { MongoServersManager, type MongoServer } from '@/utils/mongoServers';

const props = defineProps<{
  modelValue?: string;
  autoSelectDefault?: boolean;
  showDetails?: boolean;
  showConnectionStatus?: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
  'server-selected': [server: MongoServer];
}>();

const servers = ref<MongoServer[]>([]);
const selectedServerId = ref<string>('');
const connectionTested = ref<boolean | null>(null);

const selectedServer = computed(() => {
  if (!selectedServerId.value) return null;
  return servers.value.find(s => s.id === selectedServerId.value);
});

const onServerChange = () => {
  if (selectedServer.value) {
    emit('update:modelValue', selectedServerId.value);
    emit('server-selected', selectedServer.value);
    
    if (props.showConnectionStatus) {
      testConnection();
    }
  }
};

const testConnection = async () => {
  if (!selectedServer.value) return;
  
  connectionTested.value = null;
  const result = await MongoServersManager.testConnection(selectedServer.value);
  connectionTested.value = result.success;
};

const loadServers = () => {
  servers.value = MongoServersManager.getServers();
  
  // Auto-select default server if prop is set and no value is provided
  if (props.autoSelectDefault && !props.modelValue && servers.value.length > 0) {
    const defaultServer = MongoServersManager.getDefaultServer();
    if (defaultServer) {
      selectedServerId.value = defaultServer.id;
      onServerChange();
    }
  } else if (props.modelValue) {
    selectedServerId.value = props.modelValue;
  }
};

// Watch for external value changes
watch(() => props.modelValue, (newValue) => {
  if (newValue && newValue !== selectedServerId.value) {
    selectedServerId.value = newValue;
  }
});

onMounted(() => {
  loadServers();
});

// Expose method to get selected server for parent components
defineExpose({
  getSelectedServer: () => selectedServer.value
});
</script>

<style scoped>
.mongo-server-selector {
  margin: 20px 0;
}

.no-servers {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.no-servers p {
  color: #6c757d;
  margin-bottom: 16px;
}

.selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.selector-header label {
  font-weight: 500;
  color: #495057;
}

.config-link {
  color: #6c757d;
  transition: color 0.2s;
}

.config-link:hover {
  color: #007bff;
}

.server-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

.server-select:focus {
  outline: none;
  border-color: #80bdff;
  box-shadow: 0 0 0 0.2rem rgba(0, 123, 255, 0.25);
}

.server-details {
  margin-top: 16px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 13px;
}

.detail-row {
  display: flex;
  margin-bottom: 8px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row .label {
  font-weight: 500;
  margin-right: 8px;
  min-width: 60px;
  color: #6c757d;
}

.detail-row .value {
  color: #495057;
}

.connection-status {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #dee2e6;
}

.connection-status span {
  display: flex;
  align-items: center;
  gap: 8px;
}

.connection-status .testing {
  color: #007bff;
}

.connection-status .success {
  color: #28a745;
}

.connection-status .error {
  color: #dc3545;
}

.btn {
  display: inline-block;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-primary:hover {
  background: #0056b3;
}
</style>