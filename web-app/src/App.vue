<template>
  <v-app class="desktop">
    <v-toolbar class="app-toolbar shrink" dark>
      <v-spacer/>
      <div class="d-flex align-center">
        <v-img
          alt="Vuetify Logo"
          class="shrink mr-2"
          contain
          src="./assets/docker_logo.svg"
          transition="scale-transition"
          width="80"
        />
        <h2>Task Manager</h2>
      </div>
      <v-spacer/>
    </v-toolbar>
    <div class="page-layout fill-height">
      <svg width="100%" height="100%" viewBox="0,0,100,100" class="bg" preserveAspectRatio="none">
        <!-- <path d="M0,0 L 46,0 L 53,100 L 0,100 Z " fill="#263238" fill-opacity="0.8"/> -->
        <path d="M80,0 L 100,0 L 100,100 L 65,100 Z " fill="#263238" fill-opacity="0.8"/>
        <!-- <path d="M 0,20 L 45,0" stroke="#263238"/> -->
      </svg>
      <router-view class="app-content"/>
    </div>
  </v-app>
</template>

<script>

export default {
  name: 'App',
  mounted () {
    document.body.parentElement.classList.add('desktop')
  }
}
</script>

<style lang="scss">
html.desktop {
  &, body, .v-application, .v-application--wrap {
    overflow: hidden!important;
    height: 100%;
  }
  .scrollable {
    overflow: auto;
  }
  // .v-application, .v-application--wrap {
  //   overflow: hidden!important;
  //   min-height: 0;
  //   max-height: 90%;
  //   height: 90vh;
  // }
}
</style>

<style lang="scss" scoped>
.app-content {
  min-height: 0;
  flex: 1 1 auto;
}
.page-layout {
  min-height: 0;
  background-color: #eee;
  display: grid;
  grid-template-columns: 1fr 1200px 1fr;
  // justify-content: center;

  @media (max-width: 1200px) {
    grid-template-columns: 0 100% 0;
    .tasks-card {
      // justify-self: center;
    }
  }

  @media (min-width: 960px) {
    .tasks-card {
      // width: 600px;
    }
  }

  .bg {
    grid-column: 1 / 3;
    grid-column: 2 / 4;
    grid-row: 1 / 2;
  }
}
</style>
