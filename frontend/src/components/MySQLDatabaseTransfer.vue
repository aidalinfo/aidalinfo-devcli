<template>
  <Dialog :open="open" @update:open="handleClose">
    <DialogContent class="sm:max-w-[700px] max-h-[90vh] flex flex-col">
      <DialogHeader>
        <DialogTitle>Transfert de base de données MySQL</DialogTitle>
        <DialogDescription>
          Transférer des bases de données du serveur source vers un serveur de destination
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-6 overflow-y-auto flex-1 pr-2">
        <!-- Serveur source -->
        <div>
          <Label class="text-sm font-medium mb-2 block">Serveur source</Label>
          <Card class="p-3">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                <svg class="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
                </svg>
              </div>
              <div class="flex-1">
                <p class="font-medium">{{ sourceServer.name }}</p>
                <p class="text-sm text-muted-foreground">{{ sourceServer.host }}:{{ sourceServer.port }}</p>
              </div>
            </div>
          </Card>
        </div>

        <!-- Sélection du serveur de destination -->
        <div>
          <Label class="text-sm font-medium mb-2 block">Serveur de destination</Label>
          <div class="space-y-3">
            <!-- Option 1: Cloner la configuration -->
            <Card 
              class="p-3 cursor-pointer border-2 transition-colors"
              :class="destinationType === 'clone' ? 'border-primary bg-primary/5' : 'border-border hover:border-muted-foreground'"
              @click="destinationType = 'clone'"
            >
              <div class="flex items-center gap-3">
                <div class="w-5 h-5 rounded-full border-2 flex items-center justify-center"
                     :class="destinationType === 'clone' ? 'border-primary' : 'border-muted-foreground'">
                  <div v-if="destinationType === 'clone'" class="w-2 h-2 bg-primary rounded-full"></div>
                </div>
                <div class="flex-1">
                  <p class="font-medium">Créer un nouveau serveur (clone)</p>
                  <p class="text-sm text-muted-foreground">Dupliquer la configuration et modifier les paramètres</p>
                </div>
              </div>
            </Card>

            <!-- Option 2: Sélectionner un serveur existant -->
            <Card 
              class="p-3 cursor-pointer border-2 transition-colors"
              :class="destinationType === 'existing' ? 'border-primary bg-primary/5' : 'border-border hover:border-muted-foreground'"
              @click="destinationType = 'existing'"
            >
              <div class="flex items-center gap-3">
                <div class="w-5 h-5 rounded-full border-2 flex items-center justify-center"
                     :class="destinationType === 'existing' ? 'border-primary' : 'border-muted-foreground'">
                  <div v-if="destinationType === 'existing'" class="w-2 h-2 bg-primary rounded-full"></div>
                </div>
                <div class="flex-1">
                  <p class="font-medium">Utiliser un serveur existant</p>
                  <p class="text-sm text-muted-foreground">Sélectionner parmi les serveurs configurés</p>
                </div>
              </div>
            </Card>
          </div>
        </div>

        <!-- Formulaire pour nouveau serveur (si clone) -->
        <div v-if="destinationType === 'clone'" class="space-y-4 p-4 bg-muted/30 rounded-lg">
          <div>
            <Label for="clone-name">Nom du serveur *</Label>
            <Input
              v-model="clonedServer.name"
              id="clone-name"
              placeholder="ex: MySQL Dev"
              required
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="clone-host">Hôte *</Label>
              <Input
                v-model="clonedServer.host"
                id="clone-host"
                placeholder="localhost"
                required
              />
            </div>
            <div>
              <Label for="clone-port">Port *</Label>
              <Input
                v-model="clonedServer.port"
                id="clone-port"
                placeholder="3306"
                required
              />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="clone-user">Utilisateur</Label>
              <Input
                v-model="clonedServer.user"
                id="clone-user"
                placeholder="Optionnel"
              />
            </div>
            <div>
              <Label for="clone-password">Mot de passe</Label>
              <Input
                v-model="clonedServer.password"
                type="password"
                id="clone-password"
                placeholder="Optionnel"
              />
            </div>
          </div>
        </div>

        <!-- Sélection serveur existant -->
        <div v-if="destinationType === 'existing'" class="p-4 bg-muted/30 rounded-lg">
          <MySQLServerSelector
            v-model="selectedDestinationId"
            :exclude-server-id="sourceServer.id"
          />
        </div>

        <!-- Sélection des bases de données -->
        <div v-if="(destinationType === 'clone' && clonedServer.name) || (destinationType === 'existing' && selectedDestinationId)">
          <Label class="text-sm font-medium mb-2 block">Bases de données à transférer</Label>
          <Card class="p-4">
            <div v-if="loadingDatabases" class="flex items-center justify-center py-8">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
            <div v-else-if="databases.length === 0" class="text-center py-8 text-muted-foreground">
              <p>Aucune base de données trouvée</p>
              <Button @click="loadDatabases" variant="outline" size="sm" class="mt-2">
                Réessayer
              </Button>
            </div>
            <div v-else class="space-y-2">
              <div class="flex items-center gap-2 mb-3">
                <input
                  type="checkbox"
                  id="select-all"
                  :checked="selectedDatabases.length === databases.length"
                  @change="toggleSelectAll"
                />
                <Label for="select-all" class="text-sm">Sélectionner tout</Label>
              </div>
              <div class="max-h-48 overflow-y-auto space-y-1 border rounded-md p-2">
                <div v-for="db in databases" :key="db" class="flex items-center gap-2 p-1 hover:bg-muted/50 rounded">
                  <input
                    type="checkbox"
                    :id="`db-${db}`"
                    :value="db"
                    v-model="selectedDatabases"
                  />
                  <Label :for="`db-${db}`" class="text-sm cursor-pointer flex-1">{{ db }}</Label>
                </div>
              </div>
              <p class="text-xs text-muted-foreground mt-2">
                {{ selectedDatabases.length }} base(s) sélectionnée(s) sur {{ databases.length }}
              </p>
            </div>
          </Card>
        </div>

        <!-- Options de transfert -->
        <div v-if="selectedDatabases.length > 0">
          <Label class="text-sm font-medium mb-2 block">Options de transfert</Label>
          <Card class="p-4 space-y-3">
            <div class="flex items-center gap-2">
              <input
                type="checkbox"
                id="drop-existing"
                v-model="dropExisting"
              />
              <Label for="drop-existing" class="text-sm">
                Supprimer les bases existantes avant le transfert
              </Label>
            </div>
            <div class="flex items-center gap-2">
              <input
                type="checkbox"
                id="save-dest"
                v-model="saveDestination"
              />
              <Label for="save-dest" class="text-sm">
                Sauvegarder le serveur de destination dans la configuration
              </Label>
            </div>
          </Card>
        </div>
      </div>

      <!-- Statut du transfert -->
      <div v-if="transferStatus.active" class="mt-4 p-4 bg-muted/30 rounded-lg">
        <div class="space-y-2">
          <div class="flex items-center gap-2">
            <div v-if="!transferStatus.error" class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
            <span class="text-sm font-medium">{{ transferStatus.message }}</span>
          </div>
          <div v-if="transferStatus.current" class="text-xs text-muted-foreground">
            Base en cours: {{ transferStatus.current }} ({{ transferStatus.progress }}/{{ selectedDatabases.length }})
          </div>
          <div v-if="transferStatus.error" class="text-xs text-red-600">
            {{ transferStatus.error }}
          </div>
        </div>
      </div>

      <DialogFooter class="pt-4">
        <Button variant="outline" @click="handleClose" :disabled="transferStatus.active">
          {{ transferStatus.active ? 'Fermer après transfert' : 'Annuler' }}
        </Button>
        <Button 
          @click="startTransfer" 
          :disabled="!canStartTransfer || transferStatus.active"
        >
          {{ transferStatus.active ? 'Transfert en cours...' : 'Démarrer le transfert' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { MySQLServersManager, type MySQLServer } from '@/utils/mysqlServers';
import { ListMySQLDatabases, TransferMySQLDatabase } from '../../wailsjs/go/main/App';
import { toast } from 'vue-sonner';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import MySQLServerSelector from './MySQLServerSelector.vue';

const props = defineProps<{
  open: boolean;
  sourceServer: MySQLServer;
}>();

const emit = defineEmits<{
  close: [];
}>();

const destinationType = ref<'clone' | 'existing' | null>(null);
const clonedServer = ref({
  name: '',
  host: 'localhost',
  port: '3306',
  user: '',
  password: ''
});
const selectedDestinationId = ref('');
const databases = ref<string[]>([]);
const selectedDatabases = ref<string[]>([]);
const loadingDatabases = ref(false);
const dropExisting = ref(false);
const saveDestination = ref(true);
const transferStatus = ref({
  active: false,
  message: '',
  current: '',
  progress: 0,
  error: ''
});

const canStartTransfer = computed(() => {
  if (selectedDatabases.value.length === 0) return false;
  
  if (destinationType.value === 'clone') {
    return clonedServer.value.name && clonedServer.value.host && clonedServer.value.port;
  } else if (destinationType.value === 'existing') {
    return !!selectedDestinationId.value;
  }
  
  return false;
});

const loadDatabases = async () => {
  loadingDatabases.value = true;
  try {
    const dbs = await ListMySQLDatabases(
      props.sourceServer.host,
      props.sourceServer.port,
      props.sourceServer.user,
      props.sourceServer.password
    );
    databases.value = dbs;
  } catch (error) {
    toast.error('Erreur lors du chargement des bases de données');
    console.error(error);
  } finally {
    loadingDatabases.value = false;
  }
};

const toggleSelectAll = (event: Event) => {
  const checked = (event.target as HTMLInputElement).checked;
  selectedDatabases.value = checked ? [...databases.value] : [];
};

const getDestinationServer = (): MySQLServer | null => {
  if (destinationType.value === 'clone') {
    return {
      id: '',
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      user: clonedServer.value.user,
      password: clonedServer.value.password,
      createdAt: '',
      updatedAt: ''
    };
  } else if (destinationType.value === 'existing') {
    return MySQLServersManager.getServer(selectedDestinationId.value) || null;
  }
  return null;
};

const startTransfer = async () => {
  const destServer = getDestinationServer();
  if (!destServer) {
    toast.error('Serveur de destination non défini');
    return;
  }

  transferStatus.value = {
    active: true,
    message: 'Démarrage du transfert...',
    current: '',
    progress: 0,
    error: ''
  };

  // Sauvegarder le serveur de destination si nécessaire
  if (destinationType.value === 'clone' && saveDestination.value) {
    MySQLServersManager.addServer({
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      user: clonedServer.value.user,
      password: clonedServer.value.password
    });
  }

  // Transférer chaque base de données
  for (let i = 0; i < selectedDatabases.value.length; i++) {
    const db = selectedDatabases.value[i];
    transferStatus.value.current = db;
    transferStatus.value.progress = i + 1;
    transferStatus.value.message = `Transfert de ${db}...`;

    try {
      await TransferMySQLDatabase(
        props.sourceServer.host,
        props.sourceServer.port,
        props.sourceServer.user,
        props.sourceServer.password,
        destServer.host,
        destServer.port,
        destServer.user,
        destServer.password,
        db,
        dropExisting.value
      );
    } catch (error) {
      transferStatus.value.error = `Erreur lors du transfert de ${db}: ${error}`;
      toast.error(`Échec du transfert de ${db}`);
      // On continue avec les autres bases
    }
  }

  if (!transferStatus.value.error) {
    transferStatus.value.message = 'Transfert terminé avec succès!';
    toast.success('Toutes les bases de données ont été transférées');
    setTimeout(() => {
      handleClose();
    }, 2000);
  } else {
    transferStatus.value.message = 'Transfert terminé avec des erreurs';
  }
};

const handleClose = () => {
  if (!transferStatus.value.active) {
    emit('close');
  }
};

const resetForm = () => {
  destinationType.value = null;
  clonedServer.value = {
    name: '',
    host: 'localhost',
    port: '3306',
    user: '',
    password: ''
  };
  selectedDestinationId.value = '';
  selectedDatabases.value = [];
  dropExisting.value = false;
  saveDestination.value = true;
  transferStatus.value = {
    active: false,
    message: '',
    current: '',
    progress: 0,
    error: ''
  };
};

// Charger les bases quand un type de destination est sélectionné
watch([destinationType, () => clonedServer.value.name, selectedDestinationId], () => {
  if ((destinationType.value === 'clone' && clonedServer.value.name) || 
      (destinationType.value === 'existing' && selectedDestinationId.value)) {
    loadDatabases();
  }
});

// Reset form when dialog closes
watch(() => props.open, (newVal) => {
  if (!newVal) {
    resetForm();
  } else {
    // Pré-remplir avec un nom de clone
    clonedServer.value.name = `${props.sourceServer.name} - Clone`;
  }
});

onMounted(() => {
  if (props.open) {
    clonedServer.value.name = `${props.sourceServer.name} - Clone`;
  }
});
</script>