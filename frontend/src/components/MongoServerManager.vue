<template>
  <div class="mongo-server-manager">
    <!-- Liste des serveurs -->
    <div class="servers-list">
      <div class="header">
        <h3>MongoDB Servers</h3>
        <button @click="showAddModal = true" class="btn btn-primary">
          <i class="fas fa-plus"></i> Add Server
        </button>
      </div>

      <div v-if="servers.length === 0" class="empty-state">
        <i class="fas fa-database"></i>
        <p>No MongoDB servers configured</p>
        <button @click="showAddModal = true" class="btn btn-secondary">
          Add your first server
        </button>
      </div>

      <div v-else class="servers-table">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Host</th>
              <th>Port</th>
              <th>User</th>
              <th>Default</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="server in servers" :key="server.id">
              <td>
                <span class="server-name">{{ server.name }}</span>
                <span v-if="server.isDefault" class="badge badge-primary">Default</span>
              </td>
              <td>{{ server.host }}</td>
              <td>{{ server.port }}</td>
              <td>{{ server.user || '-' }}</td>
              <td>
                <button
                  v-if="!server.isDefault"
                  @click="setAsDefault(server.id)"
                  class="btn btn-sm btn-link"
                  title="Set as default"
                >
                  <i class="far fa-star"></i>
                </button>
                <span v-else class="default-star">
                  <i class="fas fa-star"></i>
                </span>
              </td>
              <td class="actions">
                <button
                  @click="testConnection(server)"
                  class="btn btn-sm btn-secondary"
                  title="Test connection"
                >
                  <i class="fas fa-plug"></i>
                </button>
                <button
                  @click="editServer(server)"
                  class="btn btn-sm btn-primary"
                  title="Edit"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  @click="confirmDelete(server)"
                  class="btn btn-sm btn-danger"
                  title="Delete"
                  :disabled="servers.length === 1"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Actions d'export/import -->
      <div class="export-import-actions">
        <button @click="exportServers" class="btn btn-secondary">
          <i class="fas fa-download"></i> Export Configuration
        </button>
        <button @click="showImportModal = true" class="btn btn-secondary">
          <i class="fas fa-upload"></i> Import Configuration
        </button>
      </div>
    </div>

    <!-- Modal Ajout/Edition -->
    <div v-if="showAddModal || editingServer" class="modal" @click.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingServer ? 'Edit Server' : 'Add MongoDB Server' }}</h3>
          <button @click="closeModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveServer">
            <div class="form-group">
              <label for="name">Server Name *</label>
              <input
                v-model="formData.name"
                type="text"
                id="name"
                required
                placeholder="e.g., Production MongoDB"
              />
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="host">Host *</label>
                <input
                  v-model="formData.host"
                  type="text"
                  id="host"
                  required
                  placeholder="localhost"
                />
              </div>

              <div class="form-group">
                <label for="port">Port *</label>
                <input
                  v-model="formData.port"
                  type="text"
                  id="port"
                  required
                  placeholder="27017"
                />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="user">Username</label>
                <input
                  v-model="formData.user"
                  type="text"
                  id="user"
                  placeholder="Optional"
                />
              </div>

              <div class="form-group">
                <label for="password">Password</label>
                <input
                  v-model="formData.password"
                  type="password"
                  id="password"
                  placeholder="Optional"
                />
              </div>
            </div>

            <div class="form-group">
              <label class="checkbox-label">
                <input
                  v-model="formData.isDefault"
                  type="checkbox"
                />
                Set as default server
              </label>
            </div>

            <div class="modal-footer">
              <button type="button" @click="closeModal" class="btn btn-secondary">
                Cancel
              </button>
              <button type="submit" class="btn btn-primary">
                {{ editingServer ? 'Update' : 'Add' }} Server
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Modal Import -->
    <div v-if="showImportModal" class="modal" @click.self="showImportModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Import Server Configuration</h3>
          <button @click="showImportModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="importData">Paste JSON configuration:</label>
            <textarea
              v-model="importData"
              id="importData"
              rows="10"
              placeholder="Paste your MongoDB servers configuration JSON here..."
            ></textarea>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="replaceOnImport" type="checkbox" />
              Replace existing servers (unchecked = merge)
            </label>
          </div>

          <div class="modal-footer">
            <button @click="showImportModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button @click="importServers" class="btn btn-primary">
              Import
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal de confirmation de suppression -->
    <div v-if="serverToDelete" class="modal" @click.self="serverToDelete = null">
      <div class="modal-content modal-small">
        <div class="modal-header">
          <h3>Confirm Deletion</h3>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete the server "{{ serverToDelete.name }}"?</p>
        </div>
        <div class="modal-footer">
          <button @click="serverToDelete = null" class="btn btn-secondary">
            Cancel
          </button>
          <button @click="deleteServer" class="btn btn-danger">
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { MongoServersManager, type MongoServer } from '@/utils/mongoServers';
import { showNotification } from '@/utils/notifications';

const servers = ref<MongoServer[]>([]);
const showAddModal = ref(false);
const showImportModal = ref(false);
const editingServer = ref<MongoServer | null>(null);
const serverToDelete = ref<MongoServer | null>(null);
const importData = ref('');
const replaceOnImport = ref(false);

const formData = ref({
  name: '',
  host: 'localhost',
  port: '27017',
  user: '',
  password: '',
  isDefault: false
});

const loadServers = () => {
  servers.value = MongoServersManager.getServers();
};

const closeModal = () => {
  showAddModal.value = false;
  editingServer.value = null;
  resetForm();
};

const resetForm = () => {
  formData.value = {
    name: '',
    host: 'localhost',
    port: '27017',
    user: '',
    password: '',
    isDefault: false
  };
};

const editServer = (server: MongoServer) => {
  editingServer.value = server;
  formData.value = {
    name: server.name,
    host: server.host,
    port: server.port,
    user: server.user,
    password: server.password,
    isDefault: server.isDefault || false
  };
};

const saveServer = () => {
  if (editingServer.value) {
    // Update existing server
    MongoServersManager.updateServer(editingServer.value.id, formData.value);
    showNotification('Server updated successfully', 'success');
  } else {
    // Add new server
    MongoServersManager.addServer(formData.value);
    showNotification('Server added successfully', 'success');
  }
  
  loadServers();
  closeModal();
};

const setAsDefault = (id: string) => {
  MongoServersManager.setDefaultServer(id);
  loadServers();
  showNotification('Default server updated', 'success');
};

const confirmDelete = (server: MongoServer) => {
  serverToDelete.value = server;
};

const deleteServer = () => {
  if (serverToDelete.value) {
    MongoServersManager.deleteServer(serverToDelete.value.id);
    loadServers();
    showNotification('Server deleted successfully', 'success');
    serverToDelete.value = null;
  }
};

const testConnection = async (server: MongoServer) => {
  const result = await MongoServersManager.testConnection(server);
  showNotification(
    result.success ? 'Connection successful' : `Connection failed: ${result.message}`,
    result.success ? 'success' : 'error'
  );
};

const exportServers = () => {
  const json = MongoServersManager.exportServers();
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `mongodb-servers-${new Date().toISOString().split('T')[0]}.json`;
  a.click();
  URL.revokeObjectURL(url);
  showNotification('Configuration exported successfully', 'success');
};

const importServers = () => {
  const result = MongoServersManager.importServers(importData.value, replaceOnImport.value);
  if (result.success) {
    showNotification(result.message, 'success');
    loadServers();
    showImportModal.value = false;
    importData.value = '';
  } else {
    showNotification(result.message, 'error');
  }
};

onMounted(() => {
  loadServers();
});
</script>

<style scoped>
.mongo-server-manager {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  margin: 0;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.empty-state i {
  font-size: 48px;
  color: #6c757d;
  margin-bottom: 20px;
}

.empty-state p {
  color: #6c757d;
  margin-bottom: 20px;
}

.servers-table {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.servers-table table {
  width: 100%;
  border-collapse: collapse;
}

.servers-table th {
  background: #f8f9fa;
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #495057;
  border-bottom: 2px solid #dee2e6;
}

.servers-table td {
  padding: 12px;
  border-bottom: 1px solid #dee2e6;
}

.servers-table tbody tr:hover {
  background: #f8f9fa;
}

.server-name {
  font-weight: 500;
  margin-right: 8px;
}

.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-primary {
  background: #007bff;
  color: white;
}

.default-star {
  color: #ffc107;
}

.actions {
  display: flex;
  gap: 8px;
}

.actions button {
  padding: 4px 8px;
}

.export-import-actions {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

/* Modal styles */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-small {
  max-width: 400px;
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #dee2e6;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6c757d;
}

.close-btn:hover {
  color: #495057;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  padding: 20px;
  border-top: 1px solid #dee2e6;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* Form styles */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  color: #495057;
}

.form-group input[type="text"],
.form-group input[type="password"],
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #80bdff;
  box-shadow: 0 0 0 0.2rem rgba(0, 123, 255, 0.25);
}

.form-group textarea {
  resize: vertical;
  font-family: monospace;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  cursor: pointer;
}

/* Button styles */
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
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

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #545b62;
}

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-danger:hover {
  background: #c82333;
}

.btn-link {
  background: none;
  color: #007bff;
  padding: 0;
}

.btn-link:hover {
  color: #0056b3;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.btn i {
  margin-right: 4px;
}
</style>