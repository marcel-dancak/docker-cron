<template>
  <v-app class="mobile">
    <v-toolbar class="app-toolbar shrink" dark>
      <v-btn
        v-if="$route.path !== '/'"
        @click="$router.back()"
        icon
      >
        <v-icon>arrow_back</v-icon>
      </v-btn>
      <v-spacer/>
      <div class="d-flex align-center">
        <v-img
          alt="Vuetify Logo"
          class="shrink mr-2"
          contain
          src="./assets/docker_logo.svg"
          width="80"
        />
        <h2>Task Manager</h2>
      </div>
      <v-spacer/>
    </v-toolbar>
    <!-- <div class="switch-container fill-height"> -->
      <!-- <transition name="switch"> -->
        <router-view/>
      <!-- </transition> -->
    <!-- </div> -->
  </v-app>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style lang="scss">
/* Slide from the left */
.slide-enter-active, .slide-leave-active {
  transition: transform .4s cubic-bezier(.25,.8,.5,1);
}
.slide-enter, .slide-leave-to {
  transform: translate3d(-100%, 0, 0);
}
.slide-enter-active, .slide-leave-active {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}
.switch-enter, .switch-leave-to {
  transform: translate(-30%, 0);
  opacity: 0;
}
.switch-leave, .switch-enter-to {
  transform: translate(0, 0);
  opacity: 1;
}
.switch-enter-active, .switch-leave-active {
  transition: all 0.5s ease;
}
.switch-enter-active, .switch-leave-active {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}

.switch-container {
  position: relative;
}
// .switch-active {
//   will-change: height;
//   transition: all 1.5s ease;
//   overflow: hidden;
// }
.v-application--wrap {
  background-color: #e0e0e0;
}
</style>

<style lang="scss" scoped>
.app-toolbar {
  // position: sticky;
  top: 0;
  z-index: 1;
}
</style>
