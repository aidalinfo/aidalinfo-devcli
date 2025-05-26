<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ListBackupsWithCreds } from '../../wailsjs/go/main/App'
import { backend } from '../../wailsjs/go/models'
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select'

const PROJECTS = [
  { label: 'Sat&Lease V2', value: 'Sat&LeaseV2', mongo: 'backup/prod-sateleasev2/mongo/', bucket: 'backup/prod-sateleasev2/bucket/' },
  // Ajoute ici d'autres projets si besoin
]

const project = ref(PROJECTS[0].value)
const loading = ref(false)
const error = ref('')
const mongoBackups = ref<string[]>([])
const bucketBackups = ref<string[]>([])

function getCurrentProject() {
  return PROJECTS.find(p => p.value === project.value)!
}

function getS3Credentials(): backend.S3Credentials {
  return new backend.S3Credentials({
    accessKey: localStorage.getItem('s3_access_key') || '',
    secretKey: localStorage.getItem('s3_secret_key') || ''
  })
}

async function fetchBackups() {
  loading.value = true
  error.value = ''
  mongoBackups.value = []
  bucketBackups.value = []
  const creds = getS3Credentials()
  const current = getCurrentProject()
  try {
    const [mongo, bucket] = await Promise.all([
      ListBackupsWithCreds(creds, current.mongo),
      ListBackupsWithCreds(creds, current.bucket)
    ])
    mongoBackups.value = mongo
    bucketBackups.value = bucket
  } catch (e: any) {
    error.value = e.message || e.toString()
  } finally {
    loading.value = false
  }
}

onMounted(fetchBackups)
</script>

<template>
  <Card class="max-w-2xl mx-auto mt-10">
    <CardHeader>
      <CardTitle>Backups S3</CardTitle>
      <CardDescription>
        Sélectionnez un projet et visualisez les backups disponibles (Mongo & Bucket).
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div class="mb-4">
        <Label for="project">Projet</Label>
        <Select v-model="project">
          <SelectTrigger id="project" class="w-64">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="p in PROJECTS" :key="p.value" :value="p.value">{{ p.label }}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button @click="fetchBackups" :disabled="loading" class="mb-6">
        {{ loading ? 'Chargement...' : 'Rafraîchir la liste des backups' }}
      </Button>
      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700">{{ error }}</div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div>
          <h3 class="font-semibold mb-2">Backups Mongo</h3>
          <Separator class="mb-2" />
          <ul class="text-sm space-y-1">
            <li v-for="file in mongoBackups" :key="file">{{ file }}</li>
            <li v-if="!mongoBackups.length && !loading">Aucun backup trouvé.</li>
          </ul>
        </div>
        <div>
          <h3 class="font-semibold mb-2">Backups Bucket</h3>
          <Separator class="mb-2" />
          <ul class="text-sm space-y-1">
            <li v-for="file in bucketBackups" :key="file">{{ file }}</li>
            <li v-if="!bucketBackups.length && !loading">Aucun backup trouvé.</li>
          </ul>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
