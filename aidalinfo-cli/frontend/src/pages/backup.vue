<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ListBackupsWithCreds, RestoreMongoBackup, RestoreMySQLBackup, RestoreS3Backup, DownloadBackupToDirectory } from '../../wailsjs/go/main/App'
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
import { Plus, Download } from 'lucide-vue-next'
import MongoServerSelector from '@/components/MongoServerSelector.vue'
import { MongoServersManager, getMongoConnectionParams } from '@/utils/mongoServers'
import type { MongoServer } from '@/utils/mongoServers'
import MySQLServerSelector from '@/components/MySQLServerSelector.vue'
import { MySQLServersManager, getMySQLConnectionParams } from '@/utils/mysqlServers'
import type { MySQLServer } from '@/utils/mysqlServers'
import S3ServerSelector from '@/components/S3ServerSelector.vue'
import { S3ServersManager, getS3ConnectionParams } from '@/utils/s3Servers'
import type { S3Server } from '@/utils/s3Servers'

const PROJECTS = [
  { label: 'Sat&Lease V2', value: 'Sat&LeaseV2', mongo: 'backup/prod-sateleasev2/mongo/', mysql: 'backup/prod-sateleasev2/mysql/', bucket: 'backup/prod-sateleasev2/bucket/' },
  { label: 'LRD (Le Rédacteur)', value: 'LRD', mongo: 'backup/leredacteur/mongo/', mysql: 'backup/leredacteur/mysql/', bucket: 'backup/leredacteur/bucket/' },
  // Ajoute ici d'autres projets si besoin
]

const project = ref(PROJECTS[0].value)
const loading = ref(false)
const error = ref('')
const mongoBackups = ref<backend.BackupInfo[]>([])
const mysqlBackups = ref<backend.BackupInfo[]>([])
const bucketBackups = ref<backend.BackupInfo[]>([])

const PAGE_SIZE = 10
const mongoPage = ref(1)
const mysqlPage = ref(1)
const bucketPage = ref(1)

const mongoTotalPages = computed(() => Math.ceil(mongoBackups.value.length / PAGE_SIZE) || 1)
const mysqlTotalPages = computed(() => Math.ceil(mysqlBackups.value.length / PAGE_SIZE) || 1)
const bucketTotalPages = computed(() => Math.ceil(bucketBackups.value.length / PAGE_SIZE) || 1)

// Pour stocker le serveur MongoDB sélectionné
const selectedMongoServerId = ref<string>('')
const showMongoServerModal = ref(false)
const pendingMongoRestore = ref<backend.BackupInfo | null>(null)

// Pour stocker le serveur MySQL sélectionné
const selectedMySQLServerId = ref<string>('')
const showMySQLServerModal = ref(false)
const pendingMySQLRestore = ref<backend.BackupInfo | null>(null)
const mysqlDatabaseName = ref<string>('')

// Pour stocker le serveur S3 sélectionné
const selectedS3ServerId = ref<string>('')
const showS3ServerModal = ref(false)
const pendingS3Restore = ref<backend.BackupInfo | null>(null)

const pagedMongoBackups = computed(() => {
  const start = (mongoPage.value - 1) * PAGE_SIZE
  return mongoBackups.value.slice(start, start + PAGE_SIZE)
})
const pagedMySQLBackups = computed(() => {
  const start = (mysqlPage.value - 1) * PAGE_SIZE
  return mysqlBackups.value.slice(start, start + PAGE_SIZE)
})
const pagedBucketBackups = computed(() => {
  const start = (bucketPage.value - 1) * PAGE_SIZE
  return bucketBackups.value.slice(start, start + PAGE_SIZE)
})

function changeMongoPage(delta: number) {
  mongoPage.value = Math.max(1, Math.min(mongoPage.value + delta, mongoTotalPages.value))
}
function changeMySQLPage(delta: number) {
  mysqlPage.value = Math.max(1, Math.min(mysqlPage.value + delta, mysqlTotalPages.value))
}
function changeBucketPage(delta: number) {
  bucketPage.value = Math.max(1, Math.min(bucketPage.value + delta, bucketTotalPages.value))
}

watch(mongoBackups, () => { mongoPage.value = 1 })
watch(mysqlBackups, () => { mysqlPage.value = 1 })
watch(bucketBackups, () => { bucketPage.value = 1 })

watch(project, () => {
  fetchBackups()
})

function getCurrentProject() {
  return PROJECTS.find(p => p.value === project.value)!
}

function getS3Credentials(): backend.S3Credentials {
  // Utiliser le serveur S3 par défaut pour la récupération des backups
  const defaultServer = S3ServersManager.getDefaultServer()
  if (defaultServer) {
    return new backend.S3Credentials({
      accessKey: defaultServer.accessKey,
      secretKey: defaultServer.secretKey
    })
  }
  // Fallback sur les anciennes clés si pas de serveur configuré
  return new backend.S3Credentials({
    accessKey: localStorage.getItem('s3_access_key') || '',
    secretKey: localStorage.getItem('s3_secret_key') || ''
  })
}

async function fetchBackups() {
  loading.value = true
  error.value = ''
  mongoBackups.value = []
  mysqlBackups.value = []
  bucketBackups.value = []
  const creds = getS3Credentials()
  const current = getCurrentProject()
  try {
    const promises = [
      ListBackupsWithCreds(creds, current.mongo),
      ListBackupsWithCreds(creds, current.bucket)
    ]
    // Ajouter MySQL si le chemin existe
    if (current.mysql) {
      promises.push(ListBackupsWithCreds(creds, current.mysql))
    }
    const results = await Promise.all(promises)
    mongoBackups.value = results[0]
    bucketBackups.value = results[1]
    if (current.mysql && results[2]) {
      mysqlBackups.value = results[2]
    }
  } catch (e: any) {
    error.value = e.message || e.toString()
  } finally {
    loading.value = false
  }
}

onMounted(fetchBackups)

function selectMongoServerForRestore(file: backend.BackupInfo) {
  pendingMongoRestore.value = file
  showMongoServerModal.value = true
}

function selectMySQLServerForRestore(file: backend.BackupInfo) {
  pendingMySQLRestore.value = file
  mysqlDatabaseName.value = '' // Reset database name
  showMySQLServerModal.value = true
}

async function restoreMongo() {
  if (!pendingMongoRestore.value || !selectedMongoServerId.value) {
    toast.error('Veuillez sélectionner un serveur MongoDB')
    return
  }
  
  const server = MongoServersManager.getServer(selectedMongoServerId.value)
  if (!server) {
    toast.error('Serveur MongoDB introuvable')
    return
  }
  
  const creds = getS3Credentials()
  const current = getCurrentProject()
  const { mongoHost, mongoPort, mongoUser, mongoPassword } = getMongoConnectionParams(server)
  
  toast.info(`Restauration MongoDB sur ${server.name} en cours...`)
  try {
    await RestoreMongoBackup(
      creds,
      current.mongo + pendingMongoRestore.value.name,
      mongoHost,
      mongoPort,
      mongoUser,
      mongoPassword
    )
    toast.success(`Restauration MongoDB sur ${server.name} terminée avec succès !`)
    showMongoServerModal.value = false
    pendingMongoRestore.value = null
    selectedMongoServerId.value = ''
  } catch (e: any) {
    toast.error('Erreur restauration MongoDB : ' + (e.message || e.toString()))
  }
}

async function restoreMySQL() {
  if (!pendingMySQLRestore.value || !selectedMySQLServerId.value || !mysqlDatabaseName.value) {
    toast.error('Veuillez sélectionner un serveur MySQL et entrer le nom de la base de données')
    return
  }
  
  const server = MySQLServersManager.getServer(selectedMySQLServerId.value)
  if (!server) {
    toast.error('Serveur MySQL introuvable')
    return
  }
  
  const creds = getS3Credentials()
  const current = getCurrentProject()
  const { mysqlHost, mysqlPort, mysqlUser, mysqlPassword } = getMySQLConnectionParams(server)
  
  toast.info(`Restauration MySQL sur ${server.name} en cours...`)
  try {
    await RestoreMySQLBackup(
      creds,
      current.mysql + pendingMySQLRestore.value.name,
      mysqlHost,
      mysqlPort,
      mysqlUser,
      mysqlPassword,
      mysqlDatabaseName.value
    )
    toast.success(`Restauration MySQL sur ${server.name} terminée avec succès !`)
    showMySQLServerModal.value = false
    pendingMySQLRestore.value = null
    selectedMySQLServerId.value = ''
    mysqlDatabaseName.value = ''
  } catch (e: any) {
    toast.error('Erreur restauration MySQL : ' + (e.message || e.toString()))
  }
}

function selectS3ServerForRestore(file: backend.BackupInfo) {
  pendingS3Restore.value = file
  showS3ServerModal.value = true
}

async function restoreS3() {
  if (!pendingS3Restore.value || !selectedS3ServerId.value) {
    toast.error('Veuillez sélectionner un serveur S3/MinIO')
    return
  }
  
  const server = S3ServersManager.getServer(selectedS3ServerId.value)
  if (!server) {
    toast.error('Serveur S3/MinIO introuvable')
    return
  }
  
  const cloudCreds = getS3Credentials()
  const current = getCurrentProject()
  const { host, port, accessKey, secretKey, region, useHttps } = getS3ConnectionParams(server)
  const localCreds = { accessKey, secretKey }
  
  toast.info(`Restauration S3 sur ${server.name} en cours...`)
  try {
    // @ts-ignore
    await RestoreS3Backup(
      cloudCreds,
      localCreds,
      current.bucket + pendingS3Restore.value.name,
      host,
      port,
      region,
      useHttps
    )
    toast.success(`Restauration S3 sur ${server.name} terminée avec succès !`)
    showS3ServerModal.value = false
    pendingS3Restore.value = null
    selectedS3ServerId.value = ''
  } catch (e: any) {
    toast.error('Erreur restauration S3 : ' + (e.message || e.toString()))
  }
}

// Download functions
async function downloadMongoBackup(file: backend.BackupInfo) {
  const creds = getS3Credentials()
  const current = getCurrentProject()
  const s3Path = current.mongo + file.name
  
  toast.info('Sélectionnez un dossier de destination...')
  try {
    const downloadPath = await DownloadBackupToDirectory(creds, s3Path, file.name)
    toast.success(`Backup MongoDB téléchargé avec succès dans : ${downloadPath}`)
  } catch (e: any) {
    if (e.message && e.message.includes('aucun dossier sélectionné')) {
      // L'utilisateur a annulé, pas d'erreur
      return
    }
    toast.error('Erreur téléchargement MongoDB : ' + (e.message || e.toString()))
  }
}

async function downloadMySQLBackup(file: backend.BackupInfo) {
  const creds = getS3Credentials()
  const current = getCurrentProject()
  const s3Path = current.mysql + file.name
  
  toast.info('Sélectionnez un dossier de destination...')
  try {
    const downloadPath = await DownloadBackupToDirectory(creds, s3Path, file.name)
    toast.success(`Backup MySQL téléchargé avec succès dans : ${downloadPath}`)
  } catch (e: any) {
    if (e.message && e.message.includes('aucun dossier sélectionné')) {
      // L'utilisateur a annulé, pas d'erreur
      return
    }
    toast.error('Erreur téléchargement MySQL : ' + (e.message || e.toString()))
  }
}

async function downloadS3Backup(file: backend.BackupInfo) {
  const creds = getS3Credentials()
  const current = getCurrentProject()
  const s3Path = current.bucket + file.name
  
  toast.info('Sélectionnez un dossier de destination...')
  try {
    const downloadPath = await DownloadBackupToDirectory(creds, s3Path, file.name)
    toast.success(`Backup S3 téléchargé avec succès dans : ${downloadPath}`)
  } catch (e: any) {
    if (e.message && e.message.includes('aucun dossier sélectionné')) {
      // L'utilisateur a annulé, pas d'erreur
      return
    }
    toast.error('Erreur téléchargement S3 : ' + (e.message || e.toString()))
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
            <TableRow v-for="file in pagedMongoBackups" :key="file.name">
              <TableCell>{{ file.name }}</TableCell>
              <TableCell>{{ (file.size / 1024 / 1024).toFixed(2) }} Mo</TableCell>
              <TableCell>{{ new Date(file.lastModified).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <div class="flex gap-1 justify-end">
                  <Button size="icon" variant="ghost" @click="downloadMongoBackup(file)" title="Télécharger">
                    <Download class="w-5 h-5 text-blue-600" />
                  </Button>
                  <Button size="icon" variant="ghost" @click="selectMongoServerForRestore(file)" title="Restaurer">
                    <Plus class="w-5 h-5 text-primary" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <TableRow v-if="!mongoBackups.length && !loading">
              <TableCell colspan="4">Aucun backup trouvé.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between mt-2" v-if="mongoTotalPages > 1">
          <Button size="sm" variant="outline" :disabled="mongoPage === 1" @click="changeMongoPage(-1)">Précédent</Button>
          <span>Page {{ mongoPage }} / {{ mongoTotalPages }}</span>
          <Button size="sm" variant="outline" :disabled="mongoPage === mongoTotalPages" @click="changeMongoPage(1)">Suivant</Button>
        </div>
      </CardContent>
    </Card>

    <Card v-if="getCurrentProject().mysql">
      <CardHeader>
        <CardTitle>Backups MySQL</CardTitle>
        <CardDescription>Liste des backups MySQL pour le projet sélectionné.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableCaption>Backups MySQL disponibles</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Nom du backup</TableHead>
              <TableHead>Taille</TableHead>
              <TableHead>Date</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="file in pagedMySQLBackups" :key="file.name">
              <TableCell>{{ file.name }}</TableCell>
              <TableCell>{{ (file.size / 1024 / 1024).toFixed(2) }} Mo</TableCell>
              <TableCell>{{ new Date(file.lastModified).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <div class="flex gap-1 justify-end">
                  <Button size="icon" variant="ghost" @click="downloadMySQLBackup(file)" title="Télécharger">
                    <Download class="w-5 h-5 text-blue-600" />
                  </Button>
                  <Button size="icon" variant="ghost" @click="selectMySQLServerForRestore(file)" title="Restaurer">
                    <Plus class="w-5 h-5 text-primary" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <TableRow v-if="!mysqlBackups.length && !loading">
              <TableCell colspan="4">Aucun backup trouvé.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between mt-2" v-if="mysqlTotalPages > 1">
          <Button size="sm" variant="outline" :disabled="mysqlPage === 1" @click="changeMySQLPage(-1)">Précédent</Button>
          <span>Page {{ mysqlPage }} / {{ mysqlTotalPages }}</span>
          <Button size="sm" variant="outline" :disabled="mysqlPage === mysqlTotalPages" @click="changeMySQLPage(1)">Suivant</Button>
        </div>
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
            <TableRow v-for="file in pagedBucketBackups" :key="file.name">
              <TableCell>{{ file.name }}</TableCell>
              <TableCell>{{ (file.size / 1024 / 1024).toFixed(2) }} Mo</TableCell>
              <TableCell>{{ new Date(file.lastModified).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <div class="flex gap-1 justify-end">
                  <Button size="icon" variant="ghost" @click="downloadS3Backup(file)" title="Télécharger">
                    <Download class="w-5 h-5 text-blue-600" />
                  </Button>
                  <Button size="icon" variant="ghost" @click="selectS3ServerForRestore(file)" title="Restaurer">
                    <Plus class="w-5 h-5 text-primary" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <TableRow v-if="!bucketBackups.length && !loading">
              <TableCell colspan="4">Aucun backup trouvé.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between mt-2" v-if="bucketTotalPages > 1">
          <Button size="sm" variant="outline" :disabled="bucketPage === 1" @click="changeBucketPage(-1)">Précédent</Button>
          <span>Page {{ bucketPage }} / {{ bucketTotalPages }}</span>
          <Button size="sm" variant="outline" :disabled="bucketPage === bucketTotalPages" @click="changeBucketPage(1)">Suivant</Button>
        </div>
      </CardContent>
    </Card>

  </div>

  <!-- Modal pour sélectionner le serveur MongoDB -->
  <div v-if="showMongoServerModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-md w-full">
      <h3 class="text-lg font-semibold mb-4">Sélectionner le serveur MongoDB de destination</h3>
      
      <div v-if="pendingMongoRestore" class="mb-4 p-3 bg-gray-50 rounded">
        <p class="text-sm text-gray-600">Backup à restaurer :</p>
        <p class="font-medium">{{ pendingMongoRestore.name }}</p>
      </div>
      
      <MongoServerSelector
        v-model="selectedMongoServerId"
        :auto-select-default="true"
        :show-details="true"
      />
      
      <div class="flex justify-end gap-3 mt-6">
        <Button variant="outline" @click="showMongoServerModal = false">
          Annuler
        </Button>
        <Button 
          @click="restoreMongo" 
          :disabled="!selectedMongoServerId"
        >
          Restaurer
        </Button>
      </div>
    </div>
  </div>

  <!-- Modal pour sélectionner le serveur MySQL -->
  <div v-if="showMySQLServerModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-md w-full">
      <h3 class="text-lg font-semibold mb-4">Sélectionner le serveur MySQL de destination</h3>
      
      <div v-if="pendingMySQLRestore" class="mb-4 p-3 bg-gray-50 rounded">
        <p class="text-sm text-gray-600">Backup à restaurer :</p>
        <p class="font-medium">{{ pendingMySQLRestore.name }}</p>
      </div>
      
      <MySQLServerSelector
        v-model="selectedMySQLServerId"
        :auto-select-default="true"
        :show-details="true"
      />
      
      <div class="mt-4">
        <Label for="mysql-database-name">Nom de la base de données cible *</Label>
        <Input
          v-model="mysqlDatabaseName"
          id="mysql-database-name"
          placeholder="Ex: my_database"
          required
        />
        <p class="text-xs text-muted-foreground mt-1">
          La base de données sera créée si elle n'existe pas
        </p>
      </div>
      
      <div class="flex justify-end gap-3 mt-6">
        <Button variant="outline" @click="showMySQLServerModal = false">
          Annuler
        </Button>
        <Button 
          @click="restoreMySQL" 
          :disabled="!selectedMySQLServerId || !mysqlDatabaseName"
        >
          Restaurer
        </Button>
      </div>
    </div>
  </div>

  <!-- Modal pour sélectionner le serveur S3 -->
  <div v-if="showS3ServerModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-md w-full">
      <h3 class="text-lg font-semibold mb-4">Sélectionner le serveur S3/MinIO de destination</h3>
      
      <div v-if="pendingS3Restore" class="mb-4 p-3 bg-gray-50 rounded">
        <p class="text-sm text-gray-600">Backup à restaurer :</p>
        <p class="font-medium">{{ pendingS3Restore.name }}</p>
      </div>
      
      <S3ServerSelector
        v-model="selectedS3ServerId"
        :auto-select-default="true"
        :show-details="true"
      />
      
      <div class="flex justify-end gap-3 mt-6">
        <Button variant="outline" @click="showS3ServerModal = false">
          Annuler
        </Button>
        <Button 
          @click="restoreS3" 
          :disabled="!selectedS3ServerId"
        >
          Restaurer
        </Button>
      </div>
    </div>
  </div>
</template>
