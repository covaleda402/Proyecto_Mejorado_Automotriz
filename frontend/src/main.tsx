/* Composition entrypoint: it only mounts the application on the real root
   element of index.html. It reads no storage and calls no API. */
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './app/App';
import { setupBrowserMockApi } from './services/mock_api';
import './styles/tokens.css';
import './styles/layout.css';
import './styles/component.css';

setupBrowserMockApi();

const container = document.getElementById('root');
if (container) {
  createRoot(container).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}
