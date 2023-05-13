<script setup lang="ts">
import WelcomeItem from './WelcomeItem.vue'
import DocumentationIcon from './icons/IconDocumentation.vue'

import { HelloReply, HelloRequest, MultiHelloRequest } from '@/gen/greeter'
import { useApi } from '@/composables/api'
import { ref } from 'vue'
import { Subscription, type Observer } from 'rxjs'

const api = useApi()

const data = ref({
  name: '',
  response: null as HelloReply | null,
  responses: [] as HelloReply[],
  streamRequest: MultiHelloRequest.create(),
  subscription: undefined as undefined | Subscription,
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

// Perform a streaming gRPC call.
//
// We'll ask the server to greet us by the name
const runMultiGreet = () => {
  const observer = {
    next: (item: HelloReply) => {
      console.log(`Got message: ${item.message}`)
      data.value.responses.push(item)
    },
    error: (err: any) => {
      console.error(err)
    },
    complete: () => {
      console.log('Complete')
      data.value.subscription = undefined
    },
  } as Observer<HelloReply>

  // Add the name that was typed into the name field into the greeting
  data.value.streamRequest.request = {
    name: data.value.name,
  } as HelloRequest

  console.log('Sending stream request:')
  console.log(data.value.streamRequest)

  // Perform a streaming gRPC call. This is not the response, but a subscription
  // object that can be cancelled if the user wants to.
  data.value.subscription = api.client.value?.GreetMany(data.value.streamRequest).subscribe(observer)
}

const cancelMultiGreet = () => {
  data.value.subscription?.unsubscribe()
  data.value.subscription = undefined
}

const resetAll = () => {
  data.value.response = null
  data.value.responses = []
  data.value.streamRequest.qty = 1
}
</script>

<template>
  <WelcomeItem>
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>Welcome!</template>
    WAB is running now and there should be a gRPC server listening. To test it out click the 'Greet' button below:<br />
    <input type="text" v-model="data.name" placeholder="Enter your name" />
    <button type="button" @click="runGreet" :disabled="data.name.length === 0">Greet</button>
  </WelcomeItem>

  <WelcomeItem v-if="data.response">
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>{{ data.response.message }}</template>
    The header of this section was the reply from the gRPC server!<br />
    <br />
    Now let's see what a streaming call looks like. You can input 2 arguments here, one for the number of responses you
    want and another for how long to sleep between responses.<br />
    <br />
    Responses:<input
      type="number"
      min="0"
      max="600"
      v-model="data.streamRequest.qty"
      placeholder="Number of Responses"
    /><br />
    Sleep Time: <input type="number" min="0" max="600" v-model="data.streamRequest.sleepSeconds" /> seconds<br />
    <br />

    <button type="button" @click="runMultiGreet" :disabled="data.name.length === 0 || data.subscription !== undefined">
      Multi-Greet
    </button>
    <button type="button" @click="cancelMultiGreet" :disabled="data.subscription === undefined">Cancel</button>
  </WelcomeItem>
  <WelcomeItem v-for="(item, idx) in data.responses" v-bind:key="`resp_${idx}`">
    <template #icon>
      {{ idx }}
    </template>
    <template #heading>{{ item.message }}</template>
  </WelcomeItem>
</template>
