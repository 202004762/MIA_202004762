import React, { useState } from 'react';
import { Button } from "@/components/ui/button";

function App() {
  const [entradaComandos, setEntradaComandos] = useState('');
  const [salidaComandos, setSalidaComandos] = useState('');

  const manejarEjecucion = async () => {
    if (!entradaComandos.trim()) return;
    const respuesta = await fetch("http://localhost:3001/execute", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ command: entradaComandos })
    });
    const datos = await respuesta.json();
    setSalidaComandos(datos.output);
  };

  const manejarArchivo = async (evento) => {
    const archivo = evento.target.files[0];
    if (!archivo) return;
    const contenido = await archivo.text();
    setEntradaComandos(contenido);
  };

  return (
    <main className="min-h-screen bg-[#0a192f] text-white px-6 py-10 font-mono">
      <div className="max-w-4xl mx-auto space-y-10">
        <h1 className="text-5xl font-bold text-gray-100">Simulador EXT2</h1>
        <h2 className="text-4xl font-bold text-gray-400">Proyecto MIA 1S2025</h2>

        <textarea
          className="w-full h-40 p-4 rounded bg-[#112240] text-white border border-gray-600 placeholder-gray-500"
          value={entradaComandos}
          onChange={(e) => setEntradaComandos(e.target.value)}
          placeholder="Escribe comandos aquí o carga un archivo .smia"
        />

        <div className="flex items-center space-x-4">
          <input
            type="file"
            accept=".smia"
            onChange={manejarArchivo}
            className="text-sm text-gray-300 file:mr-4 file:py-2 file:px-4 file:rounded-full
                        file:border-0 file:text-sm file:font-semibold
                        file:bg-teal-500 file:text-white hover:file:bg-teal-400"
          />
          <button
            onClick={manejarEjecucion}
            className="px-6 py-2 border border-teal-500 text-teal-500 rounded hover:bg-teal-500 hover:text-[#0a192f] transition"
          >
            Ejecutar
          </button>
          <button
            onClick={() => setSalidaComandos('')}
            className="px-6 py-2 border border-red-500 text-red-500 rounded hover:bg-red-500 hover:text-[#0a192f] transition"
          >
            Limpiar
          </button>
        </div>

        <div className="bg-[#112240] border border-gray-600 rounded p-4 h-80 overflow-auto text-green-400">
          <pre>{salidaComandos}</pre>
        </div>
      </div>
    </main>
  );
}

export default App;