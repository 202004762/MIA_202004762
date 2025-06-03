import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import App from './App.jsx';
import Discos from './Discos.jsx';
import Particiones from './Particiones.jsx';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/discos" element={<Discos />} />
        <Route path="/particiones/:encodedPath" element={<Particiones />} /> {}
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
