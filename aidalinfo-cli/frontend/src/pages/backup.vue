<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ListBackupsWithCreds, RestoreMongoBackup, RestoreS3Backup } from '../../wailsjs/go/main/App'
import { backend } from '../../wailsjs/go/models'
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toast } from 'vue-sonner'
import { Plus } from 'lucide-vue-next'

const PROJECTS = [
  { label: 'Sat&Lease V2', value: 'Sat&LeaseV2', mongo: 'backup/prod-sateleasev2/mongo/', bucket: 'backup/prod-sateleasev2/bucket/' },
  // Ajoute ici d'autres projets si besoin
]

const project = ref(PROJECTS[0].value)
const loading = ref(false)
const error = ref('')
const mongoBackups = ref<backend.BackupInfo[]>([])
const bucketBackups = ref<backend.BackupInfo[]>([])

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

async function restoreMongo(file: backend.BackupInfo) {
  const creds = getS3Credentials()
  const current = getCurrentProject()
  // Récupère les paramètres MongoDB depuis localStorage
  const mongoHost = localStorage.getItem('mongo_host') || 'localhost'
  const mongoPort = localStorage.getItem('mongo_port') || '27017'
  const mongoUser = localStorage.getItem('mongo_user') || ''
  const mongoPassword = localStorage.getItem('mongo_password') || ''
  toast.info('Restauration MongoDB en cours...')
  try {
    await RestoreMongoBackup(
      creds,
      current.mongo + file.name,
      mongoHost,
      mongoPort,
      mongoUser,
      mongoPassword
    )
    toast.success('Restauration MongoDB terminée avec succès !')
  } catch (e: any) {
    toast.error('Erreur restauration MongoDB : ' + (e.message || e.toString()))
  }
}

async function restoreS3(file: backend.BackupInfo) {
  const cloudCreds = getS3Credentials()
  const current = getCurrentProject()
  // Récupère les paramètres S3 local depuis localStorage
  const s3Host = localStorage.getItem('s3_host') || 'localhost'
  const s3Port = localStorage.getItem('s3_port') || '9000'
  const s3Region = localStorage.getItem('s3_region') || 'fr-par'
  const s3LocalAccessKey = localStorage.getItem('s3local_access_key') || ''
  const s3LocalSecretKey = localStorage.getItem('s3local_secret_key') || ''
  const localCreds = { accessKey: s3LocalAccessKey, secretKey: s3LocalSecretKey }
  toast.info('Restauration S3 local en cours...')
  try {
    // @ts-ignore
    await RestoreS3Backup(
      cloudCreds,
      localCreds,
      current.bucket + file.name,
      s3Host,
      s3Port,
      s3Region
    )
    toast.success('Restauration S3 local terminée avec succès !')
  } catch (e: any) {
    toast.error('Erreur restauration S3 local : ' + (e.message || e.toString()))
  }
}
</script>

<template>
  
  <div class="mx-5 mt-10 space-y-8">

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

    <Card>
      <CardHeader>
        <CardTitle>Backups Mongo</CardTitle>
        <CardDescription>Liste des backups MongoDB pour le projet sélectionné.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableCaption>Backups MongoDB disponibles</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Nom du backup</TableHead>
              <TableHead>Taille</TableHead>
              <TableHead>Date</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="file in mongoBackups" :key="file.name">
              <TableCell>{{ file.name }}</TableCell>
              <TableCell>{{ (file.size / 1024 / 1024).toFixed(2) }} Mo</TableCell>
              <TableCell>{{ new Date(file.lastModified).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <Button size="icon" variant="ghost" @click="restoreMongo(file)">
                  <Plus class="w-5 h-5 text-primary" />
                </Button>
              </TableCell>
            </TableRow>
            <TableRow v-if="!mongoBackups.length && !loading">
              <TableCell colspan="4">Aucun backup trouvé.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Backups S3 (Bucket)</CardTitle>
        <CardDescription>Liste des backups du stockage S3 pour le projet sélectionné.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableCaption>Backups S3 disponibles</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Nom du backup</TableHead>
              <TableHead>Taille</TableHead>
              <TableHead>Date</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="file in bucketBackups" :key="file.name">
              <TableCell>{{ file.name }}</TableCell>
              <TableCell>{{ (file.size / 1024 / 1024).toFixed(2) }} Mo</TableCell>
              <TableCell>{{ new Date(file.lastModified).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <Button size="icon" variant="ghost" @click="restoreS3(file)">
                  <Plus class="w-5 h-5 text-primary" />
                </Button>
              </TableCell>
            </TableRow>
            <TableRow v-if="!bucketBackups.length && !loading">
              <TableCell colspan="4">Aucun backup trouvé.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

  </div>
</template>
