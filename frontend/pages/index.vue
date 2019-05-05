<template>
<<<<<<< Updated upstream
  <section class="container">
    <div>
      <logo />
      <h1 class="title">
        robocon-timer
      </h1>
      <h2 class="subtitle">
        Robocon timer
      </h2>
      <div class="links">
        <a href="https://nuxtjs.org/" target="_blank" class="button--green"
          >Documentation</a
        >
        <a
          href="https://github.com/nuxt/nuxt.js"
          target="_blank"
          class="button--grey"
          >GitHub</a
        >
=======
  <section id="app" class="hero is-light is-fullheight is-bold">
    <div class="hero-body">
      <div class="container has-text-centered">
        <form @submit.prevent="adminLogin">
          <input v-model="username" type="text" />
          <input v-model="password" type="password" />
          <button type="submit">Submit</button>
        </form>
>>>>>>> Stashed changes
      </div>
    </div>
  </section>
</template>

<script>
export default {
  components: {},
  data: function() {
    return {
      username: '',
      password: ''
    }
  },
  methods: {
    adminLogin: async function() {
      const resp = await this.$hub.login(this.username, this.password)
      console.log(resp)
    }
  }
}
</script>

<style lang="scss"></style>
