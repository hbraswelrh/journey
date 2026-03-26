// SPDX-License-Identifier: Apache-2.0

import { StrictMode, Suspense, lazy } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import './index.css';
import App from './App';

const Playground = lazy(
  () => import('./components/playground/Playground'),
);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route
          path="/playground"
          element={
            <Suspense
              fallback={
                <div className="app" style={{ textAlign: 'center', padding: '64px 0' }}>
                  Loading Playground...
                </div>
              }
            >
              <Playground />
            </Suspense>
          }
        />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);
