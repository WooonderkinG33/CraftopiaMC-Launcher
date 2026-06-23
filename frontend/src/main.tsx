import {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';
import App from './App';

// Hide splash before React renders
const splash = document.getElementById('splash');
if (splash) splash.classList.add('hide');

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
