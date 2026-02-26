import './main.css';

import App from './App.svelte';

const skeleton = document.getElementById('app');
if (!skeleton) {
  throw new Error("Unable to find the skeleton (element with id 'app').");
}

const app = new App({ target: skeleton });

export default app;
