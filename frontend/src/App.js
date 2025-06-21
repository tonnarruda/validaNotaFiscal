import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import NfValidator from './components/NfValidator';
import EnviaNotaFiscal from './components/EnviaNotaFiscal';
import Navigation from './components/Navigation';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <Navigation />
        <div className="py-8">
          <div className="container mx-auto px-4">
            <Routes>
              <Route path="/processa-nota-fiscal" element={<NfValidator />} />
              <Route path="/envia-nota-fiscal" element={<EnviaNotaFiscal />} />
              <Route path="/" element={<Navigate to="/processa-nota-fiscal" replace />} />
            </Routes>
          </div>
        </div>
      </div>
    </Router>
  );
}

export default App; 