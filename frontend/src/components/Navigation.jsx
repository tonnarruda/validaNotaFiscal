import React from 'react';
import { Link, useLocation } from 'react-router-dom';

const Navigation = () => {
  const location = useLocation();

  return (
    <nav className="bg-white shadow-lg mb-8">
      <div className="max-w-6xl mx-auto px-4">
        <div className="flex justify-between items-center py-4">
          <div className="flex items-center space-x-8">
            <Link
              to="/processa-nota-fiscal"
              className={`px-4 py-2 rounded-lg font-semibold transition-colors ${
                location.pathname === '/processa-nota-fiscal'
                  ? 'bg-blue-500 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              Processar Nota Fiscal
            </Link>
            <Link
              to="/envia-nota-fiscal"
              className={`px-4 py-2 rounded-lg font-semibold transition-colors ${
                location.pathname === '/envia-nota-fiscal'
                  ? 'bg-blue-500 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              Enviar Nota Fiscal
            </Link>
          </div>
          <div className="text-xl font-bold text-gray-800">
            Sistema de Notas Fiscais
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navigation; 