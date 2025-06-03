import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function App() {
  const [entradaComandos, setEntradaComandos] = useState('');
  const [salidaComandos, setSalidaComandos] = useState('');
  const [mostrarLogin, setMostrarLogin] = useState(false);
  const [user, setUser] = useState('');
  const [pass, setPass] = useState('');
  const [id, setId] = useState('');
  const [sesionIniciada, setSesionIniciada] = useState(false);

  const navigate = useNavigate();

  useEffect(() => {
    fetch("http://localhost:3001/session-status")
      .then(res => res.json())
      .then(data => {
        setSesionIniciada(data.authenticated);
      })
      .catch(err => {
        console.error("Error al verificar sesión:", err);
        setSesionIniciada(false);
      });
  }, []);

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

  const manejarLogin = async () => {
    const comando = `login -user=${user} -pass=${pass} -id=${id}`;
    const respuesta = await fetch("http://localhost:3001/execute", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ command: comando })
    });
    const datos = await respuesta.json();
    setSalidaComandos(datos.output);
    if (respuesta.ok && !datos.output.includes("Error")) {
      setSesionIniciada(true);
      setMostrarLogin(false);
    }
  };

  const manejarLogout = async () => {
    const respuesta = await fetch("http://localhost:3001/execute", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ command: "logout" })
    });
    const datos = await respuesta.json();
    setSalidaComandos(datos.output);
    if (respuesta.ok && !datos.output.includes("Error")) {
      setSesionIniciada(false);
    }
  };

  return (
    <main className="min-h-screen bg-[#0a192f] text-white px-6 py-10 font-mono">
      <div className="max-w-4xl mx-auto space-y-10">
        <h1 className="text-5xl font-bold text-gray-100">Simulador EXT2</h1>
        <h2 className="text-4xl font-bold text-gray-400">Proyecto MIA 1S2025</h2>

        <div className="flex flex-wrap gap-4">
          {!sesionIniciada ? (
            <button
              onClick={() => setMostrarLogin(true)}
              className="px-6 py-2 border border-yellow-400 text-yellow-400 rounded hover:bg-yellow-400 hover:text-[#0a192f] transition"
            >
              Iniciar Sesión
            </button>
          ) : (
            <>
              <button
                onClick={manejarLogout}
                className="px-6 py-2 border border-red-400 text-red-400 rounded hover:bg-red-400 hover:text-[#0a192f] transition"
              >
                Cerrar Sesión
              </button>
              <button
                onClick={() => navigate("/discos")}
                className="px-6 py-2 border border-blue-400 text-blue-400 rounded hover:bg-blue-400 hover:text-[#0a192f] transition"
              >
                Visualizar Discos
              </button>
            </>
          )}
          <button
            onClick={() => setSalidaComandos('')}
            className="px-6 py-2 border border-red-500 text-red-500 rounded hover:bg-red-500 hover:text-[#0a192f] transition"
          >
            Limpiar
          </button>
        </div>

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
        </div>

        <div className="bg-[#112240] border border-gray-600 rounded p-4 h-80 overflow-auto text-green-400">
          <pre>{salidaComandos}</pre>
        </div>
      </div>

      {mostrarLogin && (
        <div className="fixed inset-0 bg-black bg-opacity-60 flex justify-center items-center">
          <div className="bg-[#112240] p-8 rounded-lg border border-teal-500 space-y-4 w-96">
            <h3 className="text-xl text-teal-400">Iniciar Sesión</h3>
            <input
              type="text"
              placeholder="Usuario"
              value={user}
              onChange={(e) => setUser(e.target.value)}
              className="w-full p-2 rounded bg-[#0a192f] text-white border border-gray-500"
            />
            <input
              type="password"
              placeholder="Contraseña"
              value={pass}
              onChange={(e) => setPass(e.target.value)}
              className="w-full p-2 rounded bg-[#0a192f] text-white border border-gray-500"
            />
            <input
              type="text"
              placeholder="ID de partición"
              value={id}
              onChange={(e) => setId(e.target.value)}
              className="w-full p-2 rounded bg-[#0a192f] text-white border border-gray-500"
            />
            <div className="flex justify-end space-x-2">
              <button
                onClick={() => setMostrarLogin(false)}
                className="px-4 py-1 text-red-400 border border-red-400 rounded hover:bg-red-400 hover:text-black"
              >
                Cancelar
              </button>
              <button
                onClick={manejarLogin}
                className="px-4 py-1 text-teal-400 border border-teal-400 rounded hover:bg-teal-400 hover:text-black"
              >
                Ingresar
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

export default App;
