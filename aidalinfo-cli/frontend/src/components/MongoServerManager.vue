<template>
  <div class="space-y-6">
    <!-- Header avec bouton d'ajout -->
    <div class="flex justify-between items-center">
      <h3 class="text-lg font-medium">Serveurs MongoDB</h3>
      <Button @click="showAddModal = true">
        Ajouter un serveur
      </Button>
    </div>

    <!-- État vide -->
    <div v-if="servers.length === 0" class="text-center py-12 bg-muted/50 rounded-lg">
      <div class="mx-auto w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
        </svg>
      </div>
      <h3 class="text-lg font-semibold mb-2">Aucun serveur configuré</h3>
      <p class="text-muted-foreground mb-4">Configurez votre premier serveur MongoDB pour commencer</p>
      <Button @click="showAddModal = true" variant="outline">
        Ajouter votre premier serveur
      </Button>
    </div>

    <!-- Liste des serveurs -->
    <div v-else class="space-y-4">
      <div class="grid gap-4">
        <Card v-for="server in servers" :key="server.id" class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <h4 class="font-medium">{{ server.name }}</h4>
                  <Badge v-if="server.isDefault" variant="secondary" class="text-xs">Par défaut</Badge>
                </div>
                <p class="text-sm text-muted-foreground">{{ server.host }}:{{ server.port }}</p>
                <p v-if="server.user" class="text-xs text-muted-foreground">Utilisateur: {{ server.user }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <Button
                v-if="!server.isDefault"
                @click="setAsDefault(server.id)"
                variant="ghost"
                size="sm"
                title="Définir par défaut"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.196-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"></path>
                </svg>
              </Button>
              <span v-else class="text-yellow-500 p-2">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.196-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"></path>
                </svg>
              </span>
              <Button
                @click="testConnection(server)"
                variant="outline"
                size="sm"
                title="Tester la connexion"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                </svg>
              </Button>
              <Button
                @click="editServer(server)"
                variant="outline"
                size="sm"
                title="Modifier"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                </svg>
              </Button>
              <Button
                @click="confirmDelete(server)"
                variant="destructive"
                size="sm"
                title="Supprimer"
                :disabled="servers.length === 1"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                </svg>
              </Button>
            </div>
          </div>
        </Card>
      </div>

      
      <!-- Actions d'export/import -->
      <div class="flex gap-2 pt-4 border-t">
        <Button @click="exportServers" variant="outline">
          Exporter la configuration
        </Button>
        <Button @click="showImportModal = true" variant="outline">
          Importer une configuration
        </Button>
      </div>
    </div>

    <!-- Modal Ajout/Edition -->
    <Dialog :open="showAddModal || !!editingServer" @update:open="(open: boolean) => !open && closeModal()">
      <DialogContent class="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{{ editingServer ? 'Modifier le serveur' : 'Ajouter un serveur MongoDB' }}</DialogTitle>
        </DialogHeader>
        <form @submit.prevent="saveServer" class="space-y-6">
          <div>
            <Label for="name">Nom du serveur *</Label>
            <Input
              v-model="formData.name"
              id="name"
              required
              placeholder="ex: MongoDB Production"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="host">Hôte *</Label>
              <Input
                v-model="formData.host"
                id="host"
                required
                placeholder="localhost"
              />
            </div>
            <div>
              <Label for="port">Port *</Label>
              <Input
                v-model="formData.port"
                id="port"
                required
                placeholder="27017"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="user">Nom d'utilisateur</Label>
              <Input
                v-model="formData.user"
                id="user"
                placeholder="Optionnel"
              />
            </div>
            <div>
              <Label for="password">Mot de passe</Label>
              <Input
                v-model="formData.password"
                type="password"
                id="password"
                placeholder="Optionnel"
              />
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <input
              v-model="formData.isDefault"
              type="checkbox"
              id="default-check"
            />
            <Label for="default-check">Définir comme serveur par défaut</Label>
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="closeModal">
            Annuler
          </Button>
          <Button @click="saveServer">
            {{ editingServer ? 'Mettre à jour' : 'Ajouter' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal Import -->
    <Dialog :open="showImportModal" @update:open="(open: boolean) => showImportModal = open">
      <DialogContent class="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Importer une configuration</DialogTitle>
        </DialogHeader>
        <div class="space-y-6">
          <div>
            <Label for="importData">Configuration JSON :</Label>
            <textarea
              v-model="importData"
              id="importData"
              rows="10"
              placeholder="Collez votre configuration JSON des serveurs MongoDB ici..."
            ></textarea>
          </div>

          <div class="flex items-center space-x-2">
            <input v-model="replaceOnImport" type="checkbox" id="replace-check" />
            <Label for="replace-check">Remplacer les serveurs existants (décoché = fusionner)</Label>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showImportModal = false">
            Annuler
          </Button>
          <Button @click="importServers">
            Importer
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal de confirmation de suppression -->
    <Dialog :open="!!serverToDelete" @update:open="(open: boolean) => !open && (serverToDelete = null)">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Confirmer la suppression</DialogTitle>
        </DialogHeader>
        <p>Êtes-vous sûr de vouloir supprimer le serveur "{{ serverToDelete?.name }}" ?</p>
        <DialogFooter>
          <Button variant="outline" @click="serverToDelete = null">
            Annuler
          </Button>
          <Button variant="destructive" @click="deleteServer">
            Supprimer
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { MongoServersManager, type MongoServer } from '@/utils/mongoServers';
import { toast } from 'vue-sonner';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

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
    toast.success('Serveur mis à jour avec succès');
  } else {
    // Add new server
    MongoServersManager.addServer(formData.value);
    toast.success('Serveur ajouté avec succès');
  }
  
  loadServers();
  closeModal();
};

const setAsDefault = (id: string) => {
  MongoServersManager.setDefaultServer(id);
  loadServers();
  toast.success('Serveur par défaut mis à jour');
};

const confirmDelete = (server: MongoServer) => {
  serverToDelete.value = server;
};

const deleteServer = () => {
  if (serverToDelete.value) {
    MongoServersManager.deleteServer(serverToDelete.value.id);
    loadServers();
    toast.success('Serveur supprimé avec succès');
    serverToDelete.value = null;
  }
};

const testConnection = async (server: MongoServer) => {
  const result = await MongoServersManager.testConnection(server);
  if (result.success) {
    toast.success('Connexion réussie');
  } else {
    toast.error(`Échec de la connexion: ${result.message}`);
  }
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
  toast.success('Configuration exportée avec succès');
};

const importServers = () => {
  const result = MongoServersManager.importServers(importData.value, replaceOnImport.value);
  if (result.success) {
    toast.success(result.message);
    loadServers();
    showImportModal.value = false;
    importData.value = '';
  } else {
    toast.error(result.message);
  }
};

onMounted(() => {
  loadServers();
});
</script>