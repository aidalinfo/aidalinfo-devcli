<template>
  <Card class="max-w-lg mx-auto mt-10">
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
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

const accessKey = ref('')
const secretKey = ref('')

// Charger les credentials du localStorage au montage
onMounted(() => {
  accessKey.value = localStorage.getItem('s3_access_key') || ''
  secretKey.value = localStorage.getItem('s3_secret_key') || ''
})

function saveCredentials() {
  localStorage.setItem('s3_access_key', accessKey.value)
  localStorage.setItem('s3_secret_key', secretKey.value)
  alert('Clés S3 sauvegardées localement !')
}
</script>