# Scene Scheduler — Manual de Usuario

**Versión:** 0.5  
**Fecha:** Febrero 2026  
**Aplicación:** Scene Scheduler para OBS Studio

---

## Tabla de Contenidos

1. [Primeros Pasos](#1-primeros-pasos)
2. [Configuración](#2-configuración)
3. [Interfaz Web](#3-interfaz-web)
4. [Gestión de Eventos](#4-gestión-de-eventos)
5. [Tipos de Fuente](#5-tipos-de-fuente)
6. [Vista Previa en Directo](#6-vista-previa-en-directo)
7. [Funcionamiento Interno](#7-funcionamiento-interno)
8. [Referencia del JSON de Programación](#8-referencia-del-json-de-programación)
9. [Resolución de Problemas](#9-resolución-de-problemas)

---

## 1. Primeros Pasos

### 1.1 ¿Qué es Scene Scheduler?

Scene Scheduler es una herramienta de automatización externa para OBS Studio que ejecuta tu emisión según una programación temporal — como la parrilla de una cadena de televisión. Tú defines qué fuentes se emiten y a qué hora, y Scene Scheduler gestiona las transiciones automáticamente en una escena principal, 24/7, sin intervención manual.

**Características principales:**
- Automatización basada en horarios con eventos recurrentes (diarios/semanales)
- Calendario visual (FullCalendar) accesible desde cualquier navegador en tu red
- Vista previa de fuentes en tiempo real vía WebRTC y HLS
- Preparación automática de fuentes OBS para transiciones sin cortes
- Fuente de respaldo opcional para periodos sin programación

### 1.2 Requisitos Previos

1. **OBS Studio** (versión 28.0+) con el **plugin WebSocket v5** (incluido por defecto en OBS 28+)
2. **Sistema Operativo**: Linux (probado en Ubuntu 20.04+), Windows 10/11 o macOS
3. **Red**: OBS y Scene Scheduler deben ser accesibles por red

### 1.3 Instalación

**Linux:**
```bash
tar -xzf scenescheduler-linux-amd64.tar.gz
cd scenescheduler
chmod +x build/scenescheduler
```

**Windows:**
1. Extraer `scenescheduler-windows-amd64.zip` en una carpeta (ej: `C:\scenescheduler\`)
2. Abrir el Símbolo del sistema en esa carpeta

### 1.4 Inicio Rápido

**Paso 1 — Configurar OBS WebSocket:**
1. Abrir OBS Studio → **Herramientas** → **Configuración del servidor WebSocket**
2. Activar "Habilitar servidor WebSocket"
3. Establecer una contraseña (recomendado) y anotar el puerto (por defecto: 4455)

**Paso 2 — Editar `config.json`:**
```json
{
  "obs": {
    "host": "localhost",
    "port": 4455,
    "password": "tu-contraseña-obs",
    "scheduleScene": "_SCHEDULER",
    "scheduleSceneAux": "_SCHEDULER_AUX"
  },
  "webServer": {
    "port": "8080",
    "user": "admin",
    "password": "tu-contraseña-web",
    "hlsPath": "hls"
  },
  "paths": {
    "logFile": "logs.txt",
    "schedule": "schedule.json"
  }
}
```

**Paso 3 — Iniciar Scene Scheduler:**
```bash
# Linux
./build/scenescheduler

# Windows
scenescheduler.exe
```

**Paso 4 — Abrir la interfaz web:**
- Misma máquina: `http://localhost:8080`
- Otros dispositivos: `http://<ip-del-servidor>:8080`

Si configuraste `user` y `password`, el navegador pedirá las credenciales de autenticación HTTP Basic.

---

## 2. Configuración

Scene Scheduler utiliza un único archivo `config.json` ubicado en el mismo directorio que el ejecutable.

### 2.1 Conexión OBS (`obs`)

| Campo | Tipo | Predeterminado | Descripción |
|-------|------|----------------|-------------|
| `host` | string | `"localhost"` | Hostname o IP de OBS |
| `port` | integer | `4455` | Puerto WebSocket de OBS |
| `password` | string | `""` | Contraseña WebSocket de OBS |
| `reconnectInterval` | integer | `15` | Segundos entre intentos de reconexión |
| **`scheduleScene`** | string | — | **Obligatorio.** Escena principal gestionada por el planificador |
| **`scheduleSceneAux`** | string | — | **Obligatorio.** Escena auxiliar de "preparación" para precargar fuentes |
| `sourceNamePrefix` | string | `"_sched_"` | Prefijo para las fuentes creadas por el planificador |

Ambas escenas (`scheduleScene` y `scheduleSceneAux`) se crean automáticamente en OBS si no existen.

### 2.2 Servidor Web (`webServer`)

| Campo | Tipo | Predeterminado | Descripción |
|-------|------|----------------|-------------|
| `port` | string | `"8080"` | Puerto del servidor HTTP |
| `user` | string | `""` | Usuario HTTP Basic Auth (vacío = autenticación desactivada) |
| `password` | string | `""` | Contraseña HTTP Basic Auth |
| `hlsPath` | string | `"hls"` | Directorio para archivos de vista previa HLS (debe ser relativo) |
| `enableTls` | boolean | `false` | Activar HTTPS |
| `certFilePath` | string | `""` | Ruta al certificado TLS (obligatorio si TLS activo) |
| `keyFilePath` | string | `""` | Ruta a la clave privada TLS (obligatorio si TLS activo) |

### 2.3 Fuente de Medios (`mediaSource`)

| Campo | Tipo | Predeterminado | Descripción |
|-------|------|----------------|-------------|
| `videoDeviceIdentifier` | string | `""` | Identificador del dispositivo de captura de vídeo |
| `audioDeviceIdentifier` | string | `""` | Identificador del dispositivo de captura de audio |
| `quality` | string | `"low"` | Calidad: `"low"`, `"medium"`, `"high"` |

Estos ajustes controlan la vista previa WebRTC en directo mostrada en la vista Monitor.

**Cómo encontrar los identificadores de dispositivo:**

Ejecuta el siguiente comando para listar todos los dispositivos de medios disponibles en tu sistema:
```bash
./build/scenescheduler --list-devices
```

Ejemplo de salida:
```
----------- Available Media Devices -----------
INFO: Use the 'Friendly Name' or 'DeviceID' for your config.

VIDEO DEVICES:
  #1:
    Friendly Name : HD Webcam C920
    DeviceID      : video0

AUDIO DEVICES:
  #1:
    Friendly Name : Built-in Audio Analog Stereo
    DeviceID      : default.monitor

----------------------------------------------
```

Usa los valores de `DeviceID` en tu `config.json`:
```json
"mediaSource": {
  "videoDeviceIdentifier": "video0",
  "audioDeviceIdentifier": "default.monitor",
  "quality": "medium"
}
```

### 2.4 Rutas (`paths`)

| Campo | Tipo | Predeterminado | Descripción |
|-------|------|----------------|-------------|
| `logFile` | string | `""` | Ruta del archivo de log |
| `schedule` | string | `"schedule.json"` | Ruta del archivo de programación |

### 2.5 Planificador (`scheduler`)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `defaultSource` | objeto | Fuente de respaldo opcional para periodos inactivos |
| `defaultSource.name` | string | Nombre de la fuente OBS |
| `defaultSource.inputKind` | string | Tipo de entrada OBS (ej: `"image_source"`) |
| `defaultSource.uri` | string | Ruta o URL del contenido |
| `defaultSource.inputSettings` | objeto | Ajustes adicionales de entrada OBS |
| `defaultSource.transform` | objeto | Transformación de posición/escala/recorte |

### 2.6 Validación

Al iniciar, Scene Scheduler valida:
- ✅ `obs.scheduleScene` y `obs.scheduleSceneAux` están presentes (fatal si faltan)
- ✅ `webServer.hlsPath` es una ruta relativa segura (sin `..` ni rutas absolutas)
- ✅ Rutas de certificado TLS presentes cuando `enableTls` es true
- ⚠️ Advertencia si `obs.password` está vacío
- ⚠️ Advertencia si `webServer.user` o `webServer.password` están vacíos

---

## 3. Interfaz Web

Scene Scheduler es una **aplicación de página única** servida en la URL raíz (`http://<host>:<puerto>/`). No existe una página `/editor.html` separada.

### 3.1 Cambiar de Vista

La cabecera contiene un **desplegable de Vista** con dos opciones:
- **📺 Monitor** — Calendario de solo lectura con vista previa en directo y registro de actividad
- **📝 Editor** — Calendario editable para gestionar eventos

### 3.2 Indicadores de Estado de Conexión

La cabecera muestra **tres** indicadores independientes:

| Indicador | Verde | Rojo |
|-----------|-------|------|
| **Server** | WebSocket conectado al backend | Desconectado (reconexión automática cada 5s) |
| **OBS** | Backend conectado a OBS Studio | OBS no accesible |
| **Preview** | VirtualCam activa, vista previa disponible | Sin flujo de vista previa |

### 3.3 Vista Monitor

La vista Monitor está diseñada para observación pasiva. Contiene:

- **Barra lateral izquierda:**
  - **Live Preview** — Emisión de vídeo WebRTC desde la cámara/micrófono del servidor (requiere configuración de `mediaSource`)
  - **Activity Log** — Registro en tiempo real de eventos del servidor (conexiones, cargas de programación, errores)

- **Área principal:**
  - **Calendario** (solo lectura) — Muestra todos los eventos programados como bloques de color en una línea temporal
  - **Resaltado del evento actual** — El evento activo se resalta en verde (`#22c55e`)
  - Al hacer clic en un evento se abre un **popup de vista previa** (solo lectura) mostrando la URI de la fuente, tipo de entrada y un botón Preview

### 3.4 Vista Editor

La vista Editor proporciona control total sobre la programación:

- **Calendario** (editable) — Haz clic en un hueco de tiempo para crear un evento nuevo, o en uno existente para editarlo
- **Arrastrar y redimensionar** — Los eventos se pueden mover o redimensionar directamente en el calendario
- **Indicador de sincronización** — La cabecera muestra "Synced" (verde) o "Unsaved" (naranja) para indicar si la programación local coincide con el servidor

---

## 4. Gestión de Eventos

### 4.1 Crear un Evento

1. En la **vista Editor**, haz clic en un hueco de tiempo vacío del calendario
2. Se abre el modal de **Task Editor** con cinco pestañas
3. Rellena los campos obligatorios (como mínimo: título, hora de inicio/fin)
4. Pulsa **Save Changes**

### 4.2 Modal de Evento — Cinco Pestañas

#### Pestaña 1: General

| Campo | Descripción |
|-------|-------------|
| **Description** | Descripción de texto opcional del evento |
| **Tags** | Etiquetas separadas por espacios para organización |
| **ClassNames** | Clases CSS para estilos personalizados |
| **Text Color** | Selector de color para el texto del evento |
| **Background Color** | Selector de color para el fondo del evento |
| **Border Color** | Selector de color para el borde del evento (eventos recurrentes) |

#### Pestaña 2: Source

Define la fuente OBS que se creará cuando se active este evento.

| Campo | Descripción |
|-------|-------------|
| **Input Name** * | Nombre técnico de la fuente OBS (ej: `"YT_Chillhop"`) |
| **Input Kind** * | Desplegable del tipo de fuente (ver [Sección 5](#5-tipos-de-fuente)) |
| **URI** * | Ruta o URL del contenido |
| **Settings (JSON)** | Ajustes adicionales de entrada OBS en JSON |
| **Transform (JSON)** | Posición, escala y recorte en JSON |

#### Pestaña 3: Timing

| Campo | Descripción |
|-------|-------------|
| **Start** * | Fecha y hora de inicio (selector de fecha/hora con segundos) |
| **End** * | Fecha y hora de fin |
| **Duration** | Calculada automáticamente (solo lectura) |
| **Recurring** | Interruptor para activar recurrencia semanal |

Cuando **Recurring** está activado:
- **From / Until** — Rango de fechas para la serie recurrente
- **Week Days** — Casillas para Lun–Dom
- Solo se usa la parte de **hora** de Start/End; las fechas provienen del rango de recurrencia

#### Pestaña 4: Behavior

| Campo | Descripción |
|-------|-------------|
| **Preload seconds** | Segundos de antelación para empezar a preparar la fuente (predeterminado: 0) |
| **On end action** | Qué ocurre al terminar el evento: `hide` (predeterminado), `none`, o `stop` |

#### Pestaña 5: Preview

- Botón **Preview Source** — Genera un flujo HLS de vista previa de la fuente configurada
- El reproductor de vídeo muestra la vista previa dentro del modal
- Requiere la herramienta complementaria `hls-generator`

### 4.3 Campo Título

El campo **Title** aparece encima de las pestañas, siempre visible. Sirve como nombre del evento en el calendario y como identificador rápido.

### 4.4 Interruptor Activado/Desactivado

Junto al título, un **interruptor** controla si el evento está activo. Los eventos desactivados permanecen en la programación pero no son ejecutados por el planificador.

### 4.5 Editar Eventos

Haz clic en cualquier evento del calendario del Editor para reabrir el modal con todos los campos rellenos.

### 4.6 Eliminar Eventos

El botón **Delete** aparece en la esquina inferior izquierda del modal. Eliminar un evento lo borra inmediatamente.

### 4.7 Arrastrar y Redimensionar

En el calendario del Editor:
- **Arrastra** un evento para moverlo a otro horario
- **Redimensiona** arrastrando el borde inferior para cambiar la duración

### 4.8 Guardar en el Servidor

Aplicar cambios al archivo `schedule.json` en el servidor es un proceso de dos pasos para evitar interrupciones accidentales en la emisión:

1. **Guardar borrador local**: Cuando haces clic en "Save Changes" en un evento, o al arrastrar/redimensionar en el calendario, los cambios solo se guardan en tu navegador. Aparecerá un indicador naranja arriba a la derecha diciendo "X unsaved changes".
2. **Publicar en el servidor**: Para aplicar tu borrador al servidor, haz clic en el botón de menú `...` arriba a la derecha y selecciona **"Commit to Server"**. Solo entonces la programación se enviará por WebSocket (`commitSchedule`), se guardará en disco, y el planificador comenzará a ejecutarla.

---

## 5. Tipos de Fuente

El desplegable **Input Kind** en la pestaña Source ofrece estos tipos de fuente OBS:

| Tipo de Entrada | Uso | Ejemplo de URI |
|-----------------|-----|----------------|
| `ffmpeg_source` | Archivos de vídeo/audio locales, flujos RTMP/RTSP/RTP/SRT | `/ruta/al/video.mp4` o `rtmp://servidor/flujo` |
| `browser_source` | Páginas web, overlays HTML, reproductores de vídeo embebidos | `https://www.youtube.com/embed/...` |
| `vlc_source` | Listas de reproducción VLC | `/ruta/a/playlist.m3u` |
| `ndi_source` | Flujos de vídeo NDI en red | Nombre de fuente NDI |
| `image_source` | Imágenes estáticas | `/ruta/a/imagen.png` |

### 5.1 Ajustes de Entrada (JSON)

El campo **Settings** acepta JSON que se pasa directamente a OBS al crear la fuente. Ejemplo:

**Fuente browser con dimensiones personalizadas:**
```json
{
  "css": "body { background-color: rgba(0, 0, 0, 0); margin: 0px auto; overflow: hidden; }",
  "height": 1080,
  "width": 1920
}
```

### 5.2 Transformación (JSON)

El campo **Transform** acepta JSON para posicionar la fuente en la escena:

```json
{
  "PositionX": 100,
  "PositionY": 50,
  "ScaleX": 0.5,
  "ScaleY": 0.5
}
```

---

## 6. Vista Previa en Directo

### 6.1 Vista Monitor — Vista Previa en Directo

La barra lateral izquierda de la vista Monitor muestra una **vista previa WebRTC en directo** de la cámara y micrófono del servidor. Utiliza el protocolo WHEP (endpoint `/whep/`) y requiere:
- `mediaSource.videoDeviceIdentifier` y `audioDeviceIdentifier` configurados en `config.json`
- El indicador Preview en verde (VirtualCam activa)

### 6.2 Vista Previa de Fuente (en el Modal de Evento)

La pestaña **Preview** del modal de evento permite probar una fuente antes de guardarla:
1. Configura la fuente en la pestaña **Source** (Input Kind + URI)
2. Cambia a la pestaña **Preview**
3. Pulsa **▶ Preview Source**
4. El servidor genera un flujo HLS usando la herramienta complementaria `hls-generator`
5. El vídeo se reproduce dentro del modal

El binario `hls-generator` debe estar en el mismo directorio que el ejecutable `scenescheduler`.

### 6.3 Popup de Vista Previa en Monitor

En la vista Monitor, al hacer clic en un evento del calendario se abre un **popup de vista previa** mostrando:
- URI de la fuente y tipo de entrada
- Botón **▶ Preview Source** para generar una vista previa HLS
- Botón **Edit in Editor View** para ir al editor del evento

---

## 7. Funcionamiento Interno

### 7.1 Arquitectura

```
┌─────────────────────────────────────────────────────────┐
│                   Servidor de Producción                 │
│                                                          │
│  ┌──────────────┐    WebSocket     ┌──────────────────┐ │
│  │  OBS Studio  │ ◄──────────────  │  Scene Scheduler │ │
│  │              │   (localhost)     │     (Backend)    │ │
│  └──────────────┘                  └────────┬─────────┘ │
│                                             │            │
│                                    HTTP (0.0.0.0:8080)   │
└─────────────────────────────────────────────┼────────────┘
                                              │
                        Red Local (LAN)       │
                    ┌─────────────────────────┼──────────┐
                    │                         │          │
               ┌────▼─────┐           ┌──────▼──┐  ┌───▼────┐
               │  Portátil │           │ Tablet  │  │ Móvil  │
               │ (Editor)  │           │(Monitor)│  │(Monitor│
               └───────────┘           └─────────┘  └────────┘
```

### 7.2 Comunicación

El backend expone cuatro endpoints HTTP:

| Endpoint | Protocolo | Propósito |
|----------|-----------|-----------|
| `/ws` | WebSocket | Comunicación bidireccional en tiempo real |
| `/whep/` | HTTP (WebRTC) | Vista previa en directo de cámara/micrófono vía WHEP |
| `/hls/` | HTTP (estático) | Archivos de flujo de vista previa HLS |
| `/` | HTTP (estático) | Aplicación frontend |

### 7.3 Protocolo WebSocket

Los mensajes usan el formato `{ "action": "cadena", "payload": {} }`.

**Cliente → Servidor:**

| Acción | Payload | Descripción |
|--------|---------|-------------|
| `getSchedule` | `{}` | Solicitar programación actual |
| `commitSchedule` | JSON Schedule v1.0 | Guardar cambios de programación |
| `getStatus` | `{}` | Solicitar estado de OBS y vista previa |

**Servidor → Cliente:**

| Acción | Payload | Descripción |
|--------|---------|-------------|
| `currentSchedule` | JSON Schedule v1.0 | Datos completos de programación |
| `log` | string | Mensaje de registro de actividad |
| `obsConnected` | `{ obsVersion, timestamp }` | Conexión OBS establecida |
| `obsDisconnected` | `{ timestamp }` | Conexión OBS perdida |
| `virtualCamStarted` | `{}` | Flujo de vista previa disponible |
| `virtualCamStopped` | `{}` | Flujo de vista previa detenido |
| `currentStatus` | `{ obsConnected, obsVersion, virtualCamActive }` | Estado inicial al conectar |
| `previewReady` | `{ hlsUrl }` | Flujo HLS de vista previa listo |
| `previewError` | `{ error }` | Error en vista previa |
| `previewStopped` | `{ reason }` | Vista previa detenida automáticamente |

### 7.4 Proceso de Preparación de Fuentes

Cuando llega la hora de un evento programado:

1. **Preparar** — La fuente se crea en `scheduleSceneAux` (invisible para los espectadores), configurada con todos los ajustes y transformaciones
2. **Activar** — La fuente se mueve de la escena auxiliar a `scheduleScene`
3. **Cambiar escena** — OBS transiciona a `scheduleScene`
4. **Limpiar** — Los elementos temporales se eliminan de `scheduleSceneAux`
5. **Monitorizar** — La fuente permanece activa hasta el fin del evento, momento en que se ejecuta la `onEndAction` configurada (`hide`, `stop` o `none`)

### 7.5 Fuente de Respaldo

Cuando no hay ningún evento programado (periodo inactivo), el `scheduler.defaultSource` (si está configurado) se activa automáticamente, proporcionando una imagen o contenido en espera.

---

## 8. Referencia del JSON de Programación

El archivo de programación (`schedule.json`) sigue el formato **Schedule v1.0**:

```json
{
  "version": "1.0",
  "scheduleName": "Schedule",
  "schedule": [
    {
      "id": "evt-abc123",
      "title": "Noticias Matinales",
      "enabled": true,
      "general": {
        "description": "Emisión matinal diaria",
        "tags": ["noticias", "mañana"],
        "classNames": [],
        "textColor": "#ffffff",
        "backgroundColor": "#1f2fad",
        "borderColor": "#0fc233"
      },
      "source": {
        "name": "FlujomatinalStream",
        "inputKind": "ffmpeg_source",
        "uri": "rtmp://stream.ejemplo.com/live",
        "inputSettings": {},
        "transform": {}
      },
      "timing": {
        "start": "2025-01-01T07:00:00Z",
        "end": "2025-01-01T14:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2025-01-01",
          "endRecur": ""
        }
      },
      "behavior": {
        "onEndAction": "hide",
        "preloadSeconds": 0
      }
    }
  ]
}
```

### 8.1 Referencia de Campos

| Campo | Obligatorio | Descripción |
|-------|-------------|-------------|
| `id` | Sí | Identificador único del evento (auto-generado) |
| `title` | Sí | Nombre para mostrar |
| `enabled` | Sí | Si el evento está activo |
| `general.description` | No | Descripción de texto |
| `general.tags` | No | Array de cadenas de etiquetas |
| `general.classNames` | No | Clases CSS para estilos |
| `general.textColor` | No | Color hexadecimal del texto |
| `general.backgroundColor` | No | Color hexadecimal del fondo |
| `general.borderColor` | No | Color hexadecimal del borde |
| `source.name` | Sí | Nombre de fuente OBS |
| `source.inputKind` | Sí | Tipo de entrada OBS |
| `source.uri` | Sí | Ruta o URL del contenido |
| `source.inputSettings` | No | Ajustes adicionales OBS (objeto JSON) |
| `source.transform` | No | Posición/escala/recorte (objeto JSON) |
| `timing.start` | Sí | Hora de inicio UTC en formato ISO 8601 |
| `timing.end` | Sí | Hora de fin UTC en formato ISO 8601 |
| `timing.isRecurring` | Sí | Si es un evento recurrente |
| `timing.recurrence.daysOfWeek` | Si recurrente | Array de `"MON"` a `"SUN"` |
| `timing.recurrence.startRecur` | Si recurrente | Fecha de inicio (`YYYY-MM-DD`) |
| `timing.recurrence.endRecur` | Si recurrente | Fecha de fin (vacío = indefinido) |
| `behavior.onEndAction` | No | `"hide"` (predeterminado), `"none"`, o `"stop"` |
| `behavior.preloadSeconds` | No | Segundos de precarga antes del inicio del evento |

---

## 9. Resolución de Problemas

### 9.1 Problemas de Conexión

| Problema | Solución |
|----------|----------|
| **Indicador Server en rojo** | Comprobar que Scene Scheduler está ejecutándose y que el navegador puede alcanzar el host/puerto |
| **Indicador OBS en rojo** | Verificar que OBS está ejecutándose, WebSocket está habilitado y `obs.host`/`port`/`password` coinciden |
| **Indicador Preview en rojo** | Activar VirtualCam en OBS (Herramientas → Iniciar Cámara Virtual) y configurar `mediaSource` |

### 9.2 Errores de Configuración

| Error | Causa | Solución |
|-------|-------|----------|
| `obs.scheduleScene and obs.scheduleSceneAux are required` | Campos obligatorios ausentes | Añadir ambos campos a `config.json` |
| `webServer.hlsPath must be a relative path` | Se usó ruta absoluta | Usar una ruta relativa como `"hls"` |
| `certFilePath and keyFilePath are required when TLS is enabled` | TLS activado sin certificados | Proporcionar rutas de cert/key o poner `enableTls` a `false` |

### 9.3 Problemas de Programación

| Problema | Solución |
|----------|----------|
| **Los eventos no se activan** | Comprobar que `enabled` es `true` y que la hora del evento no ha pasado |
| **Los eventos recurrentes no aparecen** | Verificar que `startRecur` está en el pasado y `daysOfWeek` incluye el día actual |
| **La programación no se guarda** | Comprobar el indicador de sincronización del Editor; asegurar que WebSocket está conectado |

### 9.4 Problemas de Vista Previa

| Problema | Solución |
|----------|----------|
| **El botón Preview no funciona** | Asegurar que el binario `hls-generator` está en el mismo directorio que `scenescheduler` |
| **La vista previa en directo está en blanco** | Comprobar la configuración de `mediaSource` y el estado de VirtualCam en OBS |

### 9.5 Reiniciar

Los cambios en `config.json` requieren reiniciar:

**Linux:**
```bash
pkill scenescheduler
./build/scenescheduler
```

**Windows:**
```cmd
REM Pulsa Ctrl+C en la ventana del símbolo del sistema, luego reinicia
scenescheduler.exe
```

Los cambios de programación (`schedule.json`) se aplican en tiempo real a través de la interfaz web — no es necesario reiniciar.

---

## Licencia

Scene Scheduler es software propietario. Consulte el archivo LICENSE para más detalles.
