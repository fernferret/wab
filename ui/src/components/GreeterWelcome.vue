<script setup lang="ts">
import WelcomeItem from './WelcomeItem.vue'
import DocumentationIcon from './icons/IconDocumentation.vue'

import { HelloReply, HelloRequest } from '@/gen/greeter'
import { useApi } from '@/composables/api'
import { ref } from 'vue'

const api = useApi()

const data = ref({
  name: '',
  response: null as HelloReply | null,
})

const runGreet = () => {
  const req = {
    name: data.value.name,
  } as HelloRequest

  // Perform the gRPC call, assign the response item to show another element
  api.client.value?.Greet(req).then((item: HelloReply) => {
    data.value.response = item
  })
}

const clearGreet = () => {
  data.value.response = null
}
</script>

<template>
  <WelcomeItem>
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>Welcome!</template>
    WAB is running now. To test it out click the 'Greet' button below:<br />
    <input type="text" v-model="data.name" />
    <button @click="runGreet">Greet</button>
  </WelcomeItem>

  <WelcomeItem v-if="data.response">
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>{{ data.response.message }}</template>
    The header of this section was the reply from the gRPC server!<br />
    <br />
    <button @click="clearGreet">Reset</button>
    <button>Greet</button>
  </WelcomeItem>
</template>
