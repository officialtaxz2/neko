import { PluginObject } from 'vue'
import axios, { AxiosStatic } from 'axios'

declare global {
  const $http: AxiosStatic

  interface Window {
    $http: AxiosStatic
  }
}

declare module 'vue/types/vue' {
  interface Vue {
    $http: AxiosStatic
  }
}

const plugin: PluginObject<undefined> = {
  install(Vue) {
    // Neko does not use Axios's XSRF cookie/header mechanism. Keep unrelated
    // origin cookies out of requests, including the external About endpoints.
    axios.defaults.xsrfCookieName = ''
    window.$http = axios
    Vue.prototype.$http = window.$http
  },
}

export default plugin
