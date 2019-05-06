<template>
  <section id="app" class="hero is-light is-fullheight is-bold">
    <div class="hero-body">
      <div class="container has-text-centered">
        <h2 id="message" class="title is-6">{{ title }}</h2>

        <div v-if="timer" class="timer">
          <span class="time">{{ minutesText }}</span>
          <span class="time">:</span>
          <span class="time">{{ secondsText }}</span>
        </div>

        <div v-if="!timer" class="timer timer--input">
          <input v-model.number="minutes" class="time" />
          <span class="time">:</span>
          <input v-model.number="seconds" class="time" />
        </div>

        <div id="buttons">
          <!--     Start TImer -->
          <button
            v-if="!timer"
            id="start"
            class="button is-dark is-large"
            @click="startTimer"
          >
            <i class="far fa-play-circle"></i>
          </button>
          <!--     Pause Timer -->
          <button
            v-if="timer"
            id="stop"
            class="button is-dark is-large"
            @click="stopTimer"
          >
            <i class="far fa-pause-circle"></i>
          </button>
          <!--     Restart Timer -->
          <button
            v-if="resetButton"
            id="reset"
            class="button is-dark is-large"
            @click="resetTimer"
          >
            <i class="fas fa-undo"></i>
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  middleware: 'authenticated',
  components: {},
  data: function() {
    return {
      timer: null,
      minutes: 0,
      seconds: 0,
      totalTime: 0,
      resetButton: false,
      title: 'Let the countdown begin!!'
    }
  },
  computed: {
    minutesText: function() {
      return this.padTime(this.minutes)
    },
    secondsText: function() {
      return this.padTime(this.seconds)
    }
  },
  watch: {
    minutes: function(val) {
      this.totalTime = val * 60 + this.seconds
    },
    seconds: function(val) {
      this.totalTime = this.minutes * 60 + val
    }
  },
  methods: {
    startTimer: async function() {
      try {
        await this.$hub.start(this.$store.state.token)
        this.timer = setInterval(() => this.countdown(), 1000)
        this.resetButton = true
        this.title = 'FIGHT!!!'
      } catch (err) {}
    },
    stopTimer: async function() {
      try {
        await this.$hub.stop(this.$store.state.token)
        clearInterval(this.timer)

        this.timer = null
        this.resetButton = true
        this.title = 'Never quit, keep going!!'
      } catch (err) {}
    },
    resetTimer: async function() {
      try {
        await this.$hub.stop(this.$store.state.token)
        clearInterval(this.timer)
        this.minutes = 0
        this.seconds = 0

        this.timer = null
        this.resetButton = false
        this.title = 'Let the countdown begin!!'
      } catch (err) {}
    },
    padTime: function(time) {
      return (time < 10 ? '0' : '') + time
    },
    countdown: function() {
      if (this.totalTime >= 1) {
        this.totalTime = this.totalTime - 1
      } else {
        this.totalTime = 0
        this.resetTimer()
      }
      this.minutes = Math.floor(this.totalTime / 60)
      this.seconds = this.totalTime - this.minutes * 60
    }
  }
}
</script>

<style lang="scss">
#message {
  color: #f56725;
  font-size: 50px;
  margin-bottom: 20px;
}

.timer {
  color: #f56725;
  font-size: 200px;
  line-height: 1;
  margin-bottom: 40px;
}

.middle {
  flex-basis: 10%;
  color: #f56725;
  max-width: 300px;
  font-size: 200px;
  line-height: 1;
  margin-bottom: 40px;
}

.time {
  text-align: right;
  border: none;
  background: inherit;
  color: #f56725;
  max-width: 250px;
  font-size: 200px;
  line-height: 1;
  margin-bottom: 40px;
}
</style>
