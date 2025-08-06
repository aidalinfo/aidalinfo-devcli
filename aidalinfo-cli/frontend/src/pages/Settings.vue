<template>
  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Paramètres S3 Backup</CardTitle>
      <CardDescription>
        Saisissez vos identifiants S3 pour accéder aux backups.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <form @submit.prevent="saveCredentials" class="space-y-6">
        <div>
          <Label for="access-key">Access Key</Label>
          <Input id="access-key" v-model="accessKey" autocomplete="off" />
        </div>
        <div>
          <Label for="secret-key">Secret Key</Label>
          <Input id="secret-key" v-model="secretKey" type="password" autocomplete="off" />
        </div>
        <Button type="submit" class="w-full">Sauvegarder</Button>
      </form>
    </CardContent>
  </Card>

  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Serveurs MongoDB</CardTitle>
      <CardDescription>
        Gérez vos différents serveurs MongoDB pour la restauration et les dumps.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <MongoServerManager />
    </CardContent>
  </Card>

  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Paramètres locaux S3</CardTitle>
      <CardDescription>
        Configurez l'hôte, le port et la région S3 pour la restauration locale (MinIO ou autre).
      </CardDescription>
    </CardHeader>
    <CardContent>
      <form @submit.prevent="saveS3Advanced" class="space-y-6">
        <div>
          <Label for="s3-host">Hôte</Label>
          <Input id="s3-host" v-model="s3Host" placeholder="localhost" />
        </div>
        <div>
          <Label for="s3-port">Port</Label>
          <Input id="s3-port" v-model="s3Port" type="number" placeholder="9000" />
        </div>
        <div>
          <Label for="s3-region">Région</Label>
          <Input id="s3-region" v-model="s3Region" placeholder="fr-par" />
        </div>
        <div>
          <Label for="s3-use-https">Protocole</Label>
          <Select v-model="s3UseHttps">
            <SelectTrigger id="s3-use-https">
              <SelectValue placeholder="Sélectionner le protocole" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="true">HTTPS (sécurisé)</SelectItem>
              <SelectItem value="false">HTTP</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label for="s3local-access-key">Access Key (local S3)</Label>
          <Input id="s3local-access-key" v-model="s3LocalAccessKey" autocomplete="off" />
        </div>
        <div>
          <Label for="s3local-secret-key">Secret Key (local S3)</Label>
          <Input id="s3local-secret-key" v-model="s3LocalSecretKey" type="password" autocomplete="off" />
        </div>
        <Button type="submit" class="w-full">Sauvegarder</Button>
      </form>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import MongoServerManager from '@/components/MongoServerManager.vue'

const accessKey = ref('')
const secretKey = ref('')
const s3Host = ref('localhost')
const s3Port = ref('9000')
const s3Region = ref('fr-par')
const s3UseHttps = ref('false')
const s3LocalAccessKey = ref('')
const s3LocalSecretKey = ref('')

// Charger les credentials du localStorage au montage
onMounted(() => {
  accessKey.value = localStorage.getItem('s3_access_key') || ''
  secretKey.value = localStorage.getItem('s3_secret_key') || ''
  s3Host.value = localStorage.getItem('s3_host') || 'localhost'
  s3Port.value = localStorage.getItem('s3_port') || '9000'
  s3Region.value = localStorage.getItem('s3_region') || 'fr-par'
  s3UseHttps.value = localStorage.getItem('s3_use_https') || 'false'
  s3LocalAccessKey.value = localStorage.getItem('s3local_access_key') || ''
  s3LocalSecretKey.value = localStorage.getItem('s3local_secret_key') || ''
})

function saveCredentials() {
  localStorage.setItem('s3_access_key', accessKey.value)
  localStorage.setItem('s3_secret_key', secretKey.value)
  toast.success('Clés S3 sauvegardées localement !')
}

function saveS3Advanced() {
  localStorage.setItem('s3_host', s3Host.value)
  localStorage.setItem('s3_port', s3Port.value)
  localStorage.setItem('s3_region', s3Region.value)
  localStorage.setItem('s3_use_https', s3UseHttps.value)
  localStorage.setItem('s3local_access_key', s3LocalAccessKey.value)
  localStorage.setItem('s3local_secret_key', s3LocalSecretKey.value)
  toast.success('Paramètres S3 avancés sauvegardés !')
}
</script>