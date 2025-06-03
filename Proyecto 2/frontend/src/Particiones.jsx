import React, { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';

function Particiones() {
  const { encodedPath } = useParams();
  const [particiones, setParticiones] = useState([]);
  const [error, setError] = useState('');
  const [cargando, setCargando] = useState(true);
  const navigate = useNavigate();
  const decodedPath = atob(decodeURIComponent(encodedPath));

  useEffect(() => {
    const obtenerDatos = async () => {
      try {
        const respuesta = await fetch("http://localhost:3001/disk-info", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ filePath: decodedPath })
        });

        const data = await respuesta.json();

        if (respuesta.ok && Array.isArray(data)) {
          setParticiones(data);
        } else {
          setError(data.error || "No se encontraron particiones.");
        }
      } catch (err) {
        console.error(err);
        setError("Error al conectar con el backend.");
      } finally {
        setCargando(false);
      }
    };

    obtenerDatos();
  }, [decodedPath]);

  return (
    <main className="min-h-screen bg-[#0a192f] text-white px-6 py-10 font-mono">
      <div className="max-w-5xl mx-auto space-y-8">
        <h1 className="text-3xl text-teal-400 font-bold">Particiones del Disco</h1>
        <p className="text-sm text-gray-400 break-all">Ruta: {decodedPath}</p>

        <div className="flex gap-4 mt-4">
          <Link
            to="/discos"
            className="px-4 py-2 border border-yellow-400 text-yellow-400 rounded hover:bg-yellow-400 hover:text-black transition"
          >
            ⬅ Volver a Discos
          </Link>
        </div>

        {cargando && <p className="text-gray-400">Cargando particiones...</p>}

        {error && <div className="bg-red-500 text-black p-4 rounded">{error}</div>}

        {!cargando && particiones.length === 0 && !error && (
          <p className="text-gray-400">Este disco no contiene particiones activas.</p>
        )}

        {particiones.length > 0 && (
          <div className="space-y-4 mt-6">
            {particiones.map((p, i) => (
              <div key={i} className="bg-[#112240] border border-gray-600 rounded p-4">
                <h3 className="text-lg text-yellow-300 font-semibold">{p.name}</h3>
                <p>Tamaño: {p.size} bytes</p>
                <p>Fit: {p.fit || 'N/A'}</p>
                <p>Estado: {p.status}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </main>
  );
}

export default Particiones;
