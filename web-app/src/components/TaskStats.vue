<template>
  <div class="grid">
    <!-- <v-icon class="c1" color="green">check_circle</v-icon> -->
    <!-- <v-icon class="c2" color="orange darken-2">error</v-icon> -->
    <!-- <v-icon class="c3" color="red darken-2">notification_important</v-icon> -->
    <icon class="c1" color="green" name="check"/>
    <icon class="c2" color="deep-orange" name="error"/>
    <icon class="c3" color="red darken-2" name="notification"/>
    <label class="c1">{{ successes }}</label>
    <label class="c2">{{ errors }}</label>
    <label class="c3">{{ crashes }}</label>
  </div>
</template>

<script>
export default {
  name: 'TaskStats',
  props: {
    task: Object
  },
  computed: {
    stats () {
      return this.task.stats || []
    },
    successes () {
      return this.stats.filter(i => i.status === 0).length
    },
    errors () {
      return this.stats.filter(i => i.status > 0).length
    },
    crashes () {
      return this.stats.filter(i => i.crashed).length
    }
  }
}
</script>

<style lang="scss" scoped>
.icon {
  grid-row: 1 / 2;
  background-color: #fff;
  border-radius: 50%;
}
label {
  grid-row: 2 / 3;
  text-align: center;
  font-size: 13px;
  margin-top: 2px;
}
.c1 {
  grid-column:  1 / 3;
  z-index: 3;
}
.c2 {
  grid-column:  2 / 5;
  z-index: 2;
}
.c3 {
  grid-column:  4 / 6;
  z-index: 1;
}
.grid {
  display: grid;
  grid-template-columns: auto 5px auto 5px auto;
}
</style>
