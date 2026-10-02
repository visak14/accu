import { createApp } from 'vue'
import { Quasar, Notify, Dialog, Loading } from 'quasar'

// Import icon libraries
import '@quasar/extras/material-icons/material-icons.css'
import '@quasar/extras/material-icons-outlined/material-icons-outlined.css'
import '@quasar/extras/material-icons-round/material-icons-round.css'

// Import Quasar css
import 'quasar/src/css/index.sass'

import App from './App.vue'

const myApp = createApp(App)

myApp.use(Quasar, {
  plugins: {
    Notify,
    Dialog,
    Loading
  },
  config: {
    notify: {
      position: 'top-right',
      timeout: 3000
    },
    brand: {
      primary: '#0284c7',
      secondary: '#0f766e',
      accent: '#8b5cf6',
      dark: '#0f172a',
      positive: '#10b981',
      negative: '#ef4444',
      info: '#3b82f6',
      warning: '#f59e0b'
    }
  }
})

myApp.mount('#app')
