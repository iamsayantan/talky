export default function ({ $axios, redirect, app }) {
  $axios.onError(error => {
    // only show the toast for browser
    if(process.client &&
      error.response &&
      error.response.data
    ) {
      let data = error.response.data;

      console.log(data)
      if(data.error) {
        app.$toast.show(data.error);
      }

      // if we get 401 then locally logout the user
      if(error.response.status === 401) {
        app.$auth.reset();
      }
    }
  });
}
