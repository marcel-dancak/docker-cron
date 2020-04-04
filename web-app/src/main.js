import Vue from 'vue'
import DesktopApp from './App.vue'
import MobileApp from './MobileApp.vue'
import WebsocketConnection from './ws'
import http from './http'
import routers from './router'
import vuetify from './plugins/vuetify'
import './global-components'
import parseISO from 'date-fns/parseISO'
import format from 'date-fns/format'

const mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent)

Vue.config.productionTip = false
Vue.prototype.$http = http

Vue.filter('datetime', v => format(parseISO(v), 'dd.MM.yyyy HH:mm'))
Vue.filter('date', v => v ? format(new Date(v), 'dd.MM.yyyy') : '-')
Vue.filter('time', v => v ? format(new Date(v), 'HH:mm') : '-')

new Vue({
  router: mobile ? routers.mobile : routers.desktop,
  vuetify,
  data () {
    return {
      tasks: {}
    }
  },
  mounted () {
    this.init()
  },
  methods: {
    onTaskStatusUpdated (e) {
      console.log(e.type)
      const { task } = e
      this.tasks[task.name] = task
    },
    async init () {
      await this.fetchTasks()
      Vue.prototype.$ws = WebsocketConnection(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`)
      this.$once('hook:beforeDestroy', this.$ws.bind('TaskStarted', this.onTaskStatusUpdated))
      this.$once('hook:beforeDestroy', this.$ws.bind('TaskFinished', this.onTaskStatusUpdated))
    },
    async fetchTasks () {
      const { data } = await this.$http.get('/api/tasks/')
      this.tasks = data
    }
  },
  render: h => h(mobile ? MobileApp : DesktopApp)
}).$mount('#app')
