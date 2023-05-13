import { ref, onBeforeMount } from 'vue'

import { GreeterClientImpl, GrpcWebImpl } from '@/gen/greeter'

export const useApi = () => {
  const client = ref(null as null | GreeterClientImpl)

  onBeforeMount(() => {
    let url = 'grpc'
    if (import.meta.env.VITE_API_ADDR !== undefined) {
      url = `${import.meta.env.VITE_API_ADDR}/grpc`
    }
    const rpc = new GrpcWebImpl(url, {
      debug: import.meta.env.DEV,
    })

    client.value = new GreeterClientImpl(rpc)
  })

  return {
    client,
  }
}
