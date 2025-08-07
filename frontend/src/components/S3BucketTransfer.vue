<template>
  <Dialog :open="open" @update:open="handleClose">
    <DialogContent class="sm:max-w-[700px] max-h-[90vh] flex flex-col">
      <DialogHeader>
        <DialogTitle>Transfert de buckets S3/MinIO</DialogTitle>
        <DialogDescription>
          Transférer des buckets du serveur source vers un serveur de destination
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
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path>
                </svg>
              </div>
              <div class="flex-1">
                <p class="font-medium">{{ sourceServer.name }}</p>
                <p class="text-sm text-muted-foreground">{{ sourceServer.host }}:{{ sourceServer.port }}</p>
                <p v-if="sourceServer.region" class="text-xs text-muted-foreground">Région: {{ sourceServer.region }}</p>
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
              placeholder="ex: MinIO Dev"
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
                placeholder="9000"
                required
              />
            </div>
          </div>
          <div>
            <Label for="clone-region">Région *</Label>
            <Input
              v-model="clonedServer.region"
              id="clone-region"
              placeholder="fr-par"
              required
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label for="clone-access">Access Key *</Label>
              <Input
                v-model="clonedServer.accessKey"
                id="clone-access"
                placeholder="Access Key"
                required
              />
            </div>
            <div>
              <Label for="clone-secret">Secret Key *</Label>
              <Input
                v-model="clonedServer.secretKey"
                type="password"
                id="clone-secret"
                placeholder="Secret Key"
                required
              />
            </div>
          </div>
          <div class="flex items-center gap-2">
            <input
              type="checkbox"
              id="clone-https"
              v-model="clonedServer.useHttps"
            />
            <Label for="clone-https">Utiliser HTTPS</Label>
          </div>
        </div>

        <!-- Sélection serveur existant -->
        <div v-if="destinationType === 'existing'" class="p-4 bg-muted/30 rounded-lg">
          <S3ServerSelector
            v-model="selectedDestinationId"
            :exclude-server-id="sourceServer.id"
          />
        </div>

        <!-- Sélection des buckets -->
        <div v-if="(destinationType === 'clone' && clonedServer.name) || (destinationType === 'existing' && selectedDestinationId)">
          <Label class="text-sm font-medium mb-2 block">Buckets à transférer</Label>
          <Card class="p-4">
            <div v-if="loadingBuckets" class="flex items-center justify-center py-8">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
            <div v-else-if="buckets.length === 0" class="text-center py-8 text-muted-foreground">
              <p>Aucun bucket trouvé</p>
              <Button @click="loadBuckets" variant="outline" size="sm" class="mt-2">
                Réessayer
              </Button>
            </div>
            <div v-else class="space-y-2">
              <div class="flex items-center gap-2 mb-3">
                <input
                  type="checkbox"
                  id="select-all"
                  :checked="selectedBuckets.length === buckets.length"
                  @change="toggleSelectAll"
                />
                <Label for="select-all" class="text-sm font-medium">
                  Tout sélectionner ({{ selectedBuckets.length }}/{{ buckets.length }})
                </Label>
              </div>
              <div class="space-y-1 max-h-48 overflow-y-auto">
                <div v-for="bucket in buckets" :key="bucket" class="flex items-center gap-2 py-1">
                  <input
                    type="checkbox"
                    :id="`bucket-${bucket}`"
                    :value="bucket"
                    v-model="selectedBuckets"
                  />
                  <Label :for="`bucket-${bucket}`" class="text-sm cursor-pointer flex-1">
                    {{ bucket }}
                  </Label>
                </div>
              </div>
            </div>
          </Card>
        </div>

        <!-- Options de transfert -->
        <div v-if="selectedBuckets.length > 0" class="space-y-3">
          <Label class="text-sm font-medium mb-2 block">Options de transfert</Label>
          <div class="flex items-center gap-2">
            <input
              type="checkbox"
              id="overwrite-existing"
              v-model="overwriteExisting"
            />
            <Label for="overwrite-existing" class="text-sm">
              Écraser les objets existants dans le bucket de destination
            </Label>
          </div>
        </div>

        <!-- Barre de progression -->
        <div v-if="transferInProgress" class="space-y-2">
          <div class="flex items-center justify-between text-sm">
            <span>Transfert en cours...</span>
            <span>{{ currentBucket }} ({{ currentStep }}/{{ totalSteps }})</span>
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
import { S3ServersManager, type S3Server } from '@/utils/s3Servers';
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
import S3ServerSelector from './S3ServerSelector.vue';
// TODO: Importer les fonctions Wails pour S3 quand elles seront disponibles
// import { ListS3Buckets, TransferS3Bucket } from '../../wailsjs/go/main/App';

interface Props {
  open: boolean;
  sourceServer: S3Server;
}

const props = defineProps<Props>();
const emit = defineEmits(['close']);

const destinationType = ref<'clone' | 'existing'>('existing');
const selectedDestinationId = ref('');
const buckets = ref<string[]>([]);
const selectedBuckets = ref<string[]>([]);
const loadingBuckets = ref(false);
const overwriteExisting = ref(false);
const transferInProgress = ref(false);
const currentBucket = ref('');
const currentStep = ref(0);
const totalSteps = ref(0);
const progress = ref(0);
const statusMessage = ref('');

const clonedServer = ref({
  name: '',
  host: props.sourceServer.host,
  port: props.sourceServer.port,
  accessKey: props.sourceServer.accessKey,
  secretKey: props.sourceServer.secretKey,
  region: props.sourceServer.region,
  useHttps: props.sourceServer.useHttps
});

const canStartTransfer = computed(() => {
  if (selectedBuckets.value.length === 0) return false;
  
  if (destinationType.value === 'clone') {
    return clonedServer.value.name && clonedServer.value.host && 
           clonedServer.value.port && clonedServer.value.accessKey && 
           clonedServer.value.secretKey && clonedServer.value.region;
  } else {
    return selectedDestinationId.value !== '';
  }
});

const getDestinationServer = (): S3Server | null => {
  if (destinationType.value === 'clone') {
    return {
      id: 'temp-clone',
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      accessKey: clonedServer.value.accessKey,
      secretKey: clonedServer.value.secretKey,
      region: clonedServer.value.region,
      useHttps: clonedServer.value.useHttps,
      isDefault: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  } else {
    return S3ServersManager.getServerById(selectedDestinationId.value) || null;
  }
};

const loadBuckets = async () => {
  loadingBuckets.value = true;
  try {
    // TODO: Implémenter l'appel à ListS3Buckets quand disponible
    // const result = await ListS3Buckets(
    //   props.sourceServer.host,
    //   props.sourceServer.port,
    //   props.sourceServer.accessKey,
    //   props.sourceServer.secretKey,
    //   props.sourceServer.region,
    //   props.sourceServer.useHttps
    // );
    // buckets.value = result;
    
    // Pour l'instant, utiliser des buckets fictifs pour le développement
    buckets.value = ['backup', 'data', 'uploads', 'archives'];
    toast.success(`${buckets.value.length} bucket(s) trouvé(s)`);
  } catch (error) {
    toast.error(`Erreur lors du chargement des buckets: ${error}`);
    buckets.value = [];
  } finally {
    loadingBuckets.value = false;
  }
};

const toggleSelectAll = () => {
  if (selectedBuckets.value.length === buckets.value.length) {
    selectedBuckets.value = [];
  } else {
    selectedBuckets.value = [...buckets.value];
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
    const newServer = S3ServersManager.addServer({
      name: clonedServer.value.name,
      host: clonedServer.value.host,
      port: clonedServer.value.port,
      accessKey: clonedServer.value.accessKey,
      secretKey: clonedServer.value.secretKey,
      region: clonedServer.value.region,
      useHttps: clonedServer.value.useHttps,
      isDefault: false
    });
    if (!newServer) {
      toast.error('Erreur lors de la création du serveur clone');
      return;
    }
    destServer.id = newServer.id;
  }

  transferInProgress.value = true;
  totalSteps.value = selectedBuckets.value.length;
  currentStep.value = 0;
  
  const errors: string[] = [];

  for (const bucket of selectedBuckets.value) {
    currentStep.value++;
    currentBucket.value = bucket;
    progress.value = (currentStep.value / totalSteps.value) * 100;
    statusMessage.value = `Transfert de ${bucket}...`;

    try {
      // TODO: Implémenter l'appel à TransferS3Bucket quand disponible
      // await TransferS3Bucket(
      //   props.sourceServer.host,
      //   props.sourceServer.port,
      //   props.sourceServer.accessKey,
      //   props.sourceServer.secretKey,
      //   props.sourceServer.region,
      //   props.sourceServer.useHttps,
      //   destServer.host,
      //   destServer.port,
      //   destServer.accessKey,
      //   destServer.secretKey,
      //   destServer.region,
      //   destServer.useHttps,
      //   bucket,
      //   overwriteExisting.value
      // );
      
      // Simulation pour le développement
      await new Promise(resolve => setTimeout(resolve, 1000));
      toast.success(`Bucket ${bucket} transféré avec succès`);
    } catch (error) {
      const errorMsg = `Erreur transfert ${bucket}: ${error}`;
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
    selectedBuckets.value = [];
    buckets.value = [];
    clonedServer.value = {
      name: '',
      host: props.sourceServer.host,
      port: props.sourceServer.port,
      accessKey: props.sourceServer.accessKey,
      secretKey: props.sourceServer.secretKey,
      region: props.sourceServer.region,
      useHttps: props.sourceServer.useHttps
    };
  }
};

// Charger les buckets quand le composant est monté
onMounted(() => {
  if (props.open) {
    loadBuckets();
  }
});

// Recharger les buckets si le modal est réouvert
watch(() => props.open, (newVal) => {
  if (newVal && buckets.value.length === 0) {
    loadBuckets();
  }
});

// Réinitialiser le serveur cloné avec les valeurs source
watch(() => props.sourceServer, (newServer) => {
  clonedServer.value = {
    name: '',
    host: newServer.host,
    port: newServer.port,
    accessKey: newServer.accessKey,
    secretKey: newServer.secretKey,
    region: newServer.region,
    useHttps: newServer.useHttps
  };
}, { immediate: true });
</script>