<template>
  <Popover>
    <PopoverTrigger as-child>
      <Button 
        variant="ghost" 
        size="icon"
        class="relative h-8 w-8"
      >
        <Terminal class="h-4 w-4" />
        <!-- Badge indicateur si des logs sont actifs -->
        <Badge 
          v-if="hasActiveLogs"
          class="absolute -top-1 -right-1 h-2 w-2 rounded-full p-0 bg-green-500"
        />
      </Button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-[600px] h-[400px] p-0">
      <div class="flex flex-col h-full">
        <!-- Header -->
        <div class="flex items-center justify-between p-3 border-b">
          <h3 class="text-sm font-semibold flex items-center gap-2">
            <Terminal class="h-4 w-4" />
            Console Backend
          </h3>
          <Button 
            variant="ghost" 
            size="sm" 
            @click="clearLogs"
            class="h-6 px-2 text-xs"
          >
            Effacer
          </Button>
        </div>
        
        <!-- Logs container -->
        <div 
          ref="logsContainer"
          class="flex-1 overflow-y-auto p-3 bg-slate-950 text-slate-100 font-mono text-xs"
        >
          <div v-if="logs.length === 0" class="text-slate-400 italic">
            En attente des logs backend...
          </div>
          <div 
            v-for="log in logs" 
            :key="log.id"
            class="mb-1 whitespace-pre-wrap"
            :class="{
              'text-red-400': log.level === 'error',
              'text-yellow-400': log.level === 'warn',
              'text-blue-400': log.level === 'debug',
              'text-green-400': log.level === 'success',
              'text-slate-200': log.level === 'info'
            }"
          >
            <span class="text-slate-500">[{{ formatTime(log.timestamp) }}]</span>
            {{ log.message }}
          </div>
        </div>
        
        <!-- Footer avec statut -->
        <div class="p-2 border-t bg-slate-50 text-xs text-slate-600">
          {{ logs.length }} lignes • 
          <span v-if="hasActiveLogs" class="text-green-600">● En cours</span>
          <span v-else class="text-slate-400">○ Inactif</span>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Terminal } from 'lucide-vue-next'
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime.js";

interface LogEntry {
  id: string
  message: string
  level: 'info' | 'debug' | 'warn' | 'error' | 'success'
  timestamp: Date
}

const logs = ref<LogEntry[]>([])
const logsContainer = ref<HTMLElement>()

// Computed pour détecter l'activité récente
const hasActiveLogs = computed(() => {
  if (logs.value.length === 0) return false
  const lastLog = logs.value[logs.value.length - 1]
  const timeSinceLastLog = Date.now() - lastLog.timestamp.getTime()
  return timeSinceLastLog < 30000 // Actif si dernier log < 30 secondes
})

const addLog = (message: string, level: LogEntry['level'] = 'info') => {
  const newLog: LogEntry = {
    id: Date.now().toString(),
    message,
    level,
    timestamp: new Date()
  }
  
  logs.value.push(newLog)
  
  // Limiter à 500 lignes pour éviter la surcharge
  if (logs.value.length > 500) {
    logs.value = logs.value.slice(-500)
  }
  
  // Auto-scroll vers le bas
  nextTick(() => {
    if (logsContainer.value) {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight
    }
  })
}

const clearLogs = () => {
  logs.value = []
}

const formatTime = (date: Date) => {
  return date.toLocaleTimeString('fr-FR', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// Fonction pour parser les logs backend
const parseBackendLog = (logLine: string) => {
  // Détecter le niveau de log basé sur le contenu
  if (logLine.includes('[DEBUG]') || logLine.includes('DEBUG')) {
    return { message: logLine, level: 'debug' as const }
  } else if (logLine.includes('ERREUR') || logLine.includes('ERROR')) {
    return { message: logLine, level: 'error' as const }
  } else if (logLine.includes('terminé avec succès') || logLine.includes('SUCCESS')) {
    return { message: logLine, level: 'success' as const }
  } else if (logLine.includes('WARN')) {
    return { message: logLine, level: 'warn' as const }
  } else {
    return { message: logLine, level: 'info' as const }
  }
}

// Simulation de réception des logs backend
// À remplacer par votre vraie intégration (WebSocket, polling, etc.)
const simulateBackendLogs = () => {
  // Exemple de logs
  setTimeout(() => {
    addLog('[DEBUG] RestoreS3Backup: Début de la restauration S3', 'debug')
  }, 1000)
  
  setTimeout(() => {
    addLog('[DEBUG] Génération de l\'URL presignée...', 'debug')
  }, 2000)
  
  setTimeout(() => {
    addLog('[DEBUG] Début du téléchargement HTTP avec retry...', 'debug')
  }, 3000)
  
  setTimeout(() => {
    addLog('[DEBUG] Téléchargement: 25% (1234567/4938269 octets)', 'info')
  }, 4000)
  
  setTimeout(() => {
    addLog('[DEBUG] Téléchargement: 50% (2469134/4938269 octets)', 'info')
  }, 5000)
  
  setTimeout(() => {
    addLog('[DEBUG] Téléchargement: 75% (3703701/4938269 octets)', 'info')
  }, 6000)
  
  setTimeout(() => {
    addLog('[DEBUG] Téléchargement terminé avec succès! Écrit: 4938269 octets', 'success')
  }, 7000)
}

onMounted(() => {
  // Démarrer la simulation (à remplacer par votre vraie intégration)
  // simulateBackendLogs()

  EventsOn('backend-log', (msg: string) => {
    addLog(msg)
  })
})

onUnmounted(() => {
  EventsOff('backend-log')
})

// Exposer les fonctions pour utilisation externe
defineExpose({
  addLog,
  parseBackendLog,
  clearLogs
})
</script>