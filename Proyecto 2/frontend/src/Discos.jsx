import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

function Discos() {
  const [ruta, setRuta] = useState('');
  const [discos, setDiscos] = useState([]);
  const [error, setError] = useState('');
  const [cargando, setCargando] = useState(false);
  const navigate = useNavigate();

  const obtenerDiscos = async () => {
    setError('');
    setCargando(true);
    setDiscos([]);

    try {
      const respuesta = await fetch("http://localhost:3001/disks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ basePath: ruta })
      });

      const data = await respuesta.json();

      if (respuesta.ok && Array.isArray(data)) {
        setDiscos(data);
      } else {
        setError(data.error || "No se pudieron obtener los discos.");
      }
    } catch (err) {
      console.error(err);
      setError("Error al conectar con el backend.");
    }

    setCargando(false);
  };

  const irADetalle = (path) => {
    const encodedPath = encodeURIComponent(btoa(path));
    navigate(`/particiones/${encodedPath}`);
  };

  return (
    <main className="min-h-screen bg-[#0a192f] text-white px-6 py-10 font-mono">
      <div className="max-w-5xl mx-auto space-y-10">
        <h1 className="text-4xl font-bold text-teal-400">Visualización de Discos</h1>

        <button
          onClick={() => navigate('/')}
          className="px-4 py-2 border border-yellow-400 text-yellow-400 rounded hover:bg-yellow-400 hover:text-black transition"
        >
          ⬅ Regresar
        </button>

        <div className="flex items-center space-x-4 mt-4">
          <input
            type="text"
            value={ruta}
            onChange={(e) => setRuta(e.target.value)}
            placeholder="Ingresa la ruta base donde están los discos (.mia)"
            className="w-full px-4 py-2 rounded bg-[#112240] text-white border border-gray-500"
          />
          <button
            onClick={obtenerDiscos}
            className="px-6 py-2 border border-blue-500 text-blue-500 rounded hover:bg-blue-500 hover:text-[#0a192f] transition"
          >
            Buscar
          </button>
        </div>

        {cargando && <p className="text-gray-400">Buscando discos...</p>}

        {error && (
          <div className="bg-red-500 text-black p-4 rounded">
            {error}
          </div>
        )}

        {discos.length > 0 && discos.map((disco, idx) => (
          <div key={idx} className="bg-[#112240] border border-gray-600 p-6 rounded shadow space-y-4">
            <h2 className="text-2xl text-yellow-300">{disco.path}</h2>
            <p>Tamaño: {disco.size} bytes</p>
            <p>Fit: {disco.fit || 'N/A'}</p>
            <p>Creado: {disco.created_at}</p>
            <button
              onClick={() => irADetalle(disco.path)}
              className="mt-2 px-4 py-1 border border-cyan-400 text-cyan-400 rounded hover:bg-cyan-400 hover:text-black transition"
            >
              Ver Particiones
            </button>
          </div>
        ))}
      </div>
    </main>
  );
}

export default Discos;
