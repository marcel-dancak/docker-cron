<template>
  <v-layout>
    <v-layout class="align-center px-3 grow">
      <v-list-item two-line class="px-0 py-0">
        <v-list-item-content class="py-0">
          <v-list-item-title class="headline">
            <router-link :to="{ name: 'task', params: { name: task.name } }">
              {{ task.name }}
            </router-link>
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ task.schedule }}
          </v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
    </v-layout>

    <v-layout class="last text--secondary column justify-start shrink mx-2 my-2">
      <h5>Last</h5>
      <date-field :value="last && last.start_time"/>
      <time-field :value="last && last.start_time"/>
      <task-result :stats="last" small/>
    </v-layout>

    <v-layout class="next text--secondary column justify-start shrink mx-2 my-2">
      <h5>Next</h5>
      <date-field :value="task.next"/>
      <time-field :value="task.next"/>
      <v-spacer/>
      <v-progress-linear v-if="running" indeterminate/>
      <v-spacer/>
    </v-layout>
  </v-layout>
</template>

<script>
import findLast from 'lodash/findLast'
import TaskResult from '@/components/TaskResult.vue'

export default {
  name: 'TaskCard',
  components: { TaskResult },
  props: {
    task: Object
  },
  computed: {
    stats () {
      return this.task.stats || []
    },
    last () {
      return findLast(this.stats, r => !r.running)
    },
    running () {
      return this.stats.some(i => i.running)
    }
  },
  methods: {
    runTask (task) {
      this.$http.post(`/api/run/${task.name}`)
    }
  }
}
</script>

<style lang="scss" scoped>
::v-deep {
  a {
    text-decoration: none;
  }
}
.last, .next {
  width: 100px;
  .v-icon {
    margin-top: 2px;
    margin-bottom: 2px;
  }
}
</style>
