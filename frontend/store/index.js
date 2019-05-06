import Vuex from 'vuex'

const cookieparser = process.server ? require('cookieparser') : undefined

const createStore = () => {
  return new Vuex.Store({
    state: () => ({
      token: null
    }),
    mutations: {
      setToken(state, token) {
        state.token = token
      }
    },
    actions: {
      nuxtServerInit({ commit }, { req }) {
        let token = null
        if (req.headers.cookie) {
          const parsed = cookieparser.parse(req.headers.cookie)
          try {
            token = parsed.RoboconToken
          } catch (err) {
            // No valid cookie found
          }
        }
        commit('setToken', token)
      }
    }
  })
}

export default createStore
