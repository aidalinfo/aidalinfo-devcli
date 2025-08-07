<template>
  <div class="space-y-4">
    <!-- État vide -->
    <div v-if="filteredServers.length === 0" class="text-center py-8 bg-muted/50 rounded-lg">
      <div class="mx-auto w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center mb-3">
        <svg class="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
        </svg>
      </div>
      <p class="text-sm text-muted-foreground mb-3">
        {{ props.excludeServerId ? 'Aucun autre serveur PostgreSQL disponible' : 'Aucun serveur PostgreSQL configuré' }}
      </p>
      <Button size="sm" as="router-link" to="/settings">
        Configurer les serveurs
      </Button>
    </div>

    <!-- Sélecteur de serveur -->
    <div v-else class="space-y-4">
      <div class="flex justify-between items-center">
        <Label for="server-select" class="text-sm font-medium">Serveur PostgreSQL :</Label>
        <Button variant="ghost" size="sm" as="router-link" to="/settings" title="Configurer les serveurs">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
          </svg>
        </Button>
      </div>

      <Select v-model="selectedServerId" @update:model-value="onServerChange">
        <SelectTrigger class="w-full">
          <SelectValue placeholder="Choisir un serveur..." />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="server in filteredServers"
            :key="server.id"
            :value="server.id"
          >
            {{ server.name }}
            {{ server.isDefault ? '(Par défaut)' : '' }}
            - {{ server.host }}:{{ server.port }}
          </SelectItem>
        </SelectContent>
      </Select>

      <!-- Card du serveur sélectionné -->
      <div v-if="selectedServer">
        <Label class="text-sm font-medium mb-2 block">Serveur sélectionné</Label>
        <Card class="p-4">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
              </svg>
            </div>
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <h4 class="font-medium">{{ selectedServer.name }}</h4>
                <span v-if="selectedServer.isDefault" class="px-2 py-0.5 bg-secondary text-secondary-foreground text-xs rounded-md">
                  Par défaut
                </span>
              </div>
              <p class="text-sm text-muted-foreground">{{ selectedServer.host }}:{{ selectedServer.port }}</p>
              <p v-if="selectedServer.user" class="text-xs text-muted-foreground">Utilisateur: {{ selectedServer.user }}</p>
              <p v-if="selectedServer.database" class="text-xs text-muted-foreground">Base: {{ selectedServer.database }}</p>
            </div>
          </div>
          
          <!-- Statut de connexion -->
          <div v-if="showConnectionStatus" class="mt-4 pt-4 border-t">
            <div v-if="connectionTested === null" class="flex items-center gap-2 text-blue-600">
              <svg class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
              </svg>
              <span class="text-sm">Test de connexion en cours...</span>
            </div>
            <div v-else-if="connectionTested" class="flex items-center gap-2 text-green-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span class="text-sm">Connexion réussie</span>
            </div>
            <div v-else class="flex items-center gap-2 text-red-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span class="text-sm">Échec de la connexion</span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { PostgresServersManager, type PostgresServer } from '@/utils/postgresServers';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const props = defineProps<{
  modelValue?: string;
  autoSelectDefault?: boolean;
  showDetails?: boolean;
  showConnectionStatus?: boolean;
  excludeServerId?: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
  'server-selected': [server: PostgresServer];
}>();

const servers = ref<PostgresServer[]>([]);
const selectedServerId = ref<string>('');
const connectionTested = ref<boolean | null>(null);

// Filtrer les serveurs pour exclure le serveur spécifié
const filteredServers = computed(() => {
  if (!props.excludeServerId) {
    return servers.value;
  }
  return servers.value.filter(server => server.id !== props.excludeServerId);
});

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
  const result = await PostgresServersManager.testConnection(selectedServer.value);
  connectionTested.value = result.success;
};

const loadServers = () => {
  servers.value = PostgresServersManager.getServers();
  
  // Auto-select default server if prop is set and no value is provided
  if (props.autoSelectDefault && !props.modelValue && filteredServers.value.length > 0) {
    const defaultServer = PostgresServersManager.getDefaultServer();
    // S'assurer que le serveur par défaut n'est pas exclu
    if (defaultServer && defaultServer.id !== props.excludeServerId) {
      selectedServerId.value = defaultServer.id;
      onServerChange();
    } else {
      // Si le serveur par défaut est exclu, sélectionner le premier serveur disponible
      const firstAvailable = filteredServers.value[0];
      if (firstAvailable) {
        selectedServerId.value = firstAvailable.id;
        onServerChange();
      }
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