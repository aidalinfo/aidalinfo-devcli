<template>
  <v-dialog v-model="dialog" max-width="500px" persistent>
    <v-card>
      <v-card-title class="text-h5">
        Privilèges administrateur requis
      </v-card-title>
      
      <v-card-text>
        <v-alert type="info" variant="tonal" class="mb-4">
          La mise à jour nécessite des privilèges administrateur pour remplacer le fichier exécutable.
        </v-alert>
        
        <v-text-field
          v-model="password"
          label="Mot de passe sudo"
          type="password"
          variant="outlined"
          density="compact"
          :error-messages="errorMessage"
          @keyup.enter="confirmPassword"
          autofocus
        />
      </v-card-text>
      
      <v-card-actions>
        <v-spacer />
        <v-btn color="error" variant="text" @click="cancel">
          Annuler
        </v-btn>
        <v-btn 
          color="primary" 
          variant="flat" 
          @click="confirmPassword"
          :loading="loading"
          :disabled="!password"
        >
          Confirmer
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue: boolean
  tmpFilePath: string
}

const props = defineProps<Props>()
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const dialog = ref(false)
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

watch(() => props.modelValue, (val) => {
  dialog.value = val
  if (val) {
    // Réinitialiser le formulaire quand le dialogue s'ouvre
    password.value = ''
    errorMessage.value = ''
    loading.value = false
  }
})

watch(dialog, (val) => {
  emit('update:modelValue', val)
})

const confirmPassword = async () => {
  if (!password.value) {
    errorMessage.value = 'Veuillez entrer votre mot de passe'
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    await emit('confirm', password.value, props.tmpFilePath)
    // Le composant parent fermera le dialogue en cas de succès
  } catch (error: any) {
    // Gérer l'erreur de mot de passe incorrect
    if (error.message?.includes('mot de passe incorrect')) {
      errorMessage.value = 'Mot de passe incorrect'
    } else {
      errorMessage.value = 'Erreur lors de la mise à jour'
    }
  } finally {
    loading.value = false
    // Effacer le mot de passe de la mémoire
    if (errorMessage.value) {
      password.value = ''
    }
  }
}

const cancel = () => {
  emit('cancel')
  dialog.value = false
}
</script>