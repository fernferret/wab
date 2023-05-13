<script setup lang="ts">
import WelcomeItem from './WelcomeItem.vue'
import DocumentationIcon from './icons/IconDocumentation.vue'

import { HelloReply, HelloRequest, MultiHelloRequest } from '@/gen/greeter'
import { useApi } from '@/composables/api'
import { ref, onBeforeMount } from 'vue'
import { type Observer } from 'rxjs'

const api = useApi()

const data = ref({
  name: '',
  response: null as HelloReply | null,
  responses: [] as HelloReply[],
  streamRequest: MultiHelloRequest.create(),
  abort: undefined as undefined | AbortController,
  errMsg: '',
  finished: false,
})

onBeforeMount(() => {
  // Perform a reset when the page loads. This will set some rational defaults
  // for the streaming example that will send 3 messages at a 3 second interval.
  resetAll()
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
  // Reset the error message and responses
  data.value.errMsg = ''
  data.value.responses = []

  const observer = {
    next: (item: HelloReply) => {
      console.log(`Got message: ${item.message}`)
      data.value.responses.push(item)
    },
    error: (err: any) => {
      data.value.abort = undefined
      // When the AbortController is used, an error with the string 'Aborted' is
      // raises. In all other cases we'll get a proper error back from the
      // GRPC-Web implementation that will have GRPC-like details.
      if (err === 'Aborted') {
        console.warn('Client cancelled request')
        return
      }
      if (err && err.code !== undefined && err.message !== undefined) {
        // A GRPC error was seen, this is likely.
        const msg = `GRPC Error (${err.code}): ${err.message}`
        console.error(msg)
        data.value.errMsg = msg
      } else {
        data.value.errMsg = 'An unknown error has occurred, see console.'
      }
      // Always log the error, in case something internal went wrong.
      console.error(err)
    },
    complete: () => {
      console.log('Complete')
      data.value.abort = undefined

      // Once a user has gone all the way through the demo once, show the final
      // documentation block which instructs them to read the documentation,
      // repo and check out grpcui.
      data.value.finished = true
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
  data.value.abort = new AbortController()
  api.client.value?.GreetMany(data.value.streamRequest, undefined, data.value.abort.signal).subscribe(observer)
}

// Cancel an in-progress streaming call.
const cancelMultiGreet = () => {
  data.value.abort?.abort()
  data.value.abort = undefined
}

const resetAll = () => {
  data.value.response = null
  data.value.responses = []
  data.value.streamRequest.qty = 3
  data.value.streamRequest.sleepSeconds = 3
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
    <button type="button" @click="resetAll" :disabled="data.response === null && data.responses.length === 0">
      Reset
    </button>
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
      :disabled="data.abort !== undefined"
      min="0"
      max="600"
      v-model="data.streamRequest.qty"
      placeholder="Number of Responses"
    /><br />
    Sleep Time:
    <input
      :disabled="data.abort !== undefined"
      type="number"
      min="0"
      max="600"
      v-model="data.streamRequest.sleepSeconds"
    />
    seconds<br />
    <br />

    <button type="button" @click="runMultiGreet" :disabled="data.name.length === 0 || data.abort !== undefined">
      Multi-Greet
    </button>
    <button type="button" @click="cancelMultiGreet" :disabled="data.abort === undefined">Cancel</button>
    <div v-if="data.errMsg">
      <br />
      <span class="errmsg">Error: {{ data.errMsg }}</span>
    </div>
  </WelcomeItem>
  <WelcomeItem v-for="(item, idx) in data.responses" v-bind:key="`resp_${idx}`">
    <template #icon>
      {{ idx }}
    </template>
    <template #heading>{{ item.message }}</template>
  </WelcomeItem>
  <WelcomeItem v-if="data.finished">
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>What's next...</template>
    Huzzah! You've gone all the way through. Here's what I recommend you do next:
    <ul>
      <li>
        Check out the embedded GRPC tester called <a href="#"><code>grpcui</code></a
        >.
      </li>
      <li>
        Have a look at the defined
        <a href="https://github.com/fernferret/wab/blob/master/proto/greeter.proto"><code>greeter.proto</code></a> file
        to see how the proto and <code>gRPC</code> services are defined.
      </li>
      <li>
        Glance at the
        <a href="https://github.com/fernferret/wab/blob/master/proto/greeter.proto"><code>Makefile</code></a> to see how
        everything is built.
      </li>
    </ul>
  </WelcomeItem>
</template>

<style scoped>
.errmsg {
  font-weight: bold;
  color: #ff8c00;
}
</style>
