<template>
  <div class="space-y-4">
    <p class="text-sm text-muted-foreground">
      Choisissez le serveur S3 utilisé pour récupérer les backups (liste, téléchargement et restauration).
    </p>

    <S3ServerSelector
      v-model="selectedServerId"
      :auto-select-default="true"
      :show-details="true"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import S3ServerSelector from '@/components/S3ServerSelector.vue';
import {
  S3ServersManager,
  getBackupRepositoryServerId,
  setBackupRepositoryServerId,
  clearBackupRepositoryServerId,
} from '@/utils/s3Servers';

const selectedServerId = ref<string>('');

const loadSelection = () => {
  const storedId = getBackupRepositoryServerId();
  if (storedId && !S3ServersManager.getServer(storedId)) {
    clearBackupRepositoryServerId();
    selectedServerId.value = '';
    return;
  }
  selectedServerId.value = storedId || '';
};

watch(selectedServerId, (id) => {
  if (id) {
    setBackupRepositoryServerId(id);
  } else {
    clearBackupRepositoryServerId();
  }
});

onMounted(() => {
  loadSelection();
});
</script>
