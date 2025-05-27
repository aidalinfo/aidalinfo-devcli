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
      <CardTitle>Paramètres connexion MongoDB</CardTitle>
      <CardDescription>
        Configurez la connexion à votre base MongoDB (pour restauration ou dump).
      </CardDescription>
    </CardHeader>
    <CardContent>
      <form @submit.prevent="saveMongo" class="space-y-6">
        <div>
          <Label for="mongo-host">Adresse</Label>
          <Input id="mongo-host" v-model="mongoHost" placeholder="localhost" />
        </div>
        <div>
          <Label for="mongo-port">Port</Label>
          <Input id="mongo-port" v-model="mongoPort" type="number" placeholder="27017" />
        </div>
        <div>
          <Label for="mongo-user">Utilisateur</Label>
          <Input id="mongo-user" v-model="mongoUser" placeholder="" />
        </div>
        <div>
          <Label for="mongo-password">Mot de passe</Label>
          <Input id="mongo-password" v-model="mongoPassword" type="password" placeholder="" />
        </div>
        <Button type="submit" class="w-full">Sauvegarder</Button>
      </form>
    </CardContent>
  </Card>

  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Paramètres avancés S3</CardTitle>
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

const accessKey = ref('')
const secretKey = ref('')
const mongoHost = ref('localhost')
const mongoPort = ref('27017')
const mongoUser = ref('')
const mongoPassword = ref('')
const s3Host = ref('localhost')
const s3Port = ref('9000')
const s3Region = ref('fr-par')
const s3LocalAccessKey = ref('')
const s3LocalSecretKey = ref('')

// Charger les credentials du localStorage au montage
onMounted(() => {
  accessKey.value = localStorage.getItem('s3_access_key') || ''
  secretKey.value = localStorage.getItem('s3_secret_key') || ''
  mongoHost.value = localStorage.getItem('mongo_host') || 'localhost'
  mongoPort.value = localStorage.getItem('mongo_port') || '27017'
  mongoUser.value = localStorage.getItem('mongo_user') || ''
  mongoPassword.value = localStorage.getItem('mongo_password') || ''
  s3Host.value = localStorage.getItem('s3_host') || 'localhost'
  s3Port.value = localStorage.getItem('s3_port') || '9000'
  s3Region.value = localStorage.getItem('s3_region') || 'fr-par'
  s3LocalAccessKey.value = localStorage.getItem('s3local_access_key') || ''
  s3LocalSecretKey.value = localStorage.getItem('s3local_secret_key') || ''
})

function saveCredentials() {
  localStorage.setItem('s3_access_key', accessKey.value)
  localStorage.setItem('s3_secret_key', secretKey.value)
  alert('Clés S3 sauvegardées localement !')
}
function saveMongo() {
  localStorage.setItem('mongo_host', mongoHost.value)
  localStorage.setItem('mongo_port', mongoPort.value)
  localStorage.setItem('mongo_user', mongoUser.value)
  localStorage.setItem('mongo_password', mongoPassword.value)
  alert('Paramètres Mongo sauvegardés !')
}
function saveS3Advanced() {
  localStorage.setItem('s3_host', s3Host.value)
  localStorage.setItem('s3_port', s3Port.value)
  localStorage.setItem('s3_region', s3Region.value)
  localStorage.setItem('s3local_access_key', s3LocalAccessKey.value)
  localStorage.setItem('s3local_secret_key', s3LocalSecretKey.value)
  alert('Paramètres S3 avancés sauvegardés !')
}
</script>