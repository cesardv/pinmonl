export default () => {
  return {
    mounted () {
      document.addEventListener('keydown', this.handleKeyPress)
    },
    beforeDestroy () {
      document.removeEventListener('keydown', this.handleKeyPress)
    },
    computed: {
      shouldDisableKeys () {
        return this.$store.getters.globalSearch
      },
    },
    methods: {
      handleKeyPress () {
        // extends method
      },
    },
  }
}