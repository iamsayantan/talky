require('dotenv').config();
const colors = require('vuetify/es5/util/colors').default

module.exports = {
  mode: 'universal',
  /*
  ** Headers of the page
  */
  head: {
    titleTemplate: 'Talky - Realtime voice and video chat',
    title: process.env.npm_package_name || '',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: process.env.npm_package_description || '' }
    ],
    link: [
      { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' },
      {
        rel: 'stylesheet',
        href:'https://fonts.googleapis.com/css?family=Raleway:400|Roboto+Slab:200&display=swap'
      }
    ],
  },
  server: {
    port: 8050,
    host: '0.0.0.0'
  },

  router: {
    middleware: ['auth']
  },

  /*
  ** Customize the progress-bar color
  */
  loading: { color: '#fff' },
  /*
  ** Global CSS
  */
  css: [
  ],
  /*
  ** Plugins to load before mounting the App
  */
  plugins: [
    {
      src: '~/plugins/webrtc-adapter',
      mode: 'client'
    },
    {
      src: '~/plugins/signalling',
      mode: 'client'
    },
    {src: '~/plugins/axios'}
  ],
  /*
  ** Nuxt.js dev-modules
  */
  buildModules: [
    '@nuxt/typescript-build',
    ['@nuxtjs/vuetify', {
      theme: {
        dark: false
      }
    }],
  ],
  /*
  ** Nuxt.js modules
  */
  modules: [
    '@nuxtjs/axios',
    '@nuxtjs/auth',
    '@nuxtjs/pwa',
    '@nuxtjs/dotenv',
    '@nuxtjs/toast',
  ],
  /*
  ** Axios module configuration
  ** See https://axios.nuxtjs.org/options
  */
  axios: {
  },

  auth: {
    strategies: {
      local: {
        endpoints: {
          login: { url: '/user/v1/login', method: 'post', propertyName: 'access_token' },
          logout: { url: '/user/v1/me', method: 'get' },
          user: { url: '/user/v1/me', method: 'get', propertyName: 'user' }
        },
        tokenType: '',
        watchLoggedIn: true
      },
      redirect: {
        login: '/login',
        logout: '/login',
        home: '/'
      }
    }
  },

  toast: {
    position: 'bottom-center',
    duration: 5000
  },
  /*
  ** vuetify module configuration
  ** https://github.com/nuxt-community/vuetify-module
  */
  vuetify: {
    customVariables: ['~/assets/variables.scss'],
    theme: {
      dark: false,
    }
  },
  /*
  ** Build configuration
  */
  build: {
    vendor: ['webrtc-adapter'],
    /*
    ** You can extend webpack config here
    */
    extend (config, ctx) {
    }
  }
}
