import './dashboard.css'
import '@fortawesome/fontawesome-free/js/all.js';

import Dashboard from './Dashboard.svelte'

const app = new Dashboard({ target: document.getElementById('app') });

export default app;
