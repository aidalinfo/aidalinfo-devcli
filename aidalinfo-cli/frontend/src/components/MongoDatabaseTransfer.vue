<template>
  <Dialog :open="open" @update:open="handleClose">
    <DialogContent class="sm:max-w-[700px] max-h-[90vh] flex flex-col">
      <DialogHeader>
        <DialogTitle>Transfert de base de données MongoDB</DialogTitle>
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
              placeholder="ex: MongoDB Dev"
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
                placeholder="27017"
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
          <MongoServerSelector
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
                <Label for="select-all" class="text-sm font-medium">
                  Tout sélectionner ({{ selectedDatabases.length }}/{{ databases.length }})
                </Label>
              </div>
              <div class="space-y-1 max-h-48 overflow-y-auto">
                <div v-for="db in databases" :key="db" class="flex items-center gap-2 py-1">
                  <input
                    type="checkbox"
                    :id="`db-${db}`"
                    :value="db"
                    v-model="selectedDatabases"
                  />
                  <Label :for="`db-${db}`" class="text-sm cursor-pointer flex-1">
                    {{ db }}
                  </Label>
                </div>
              </div>
            </div>
          </Card>
        </div>

        <!-- Options de transfert -->
        <div v-if="selectedDatabases.length > 0" class="space-y-3">
          <Label class="text-sm font-medium mb-2 block">Options de transfert</Label>
          <div class="flex items-center gap-2">
            <input
              type="checkbox"
              id="drop-existing"
              v-model="dropExisting"
            />
            <Label for="drop-existing" class="text-sm">
              Supprimer les bases existantes avant le transfert (--drop)
            </Label>
          </div>
        </div>

        <!-- Barre de progression -->
        <div v-if="transferInProgress" class="space-y-2">
          <div class="flex items-center justify-between text-sm">
            <span>Transfert en cours...</span>
            <span>{{ currentDatabase }} ({{ currentStep }}/{{ totalSteps }})</span>
          </div>
          <div class="w-full bg-muted rounded-full h-2">
            <div 
              class="bg-primary h-2 rounded-full transition-all duration-300"
              :style="`width: ${progress}%`"
            ></div>
          </div>
          <p class="text-xs text-muted-foreground">{{ statusMessage }}</p>
        </div>
      </div>

      <DialogFooter class="flex-shrink-0">
        <Button variant="outline" @click="handleClose" :disabled="transferInProgress">
          Annuler
        </Button>
        <Button 
          @click="startTransfer" 
          :disabled="!canStartTransfer || transferInProgress"
        >
          <span v-if="transferInProgress" class="flex items-center gap-2">
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
            Transfert en cours...
          </span>
          <span v-else>
            Démarrer le transfert
          </span>
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { toast } from 'vue-sonner';
import { MongoServersManager, type MongoServer } from '@/utils/mongoServers';
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogDescription,
  DialogFooter 
} from '@/components/ui/dialog';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import MongoServerSelector from './MongoServerSelector.vue';
import { ListMongoDatabases, TransferMongoDatabase } from '../../wailsjs/go/main/App';

interface Props {
  open: boolean;
  sourceServer: MongoServer;
}

const props = defineProps<Props>();
const emit = defineEmits(['close']);

const destinationType = ref<'clone' | 'existing'>('existing');
const selectedDestinationId = ref('');
const databases = ref<string[]>([]);
const selectedDatabases = ref<string[]>([]);
const loadingDatabases = ref(false);
const dropExisting = ref(false);
const transferInProgress = ref(false);
const currentDatabase = ref('');
const currentStep = ref(0);
const totalSteps = ref(0);
const progress = ref(0);
const statusMessage = ref('');

const clonedServer = ref({
  name: '',
  host: props.sourceServer.host,
  port: props.sourceServer.port,
  user: props.sourceServer.user,
  password: props.sourceServer.password
});

const canStartTransfer = computed(() => {
  if (selectedDatabases.value.length === 0) return false;
  
  if (destinationType.value === 'clone') {
    return clonedServer.value.name && clonedServer.value.host && clonedServer.value.port;
  } else {
    return selectedDestinationId.value !== '';
  }
});

const getDestinationServer = (): MongoServer | null => {
  if (destinationType.value === 'clone') {
    return {
      id: 'temp-clone',
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      user: clonedServer.value.user,
      password: clonedServer.value.password,
      isDefault: false
    };
  } else {
    return MongoServersManager.getServerById(selectedDestinationId.value);
  }
};

const loadDatabases = async () => {
  loadingDatabases.value = true;
  try {
    const result = await ListMongoDatabases(
      props.sourceServer.host,
      props.sourceServer.port,
      props.sourceServer.user || '',
      props.sourceServer.password || ''
    );
    databases.value = result.filter(db => !['admin', 'config', 'local'].includes(db));
    toast.success(`${databases.value.length} base(s) de données trouvée(s)`);
  } catch (error) {
    toast.error(`Erreur lors du chargement des bases: ${error}`);
    databases.value = [];
  } finally {
    loadingDatabases.value = false;
  }
};

const toggleSelectAll = () => {
  if (selectedDatabases.value.length === databases.value.length) {
    selectedDatabases.value = [];
  } else {
    selectedDatabases.value = [...databases.value];
  }
};

const startTransfer = async () => {
  const destServer = getDestinationServer();
  if (!destServer) {
    toast.error('Serveur de destination invalide');
    return;
  }

  // Si c'est un clone, créer d'abord le serveur
  if (destinationType.value === 'clone') {
    const newServer = MongoServersManager.addServer({
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      user: clonedServer.value.user,
      password: clonedServer.value.password,
      isDefault: false
    });
    if (!newServer) {
      toast.error('Erreur lors de la création du serveur clone');
      return;
    }
    destServer.id = newServer.id;
  }

  transferInProgress.value = true;
  totalSteps.value = selectedDatabases.value.length;
  currentStep.value = 0;
  
  const errors: string[] = [];

  for (const database of selectedDatabases.value) {
    currentStep.value++;
    currentDatabase.value = database;
    progress.value = (currentStep.value / totalSteps.value) * 100;
    statusMessage.value = `Transfert de ${database}...`;

    try {
      await TransferMongoDatabase(
        props.sourceServer.host,
        props.sourceServer.port,
        props.sourceServer.user || '',
        props.sourceServer.password || '',
        destServer.host,
        destServer.port,
        destServer.user || '',
        destServer.password || '',
        database,
        dropExisting.value
      );
      toast.success(`Base ${database} transférée avec succès`);
    } catch (error) {
      const errorMsg = `Erreur transfert ${database}: ${error}`;
      errors.push(errorMsg);
      toast.error(errorMsg);
    }
  }

  transferInProgress.value = false;
  
  if (errors.length === 0) {
    toast.success('Tous les transferts ont été complétés avec succès!');
    handleClose();
  } else {
    toast.error(`${errors.length} erreur(s) lors du transfert`);
  }
};

const handleClose = () => {
  if (!transferInProgress.value) {
    emit('close');
    // Reset form
    destinationType.value = 'existing';
    selectedDestinationId.value = '';
    selectedDatabases.value = [];
    databases.value = [];
    clonedServer.value = {
      name: '',
      host: props.sourceServer.host,
      port: props.sourceServer.port,
      user: props.sourceServer.user,
      password: props.sourceServer.password
    };
  }
};

// Charger les bases quand le composant est monté
onMounted(() => {
  if (props.open) {
    loadDatabases();
  }
});

// Recharger les bases si le modal est réouvert
watch(() => props.open, (newVal) => {
  if (newVal && databases.value.length === 0) {
    loadDatabases();
  }
});

// Réinitialiser le serveur cloné avec les valeurs source
watch(() => props.sourceServer, (newServer) => {
  clonedServer.value = {
    name: '',
    host: newServer.host,
    port: newServer.port,
    user: newServer.user,
    password: newServer.password
  };
}, { immediate: true });
</script>