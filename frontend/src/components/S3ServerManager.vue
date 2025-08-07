<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h3 class="text-lg font-medium">Serveurs S3/MinIO</h3>
      <Button @click="showAddModal = true">
        Ajouter un serveur
      </Button>
    </div>

    <div v-if="servers.length === 0" class="text-center py-12 bg-muted/50 rounded-lg">
      <div class="mx-auto w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path>
        </svg>
      </div>
      <h3 class="text-lg font-semibold mb-2">Aucun serveur configuré</h3>
      <p class="text-muted-foreground mb-4">Configurez votre premier serveur S3 ou MinIO</p>
      <Button @click="showAddModal = true" variant="outline">
        Ajouter votre premier serveur
      </Button>
    </div>

    <div v-else class="space-y-4">
      <div class="grid gap-4">
        <Card v-for="server in servers" :key="server.id" class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <h4 class="font-medium">{{ server.name }}</h4>
                  <Badge v-if="server.isDefault" variant="secondary" class="text-xs">Par défaut</Badge>
                  <Badge v-if="server.useHttps" variant="outline" class="text-xs">HTTPS</Badge>
                </div>
                <p class="text-sm text-muted-foreground">{{ server.host }}:{{ server.port }}</p>
                <p class="text-xs text-muted-foreground">Région: {{ server.region }}</p>
                <p v-if="server.bucket" class="text-xs text-muted-foreground">Bucket: {{ server.bucket }}</p>
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
                @click="cloneServer(server)"
                variant="outline"
                size="sm"
                title="Cloner et transférer"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
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
          <DialogTitle>{{ editingServer ? 'Modifier le serveur' : 'Ajouter un serveur S3/MinIO' }}</DialogTitle>
        </DialogHeader>
        <form @submit.prevent="saveServer" class="space-y-6">
          <div>
            <Label for="name">Nom du serveur *</Label>
            <Input
              v-model="formData.name"
              id="name"
              required
              placeholder="ex: MinIO Production"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="host">Hôte *</Label>
              <Input
                v-model="formData.host"
                id="host"
                required
                placeholder="localhost ou s3.fr-par.scw.cloud"
              />
            </div>
            <div>
              <Label for="port">Port *</Label>
              <Input
                v-model="formData.port"
                id="port"
                required
                placeholder="9000 ou 443"
              />
            </div>
          </div>

          <div>
            <Label for="region">Région *</Label>
            <Input
              v-model="formData.region"
              id="region"
              required
              placeholder="fr-par"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="accessKey">Access Key *</Label>
              <Input
                v-model="formData.accessKey"
                id="accessKey"
                required
                placeholder="Access Key"
              />
            </div>
            <div>
              <Label for="secretKey">Secret Key *</Label>
              <Input
                v-model="formData.secretKey"
                type="password"
                id="secretKey"
                required
                placeholder="Secret Key"
              />
            </div>
          </div>

          <div>
            <Label for="bucket">Bucket par défaut</Label>
            <Input
              v-model="formData.bucket"
              id="bucket"
              placeholder="Optionnel"
            />
            <p class="text-xs text-muted-foreground mt-1">
              Bucket par défaut pour ce serveur (optionnel)
            </p>
          </div>

          <div class="flex items-center space-x-2">
            <input
              v-model="formData.useHttps"
              type="checkbox"
              id="https-check"
            />
            <Label for="https-check">Utiliser HTTPS</Label>
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
              class="w-full p-2 border rounded"
              placeholder="Collez votre configuration JSON des serveurs S3 ici..."
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

    <!-- Modal de transfert de buckets -->
    <S3BucketTransfer 
      v-if="sourceServerForTransfer"
      :open="showTransferModal"
      :source-server="sourceServerForTransfer"
      @close="showTransferModal = false; sourceServerForTransfer = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { S3ServersManager, type S3Server } from '@/utils/s3Servers';
import { toast } from 'vue-sonner';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import S3BucketTransfer from './S3BucketTransfer.vue';

const servers = ref<S3Server[]>([]);
const showAddModal = ref(false);
const showImportModal = ref(false);
const showTransferModal = ref(false);
const sourceServerForTransfer = ref<S3Server | null>(null);
const editingServer = ref<S3Server | null>(null);
const serverToDelete = ref<S3Server | null>(null);
const importData = ref('');
const replaceOnImport = ref(false);

const formData = ref({
  name: '',
  host: 'localhost',
  port: '9000',
  accessKey: '',
  secretKey: '',
  region: 'fr-par',
  useHttps: false,
  bucket: '',
  isDefault: false
});

const loadServers = () => {
  servers.value = S3ServersManager.getServers();
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
    port: '9000',
    accessKey: '',
    secretKey: '',
    region: 'fr-par',
    useHttps: false,
    bucket: '',
    isDefault: false
  };
};

const editServer = (server: S3Server) => {
  editingServer.value = server;
  formData.value = {
    name: server.name,
    host: server.host,
    port: server.port,
    accessKey: server.accessKey,
    secretKey: server.secretKey,
    region: server.region,
    useHttps: server.useHttps,
    bucket: server.bucket || '',
    isDefault: server.isDefault || false
  };
};

const saveServer = () => {
  if (editingServer.value) {
    S3ServersManager.updateServer(editingServer.value.id, formData.value);
    toast.success('Serveur mis à jour avec succès');
  } else {
    S3ServersManager.addServer(formData.value);
    toast.success('Serveur ajouté avec succès');
  }
  
  loadServers();
  closeModal();
};

const setAsDefault = (id: string) => {
  S3ServersManager.setDefaultServer(id);
  loadServers();
  toast.success('Serveur par défaut mis à jour');
};

const confirmDelete = (server: S3Server) => {
  serverToDelete.value = server;
};

const deleteServer = () => {
  if (serverToDelete.value) {
    S3ServersManager.deleteServer(serverToDelete.value.id);
    loadServers();
    toast.success('Serveur supprimé avec succès');
    serverToDelete.value = null;
  }
};

const testConnection = async (server: S3Server) => {
  const result = await S3ServersManager.testConnection(server);
  if (result.success) {
    toast.success('Connexion réussie');
  } else {
    toast.error(`Échec de la connexion: ${result.message}`);
  }
};

const cloneServer = (server: S3Server) => {
  sourceServerForTransfer.value = server;
  showTransferModal.value = true;
};

const exportServers = () => {
  const json = S3ServersManager.exportServers();
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `s3-servers-${new Date().toISOString().split('T')[0]}.json`;
  a.click();
  URL.revokeObjectURL(url);
  toast.success('Configuration exportée avec succès');
};

const importServers = () => {
  const result = S3ServersManager.importServers(importData.value, replaceOnImport.value);
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