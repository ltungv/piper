import hub from '~/api/hub'

export default (ctx, inject) => {
  inject('hub', hub(ctx.$axios))
}
