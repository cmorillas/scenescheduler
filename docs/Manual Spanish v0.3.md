# Manual de Usuario de Scene Scheduler Beta 0.1

---

## 🚀 1. Instalación Rápida (Para Impacientes)

Esta sección te permite empezar a funcionar en minutos. Sigue estos 4 pasos esenciales:

### Paso 1: Preparación en OBS

**Activa el WebSocket:**
- En OBS Studio, ve al menú **Herramientas → Ajustes del servidor WebSocket**
- Marca la casilla **"Activar el servidor WebSocket"**
- Anota el puerto (por defecto es **4455**)
- Establece una **contraseña segura** y anótala
- Haz clic en **"Aplicar"** y luego en **"Aceptar"**

**Crea las Escenas Requeridas:**
- En OBS, crea dos escenas nuevas y vacías:
  - Haz clic derecho en el panel de Escenas
  - Selecciona **"Añadir" → "Escena"**
  - Crea una escena llamada **Schedule** (será la escena principal visible)
  - Crea otra escena llamada **Schedule_Temp** (escena temporal para staging)
  - **Importante:** Estas escenas deben estar completamente vacías al inicio

### Paso 2: Configuración Mínima

- Descomprime el archivo .zip que has descargado en una carpeta de tu elección
- Abre el fichero **config.json** con un editor de texto (Notepad, Notepad++, VS Code, etc.)
- Rellena obligatoriamente estos campos en la sección "obs":

```json
"obs": {
  "host": "localhost",
  "port": 4455,
  "password": "tu_contraseña_de_obs",
  "scheduleScene": "Schedule",
  "scheduleSceneTmp": "Schedule_Temp"
}
```

- Guarda el archivo config.json

### Paso 3: Ejecución

- Haz doble clic en **scenescheduler.exe** (Windows) o ejecuta `./scenescheduler` (Linux/Mac)
- Se abrirá una ventana de terminal con logs. **No la cierres** - esta ventana debe permanecer abierta
- Espera a que aparezca el mensaje **"WebServer running on port 8080"**
- Abre tu navegador web (Chrome, Firefox, Edge) y ve a: **http://localhost:8080**

### Paso 4: Interfaz Web y Flujo de Trabajo

**Entendiendo la Interfaz:**

La interfaz tiene dos vistas principales:

- **Vista Monitor** (solo lectura): Para observar el sistema en tiempo real
  - Registro de actividad del backend
  - Vista previa en vivo del stream
  - Calendario con el schedule activo del servidor

- **Vista Editor** (edición completa): Para modificar la programación
  - Calendario editable
  - Menú de acciones (⋯) con operaciones de schedule

**Indicadores de Estado:**

En la parte superior verás tres indicadores de estado:

- **Server** (WebSocket): Verde = Conectado al backend | Rojo = Desconectado
- **OBS**: Verde = Backend conectado a OBS | Rojo = OBS desconectado
- **Preview**: Verde = Stream disponible | Naranja = Conectando | Rojo = No disponible

**Flujo Básico de Edición:**

1. **Cargar:** Cambia a la Vista Editor, haz clic en el menú **⋯** y selecciona **"Get from Server"** para cargar la programación actual
2. **Editar:**
   - Haz clic en el calendario para crear eventos nuevos
   - Doble clic en eventos existentes para editarlos
   - Arrastra eventos para moverlos
3. **Guardar:** Cuando termines, vuelve al menú **⋯** y selecciona **"Commit to Server"**. Los cambios se aplicarán automáticamente en OBS

---

## 2. Introducción

### ¡Bienvenido a Scene Scheduler!

Scene Scheduler es una potente herramienta diseñada para automatizar por completo tu producción en OBS Studio. Te permite planificar con antelación qué contenido se mostrará y cuándo, creando una parrilla de programación similar a la de un canal de televisión profesional.

El sistema funciona con un calendario web muy intuitivo donde puedes añadir, mover y editar eventos de forma visual. Una vez guardada la planificación, Scene Scheduler se encarga de cambiar las fuentes en OBS de forma automática, precisa y sin cortes visuales, garantizando una operación continua 24/7.

### Características Principales

- **Automatización Total:** Una vez configurado, Scene Scheduler gestiona todos los cambios de escena sin intervención manual
- **Interfaz Web Dual:** Vista Monitor (observación) y Vista Editor (modificación completa)
- **Triple Sistema de Estado:** Indicadores independientes para Server, OBS y Preview
- **Previsualización en Vivo:** Streaming WebRTC de ultra-baja latencia con protocolo WHEP
- **Cambios Sin Cortes:** Sistema de staging de 5 pasos que garantiza transiciones suaves sin artefactos visuales
- **Eventos Recurrentes:** Programa eventos que se repiten diariamente, semanalmente o en días específicos
- **Hot-Reload:** Los cambios en la programación se aplican automáticamente sin reiniciar
- **Reconexión Automática:** Sistema de reconexión inteligente con sincronización de estado
- **Operación 24/7:** Diseñado para funcionar de forma continua sin interrupciones

### ¿Para quién es este manual?

Este manual está dirigido a los usuarios finales de Scene Scheduler. Te guiaremos paso a paso, desde la configuración inicial hasta la gestión diaria de tu programación, sin necesidad de tener conocimientos técnicos de programación. Cubriremos:

- Instalación y configuración inicial
- Uso del calendario para crear y gestionar eventos
- Configuración de diferentes tipos de fuentes (videos, imágenes, páginas web)
- Solución de problemas comunes
- Mejores prácticas para una operación eficiente

---

## 3. Interfaz Web - Visión General

La interfaz web de Scene Scheduler proporciona dos vistas especializadas para diferentes propósitos.

### 3.1. Sistema de Vistas Dual

**Vista Monitor (Solo Lectura)**

Propósito: Observar el estado actual del sistema sin modificar nada.

Componentes:
- **Registro de Actividad:** Muestra todos los eventos del backend en tiempo real
  - Conexiones y desconexiones
  - Cambios de programa
  - Recarga de schedule
  - Eventos de VirtualCam
- **Previsualización en Vivo:** Stream WebRTC de lo que OBS está emitiendo
  - Protocolo WHEP para latencia ultra-baja
  - Controles de reproducción (play/pause)
  - Conexión bajo demanda (solo cuando se reproduce)
- **Calendario de Solo Lectura:** Visualización del schedule activo del servidor
  - No se pueden crear ni editar eventos
  - Muestra el programa actual resaltado
  - Click en eventos abre modal de solo lectura

**Vista Editor (Edición Completa)**

Propósito: Espacio de trabajo seguro para modificar la programación.

Componentes:
- **Calendario Editable:** Completa funcionalidad de edición
  - Crear eventos: Click o click-and-drag
  - Modificar eventos: Doble click para abrir editor
  - Mover eventos: Arrastrar a nueva posición
  - Cambiar duración: Arrastrar bordes
  - Eliminar eventos: Tecla Delete o botón en modal
- **Menú de Acciones (⋯):** Operaciones principales del schedule
  - New Schedule: Limpiar calendario
  - Load from File: Importar schedule desde JSON local
  - Save to File: Exportar schedule actual a JSON
  - Get from Server: Cargar schedule activo del servidor
  - Commit to Server: Guardar cambios en el servidor
- **Barra de Estado:** Muestra estado de sincronización
  - Verde "Synced with server": Sin cambios pendientes
  - Naranja "X unsaved changes": Cambios sin guardar
  - Azul "Saving...": Operación en curso
  - Rojo: Mensaje de error

### 3.2. Triple Sistema de Indicadores de Estado

En la parte superior de la interfaz web encontrarás tres indicadores independientes que muestran el estado de las conexiones:

**Indicador Server (WebSocket)**

Muestra el estado de la conexión entre el navegador y el backend:
- **Verde:** Conectado al servidor backend
- **Rojo:** Desconectado del servidor backend
- Tooltip: Muestra estado de conexión

Cuando se pierde la conexión, el sistema intenta reconectar automáticamente cada 5 segundos. Al reconectar exitosamente, se re-sincroniza todo el estado (status y schedule) para asegurar información actualizada.

**Indicador OBS (Backend ↔ OBS)**

Muestra el estado de la conexión entre el backend y OBS Studio:
- **Verde:** Backend conectado a OBS
- **Rojo:** Backend no conectado a OBS
- Tooltip: Muestra versión de OBS cuando está conectado

Este indicador refleja si el backend puede comunicarse con OBS Studio a través del protocolo obs-websocket.

**Indicador Preview (VirtualCam Stream)**

Muestra la disponibilidad del stream de previsualización:
- **Verde:** Stream disponible o activamente conectado
- **Naranja:** Conexión WebRTC en progreso
- **Rojo:** Stream no disponible (VirtualCam detenida)

Estados en detalle:
- **unavailable (Rojo):** VirtualCam está detenida en OBS o stream no disponible (503)
- **available (Verde):** VirtualCam activa, stream disponible para reproducir
- **connecting (Naranja):** Estableciendo conexión WebRTC
- **connected (Verde):** WebRTC conectado, stream reproduciendo activamente

**Nota importante:** Las acciones de pausar/reproducir del usuario no cambian el estado de disponibilidad. Solo cambia cuando realmente el stream deja de estar disponible (VirtualCam se detiene, error de red, etc.).

### 3.3. Previsualización en Vivo con WHEP

Scene Scheduler utiliza el protocolo WHEP (WebRTC-HTTP Egress Protocol) para streaming de video de ultra-baja latencia.

**Funcionamiento:**

1. En OBS, haz clic en **"Iniciar cámara virtual"** (VirtualCam)
2. El backend captura este stream y lo prepara para distribución WebRTC
3. En la Vista Monitor, haz clic en el botón **Play** del reproductor
4. El navegador establece conexión WebRTC con el backend
5. El stream se muestra en el reproductor con latencia mínima

**Controles:**

- **Play:** Establece conexión WebRTC y comienza reproducción
- **Pause:** Desconecta la sesión WebRTC (mantiene disponibilidad del stream)
- **Volumen:** Controla nivel de audio

**Comportamiento del Stream:**

- La conexión WebRTC se establece **solo cuando se pulsa Play**
- Al pausar, la sesión WebRTC se desconecta para liberar recursos
- Si el stream sigue disponible, puedes volver a reproducir inmediatamente
- Si VirtualCam se detiene en OBS, el estado cambia a **unavailable** (rojo)
- La conexión se mantiene mientras el stream está disponible y reproduciendo

**Manejo de Errores:**

El sistema distingue entre diferentes estados:
- **503 Service Unavailable:** Respuesta esperada cuando VirtualCam no está activa (no se registra como error)
- **Errores de red:** Se muestran mensajes apropiados
- **Stream remoto finaliza:** Desconexión automática

---

## 4. Instalación y Configuración Detallada

Para poner en marcha Scene Scheduler con todas sus funcionalidades, sigue estos pasos detallados.

### Paso 1: Requisitos del Sistema

Antes de instalar, asegúrate de tener:

- **Sistema Operativo:** Windows 10/11, macOS 10.15+, o Linux (Ubuntu 20.04+)
- **OBS Studio:** Versión 28.0 o superior con WebSocket Plugin
- **Navegador Web:** Chrome 90+, Firefox 88+, Edge 90+ o Safari 14+ (con soporte WebRTC)
- **RAM:** Mínimo 4GB (8GB recomendado)
- **Espacio en Disco:** 100MB para la aplicación + espacio para logs

### Paso 2: Descomprimir los Archivos

Recibirás un archivo .zip con la distribución de Scene Scheduler. Descomprímelo en una carpeta permanente en tu ordenador (evita carpetas temporales o de descargas). Dentro encontrarás:

**Archivos Esenciales:**
- `scenescheduler.exe` (Windows) o `scenescheduler` (Linux/Mac): El programa principal
- `config.json`: El archivo de configuración principal
- `schedule.json`: El archivo donde se guarda tu calendario (inicialmente con ejemplos)

**Archivos Generados:**
- `logs.txt`: Archivo de texto con los logs (se crea automáticamente al ejecutar)
- Archivos `.log` adicionales pueden crearse con fecha/hora según configuración

### Paso 3: Configurar la Conexión (config.json)

Abre el archivo `config.json` con un editor de texto. Este archivo controla todos los aspectos de Scene Scheduler. Vamos a revisar cada sección en detalle:

#### 3.1. Conexión con OBS (Sección "obs")

Esta es la sección más importante y debe configurarse correctamente para que Scene Scheduler funcione.

Antes de empezar:
- Abre OBS Studio
- Ve al menú **Herramientas → Ajustes del servidor WebSocket**
- Asegúrate de que **"Activar el servidor WebSocket"** esté marcado
- Configura un puerto (por defecto 4455) y una contraseña segura
- Crea las dos escenas vacías requeridas en OBS

Parámetros de configuración:

```json
"obs": {
  "host": "localhost",              // Dirección del PC con OBS
  "port": 4455,                     // Puerto del WebSocket
  "password": "tu_contraseña",      // Contraseña del WebSocket
  "reconnectInterval": 5,           // Segundos entre reintentos
  "scheduleScene": "Schedule",      // Nombre de la escena principal
  "scheduleSceneTmp": "Schedule_Temp",  // Escena temporal
  "sourceNamePrefix": "SS_"         // Prefijo para las fuentes
}
```

Notas importantes:
- **host:** Usa "localhost" si OBS está en el mismo PC. Para control remoto, usa la IP del PC con OBS
- **scheduleScene y scheduleSceneTmp:** Los nombres deben coincidir EXACTAMENTE con las escenas en OBS
- **sourceNamePrefix:** Todas las fuentes creadas por Scene Scheduler tendrán este prefijo para identificación

#### 3.2. Servidor Web (Sección "webServer")

Configura el acceso a la interfaz del calendario web:

```json
"webServer": {
  "port": "8080",           // Puerto para la interfaz web
  "user": "",               // Usuario (vacío = sin autenticación)
  "password": "",           // Contraseña (vacío = sin autenticación)
  "hlsPath": "hls",         // Directorio para previsualizaciones HLS
  "enableTls": false,       // HTTPS activado/desactivado
  "certFilePath": "",       // Ruta al certificado SSL
  "keyFilePath": ""         // Ruta a la clave SSL
}
```

Configuraciones comunes:
- **Acceso local sin seguridad:** Deja user y password vacíos
- **Acceso protegido:** Establece user y password para requerir autenticación
- **HTTPS:** Configura `enableTls: true` y proporciona los archivos de certificado

Notas sobre hlsPath:
- **Debe ser un path relativo** (ej: "hls", "data/previews")
- No se permiten paths absolutos (ej: "/etc/hls") por seguridad
- No se permite navegación de directorios (ej: "../hls")

#### 3.3. Planificador (Sección "scheduler")

Define qué mostrar cuando no hay eventos programados:

```json
"scheduler": {
  "defaultSource": {
    "name": "standby_image",
    "inputKind": "image_source",
    "uri": "C:/imagenes/standby.png",
    "inputSettings": {
      "file": "C:/imagenes/standby.png"
    },
    "transform": {
      "positionX": 0,
      "positionY": 0,
      "scaleX": 1.0,
      "scaleY": 1.0
    }
  }
}
```

Tipos de fuente por defecto:
- Imagen estática: `inputKind: "image_source"`
- Video en bucle: `inputKind: "ffmpeg_source"`
- Página web: `inputKind: "browser_source"`

#### 3.4. Previsualización en Vivo (Sección "mediaSource")

Configura la captura para previsualización:

```json
"mediaSource": {
  "videoDeviceIdentifier": "OBS Virtual Camera",
  "audioDeviceIdentifier": "default",
  "quality": "low"  // "low", "medium", o "high"
}
```

Configuración paso a paso:
- En OBS, haz clic en **"Iniciar cámara virtual"**
- Ejecuta `scene-scheduler -list-devices` para ver dispositivos disponibles
- Copia el nombre exacto del dispositivo en `videoDeviceIdentifier`

#### 3.5. Rutas de Archivos (Sección "paths")

Define ubicaciones de archivos importantes:

```json
"paths": {
  "logFile": "./scene-scheduler.log",   // Archivo de logs
  "schedule": "./schedule.json"         // Archivo de programación
}
```

---

## 5. Fichero de Planificación (schedule.json) - Formato Completo

El archivo `schedule.json` es el corazón de Scene Scheduler. Contiene toda tu programación y debe seguir un formato JSON estricto. A continuación se detalla la estructura completa con todos los campos disponibles.

### 5.1. Estructura General del Archivo

El archivo completo está envuelto en un objeto que contiene metadatos y el array de eventos:

```json
{
  "version": "1.0",
  "scheduleName": "Mi Programación de Streaming",
  "schedule": [
    // Array de eventos (programas)
  ]
}
```

Campos principales:
- **version:** Versión del formato (actualmente "1.0")
- **scheduleName:** Nombre descriptivo de tu programación
- **schedule:** Array que contiene todos los eventos programados

### 5.2. Estructura de un Evento

Cada elemento del array schedule es un objeto con la siguiente estructura:

```json
{
  "id": "evt-001",
  "title": "Programa de la Mañana",
  "enabled": true,
  "general": { /* ... */ },
  "source": { /* ... */ },
  "timing": { /* ... */ },
  "behavior": { /* ... */ }
}
```

### 5.3. Campos Principales del Evento

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| id | string | Sí | Identificador único del evento (ej: "evt-001") |
| title | string | Sí | Título descriptivo que aparece en el calendario |
| enabled | boolean | Sí | Si es true el evento se ejecutará, si es false se ignora |
| general | object | No | Configuración visual y metadata |
| source | object | Sí | Define qué contenido mostrar en OBS |
| timing | object | Sí | Define cuándo se ejecuta el evento |
| behavior | object | No | Comportamientos automáticos |

### 5.4. Sección "general" - Apariencia y Metadata

```json
"general": {
  "description": "Noticias matutinas con el equipo de producción",
  "tags": ["noticias", "diario", "prioritario"],
  "classNames": ["high-priority", "news-segment"],
  "textColor": "#FFFFFF",
  "backgroundColor": "#2196F3",
  "borderColor": "#1976D2"
}
```

| Campo | Tipo | Descripción | Ejemplo |
|-------|------|-------------|---------|
| description | string | Texto descriptivo del evento | "Segmento de entrevistas" |
| tags | array[string] | Etiquetas para categorización | ["entrevista", "live"] |
| classNames | array[string] | Clases CSS personalizadas | ["premium-content"] |
| textColor | string | Color hexadecimal del texto | "#FFFFFF" |
| backgroundColor | string | Color hexadecimal de fondo | "#FF5722" |
| borderColor | string | Color hexadecimal del borde | "#E64A19" |

### 5.5. Sección "source" - Configuración del Contenido

Esta sección define exactamente qué contenido mostrará OBS:

```json
"source": {
  "name": "morning_news_feed",
  "inputKind": "ffmpeg_source",
  "uri": "C:/Videos/morning_news.mp4",
  "inputSettings": {
    "local_file": true,
    "looping": false,
    "restart_on_activate": true
  },
  "transform": {
    "positionX": 0,
    "positionY": 0,
    "scaleX": 1.0,
    "scaleY": 1.0
  }
}
```

Campos de source:

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| name | string | Sí | Nombre técnico único de la fuente (sin espacios) |
| inputKind | string | Sí | Tipo de fuente OBS (ver tipos disponibles) |
| uri | string | Sí* | Ubicación del contenido (ruta o URL) |
| inputSettings | object | No | Configuración específica del tipo de fuente |
| transform | object | No | Posición y transformación en la escena |

Tipos de inputKind disponibles:
- `ffmpeg_source`: Videos locales y streams
- `browser_source`: Páginas web y contenido HTML
- `image_source`: Imágenes estáticas
- `vlc_source`: Reproducción con VLC

### 5.6. Sección "timing" - Programación Temporal

**IMPORTANTE:** Los campos start y end deben usar formato ISO 8601 con zona horaria (Z para UTC):

```json
"timing": {
  "start": "2024-03-15T09:00:00Z",
  "end": "2024-03-15T10:30:00Z",
  "isRecurring": false,
  "recurrence": {
    "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
    "startRecur": "2024-01-01",
    "endRecur": "2024-12-31"
  }
}
```

Campos de timing:

| Campo | Tipo | Formato | Descripción |
|-------|------|---------|-------------|
| start | string | ISO 8601 | Fecha/hora de inicio: YYYY-MM-DDTHH:MM:SSZ |
| end | string | ISO 8601 | Fecha/hora de fin: YYYY-MM-DDTHH:MM:SSZ |
| isRecurring | boolean | - | Si es true, el evento se repite |
| recurrence | object | - | Configuración de recurrencia (si aplica) |

Campos de recurrence:

| Campo | Tipo | Formato | Descripción |
|-------|------|---------|-------------|
| daysOfWeek | array | - | Días de repetición: ["MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"] |
| startRecur | string | YYYY-MM-DD | Primera fecha de la serie recurrente |
| endRecur | string | YYYY-MM-DD | Última fecha de la serie recurrente |

Nota sobre eventos recurrentes: Para eventos que se repiten, los campos start y end definen solo la hora del día (la parte de tiempo), mientras que las fechas de repetición se controlan con startRecur y endRecur.

### 5.7. Sección "behavior" - Comportamiento Automático

```json
"behavior": {
  "onEndAction": "hide",
  "preloadSeconds": 30
}
```

| Campo | Tipo | Valores | Descripción |
|-------|------|---------|-------------|
| onEndAction | string | "hide", "stop", "none" | Acción al finalizar el evento |
| preloadSeconds | number | 0-300 | Segundos para precargar antes del inicio |

### 5.8. Ejemplo Completo de schedule.json

```json
{
  "version": "1.0",
  "scheduleName": "Programación Canal Web TV",
  "schedule": [
    {
      "id": "morning-news-001",
      "title": "Noticias de la Mañana",
      "enabled": true,
      "general": {
        "description": "Resumen informativo matutino con las últimas noticias",
        "tags": ["noticias", "informativo", "diario"],
        "classNames": ["news-program", "high-priority"],
        "textColor": "#FFFFFF",
        "backgroundColor": "#1E88E5",
        "borderColor": "#1565C0"
      },
      "source": {
        "name": "morning_news_source",
        "inputKind": "browser_source",
        "uri": "https://news.example.com/live",
        "inputSettings": {
          "url": "https://news.example.com/live",
          "width": 1920,
          "height": 1080,
          "fps": 30,
          "css": "body { overflow: hidden; }"
        },
        "transform": {
          "positionX": 0,
          "positionY": 0,
          "scaleX": 1.0,
          "scaleY": 1.0
        }
      },
      "timing": {
        "start": "2024-03-15T09:00:00Z",
        "end": "2024-03-15T10:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2024-03-01",
          "endRecur": "2024-12-31"
        }
      },
      "behavior": {
        "onEndAction": "hide",
        "preloadSeconds": 30
      }
    },
    {
      "id": "lunch-break-002",
      "title": "Pantalla de Pausa",
      "enabled": true,
      "general": {
        "description": "Imagen estática durante el horario de almuerzo",
        "tags": ["pausa", "imagen", "diario"],
        "classNames": ["break-screen"],
        "textColor": "#000000",
        "backgroundColor": "#4CAF50",
        "borderColor": "#388E3C"
      },
      "source": {
        "name": "lunch_break_image",
        "inputKind": "image_source",
        "uri": "C:/Images/lunch_break.png",
        "inputSettings": {
          "file": "C:/Images/lunch_break.png",
          "unload": false
        },
        "transform": {
          "positionX": 0,
          "positionY": 0,
          "scaleX": 1.0,
          "scaleY": 1.0
        }
      },
      "timing": {
        "start": "2024-03-15T12:00:00Z",
        "end": "2024-03-15T13:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2024-03-01",
          "endRecur": "2024-12-31"
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

### 5.9. Notas Importantes sobre el Formato

- **Formato de Fecha/Hora ISO 8601:**
  - Siempre usa el formato YYYY-MM-DDTHH:MM:SSZ
  - La T separa fecha y hora
  - La Z al final indica UTC (tiempo universal)
  - Ejemplo: 2024-03-15T09:00:00Z = 15 de marzo 2024, 9:00 AM UTC

- **IDs únicos:** Cada evento debe tener un id único en todo el archivo

- **Validación JSON:** El archivo debe ser JSON válido (cuidado con comas y comillas)

- **Campos opcionales:** Solo id, title, enabled, source y timing son obligatorios

- **Eventos deshabilitados:** Los eventos con `enabled: false` permanecen en el archivo pero no se ejecutan

---

## 6. Gestionando tu Programación

### 6.1. Conceptos Fundamentales

Antes de crear eventos, es importante entender estos conceptos:

- **Evento/Programa:** Una unidad de contenido con hora de inicio y fin
- **Fuente (Source):** El contenido real que OBS mostrará (video, imagen, web)
- **Escena:** El contenedor en OBS donde se colocan las fuentes
- **Recurrencia:** Eventos que se repiten automáticamente según un patrón
- **Schedule del Servidor:** El schedule oficial que el backend está ejecutando
- **Working Schedule:** Copia local en el Editor que puede divergir del servidor

### 6.2. El Menú de Acciones (⋯)

Ubicado en la esquina superior derecha del calendario en la Vista Editor, contiene las acciones principales:

#### Opciones del Menú:

**1. New Schedule**
- Limpia completamente el calendario
- **Advertencia:** Esta acción no se puede deshacer
- Útil para empezar una programación desde cero

**2. Load from File**
- Carga una programación desde un archivo .json en tu PC
- Permite mantener múltiples programaciones y cambiar entre ellas
- No afecta la programación activa en el servidor hasta hacer "Commit"

**3. Save to File**
- Guarda la programación actual en un archivo .json
- Útil para hacer copias de seguridad
- Incluye todos los eventos y sus configuraciones

**4. Get from Server ⭐ (Acción Principal)**
- Carga la programación activa del servidor
- Sincroniza tu calendario con lo que Scene Scheduler está usando
- Siempre úsalo al iniciar para estar sincronizado
- Si hay cambios sin guardar, pedirá confirmación

**5. Commit to Server ⭐ (Acción Principal)**
- Guarda todos los cambios en el servidor
- Los cambios se aplican inmediatamente en OBS
- Scene Scheduler recarga automáticamente la nueva programación
- Actualiza el estado a "Synced with server"

### 6.3. Crear y Modificar Eventos

#### Crear un Evento Nuevo:

**Método 1: Click Simple**
- Haz clic en cualquier espacio vacío del calendario
- Se abrirá el modal de creación con la hora seleccionada
- El evento tendrá duración predeterminada de 1 hora

**Método 2: Click y Arrastrar**
- Haz clic y mantén presionado en la hora de inicio
- Arrastra hasta la hora de fin deseada
- Suelta para crear el evento con la duración exacta

#### Modificar Eventos Existentes:

**Editar Detalles:**
- Doble clic sobre el evento para abrir el editor completo
- Modifica cualquier parámetro y guarda los cambios

**Mover en el Tiempo:**
- Click y arrastra el evento a su nueva posición
- El evento mantendrá su duración original

**Cambiar Duración:**
- Posiciona el cursor en el borde inferior del evento
- Arrastra hacia arriba o abajo para ajustar la duración

**Eliminar Eventos:**
- Selecciona el evento haciendo clic sobre él
- Presiona la tecla Delete o Supr
- O abre el editor y usa el botón "Delete"

### 6.4. El Modal de Edición: Configuración Detallada

El modal de edición es donde configuras todos los aspectos de un evento. Se organiza en cuatro pestañas:

#### Pestaña "General" - Información Básica

**1. Title (Obligatorio)**
- Nombre descriptivo del evento
- Se muestra en el calendario y en los registros
- Ejemplos: "Noticias Matutinas", "Video Promocional", "Pausa Técnica"

**2. Enabled**
- Casilla de verificación para activar/desactivar el evento
- Eventos deshabilitados permanecen en el calendario pero no se ejecutan
- Útil para programación temporal o pruebas

**3. Description**
- Texto descriptivo opcional
- Notas internas sobre el evento
- No afecta la operación, solo informativo

**4. Tags**
- Etiquetas separadas por espacios
- Facilitan la búsqueda y categorización
- Ejemplos: "noticias", "publicidad", "educativo"

**5. ClassNames**
- Clases CSS personalizadas para estilizado avanzado
- Para usuarios avanzados que quieran personalizar la apariencia

**6. Colors**
- **Text Color:** Color del texto en el calendario
- **Background Color:** Color de fondo del evento
- **Border Color:** Color del borde (útil para eventos recurrentes)
- Usa el selector de color o introduce códigos hexadecimales

#### Pestaña "Source" - Configuración del Contenido

Esta es la pestaña más importante, define qué contenido mostrará OBS.

**1. Input Name (Obligatorio)**
- Nombre técnico único de la fuente en OBS
- No uses espacios ni caracteres especiales
- Scene Scheduler añadirá automáticamente el prefijo configurado
- Ejemplos: "video_intro", "imagen_pausa", "web_noticias"

**2. Input Kind (Obligatorio)**
- Tipo de fuente de OBS a crear
- Opciones comunes:
  - `ffmpeg_source`: Videos y streams de medios
  - `image_source`: Imágenes estáticas
  - `browser_source`: Páginas web y HTML
  - `vlc_source`: Videos con VLC (si está instalado)

**3. URI (Obligatorio según el tipo)**
- Ubicación del contenido
- Para archivos locales: Ruta completa (ej: C:/videos/intro.mp4)
- Para contenido web: URL completa (ej: https://example.com)
- Para imágenes: Ruta al archivo de imagen

**4. Input Settings (JSON)**

Configuración específica del tipo de fuente. Ejemplos por tipo:

Para ffmpeg_source (videos):
```json
{
  "local_file": true,
  "is_local_file": true,
  "looping": true,
  "restart_on_activate": true,
  "clear_on_media_end": false
}
```

Para browser_source (web):
```json
{
  "url": "https://example.com",
  "width": 1920,
  "height": 1080,
  "fps": 30,
  "css": "body { background-color: transparent; }"
}
```

Para image_source:
```json
{
  "file": "C:/imagenes/logo.png",
  "unload": false
}
```

**5. Transform (JSON)**

Posición y transformación de la fuente en la escena:

```json
{
  "positionX": 0,           // Posición horizontal (píxeles)
  "positionY": 0,           // Posición vertical (píxeles)
  "scaleX": 1.0,            // Escala horizontal (1.0 = 100%)
  "scaleY": 1.0,            // Escala vertical
  "rotation": 0,            // Rotación en grados
  "cropTop": 0,             // Recorte superior (píxeles)
  "cropBottom": 0,          // Recorte inferior
  "cropLeft": 0,            // Recorte izquierdo
  "cropRight": 0            // Recorte derecho
}
```

#### Pestaña "Timing" - Programación Temporal

Define cuándo y cómo se programa el evento.

**Para Eventos Únicos:**

**1. Start Date/Time**
- Fecha y hora exacta de inicio
- Usa el selector de fecha/hora o escribe directamente
- Formato: YYYY-MM-DD HH:MM:SS

**2. End Date/Time**
- Fecha y hora exacta de finalización
- Debe ser posterior a la hora de inicio
- Define la duración total del evento

**Para Eventos Recurrentes:**

**1. Recurring (Casilla de verificación)**
- Activa el modo de recurrencia
- Cambia el comportamiento de los campos de fecha

**2. Recurrence Pattern**
- **Days of Week:** Selecciona los días que se repite
  - Lunes a Domingo disponibles
  - Puedes seleccionar múltiples días
- **Time:** Para eventos recurrentes, solo se usa la hora de Start/End
- **Date Range:**
  - **Start Recur:** Primera fecha de la serie
  - **End Recur:** Última fecha de la serie

Ejemplos de Recurrencia:
- **Diario a las 9 AM:** Todos los días marcados, Start: 09:00, End: 10:00
- **Lunes a Viernes:** Solo días laborables marcados
- **Fines de semana:** Solo Sábado y Domingo marcados

#### Pestaña "Behavior" - Comportamiento Avanzado

**1. Preload Seconds**
- Segundos de anticipación para preparar la fuente
- Útil para videos pesados o streams de red
- Valor 0 = carga justo al momento del cambio

**2. On End Action**
- Qué hacer cuando el evento termina:
  - **hide:** Ocultar la fuente (predeterminado)
  - **stop:** Detener y liberar recursos
  - **none:** No hacer nada (mantener visible)

### 6.5. Mejores Prácticas para la Programación

#### Organización Eficiente:

- **Usa nombres descriptivos:** Facilita la identificación rápida
- **Codificación por colores:** Asigna colores por categoría (ej: azul para noticias, verde para publicidad)
- **Tags consistentes:** Crea un sistema de etiquetas y úsalo consistentemente
- **Documenta con descripciones:** Añade notas importantes en el campo descripción

#### Evitar Problemas:

- **No superponer eventos:** Scene Scheduler ejecutará el más reciente
- **Verifica rutas de archivos:** Asegúrate de que todos los archivos existen
- **Prueba antes de emitir:** Usa eventos deshabilitados para probar
- **Backup regular:** Guarda copias de tu programación frecuentemente

#### Optimización de Recursos:

- **Reutiliza fuentes:** Usa el mismo Input Name para contenido que se repite
- **Preload estratégico:** Configura preload solo donde sea necesario
- **Limpieza periódica:** Elimina eventos antiguos que ya no necesitas

---

## 7. Cómo Funciona el Sistema de Cambio

### 7.1. El Proceso de Cambio Seguro

Scene Scheduler utiliza un sofisticado sistema de "staging" para garantizar cambios sin artefactos visuales. Este proceso de 5 pasos asegura que tu audiencia nunca vea cortes, pantallas negras o errores durante las transiciones.

#### Los 5 Pasos del Cambio:

**Paso 1: STAGING (Preparación)**
- La nueva fuente se crea en la escena temporal (Schedule_Temp)
- Se configura completamente pero permanece oculta
- Se aplican todas las transformaciones (posición, escala, etc.)
- Si falla: El proceso se detiene sin afectar la emisión actual

**Paso 2: PROMOTION (Promoción)**
- El elemento preparado se duplica a la escena principal (Schedule)
- Todavía permanece oculto en la escena principal
- Se verifica que la duplicación fue exitosa
- Si falla: Se ejecuta rollback completo

**Paso 3: ACTIVATION (Activación)**
- Se hace visible el nuevo elemento en la escena principal
- Este es el momento exacto del cambio para la audiencia
- El cambio es instantáneo y sin cortes
- Si falla: Rollback y se mantiene el contenido anterior

**Paso 4: CLEANUP (Limpieza del Staging)**
- Se elimina el elemento temporal de Schedule_Temp
- Se liberan recursos no necesarios
- La escena temporal queda lista para el próximo cambio

**Paso 5: RETIREMENT (Retirada del Anterior)**
- Se oculta el programa anterior en la escena principal
- Se elimina completamente después de ocultarlo
- Se liberan todos los recursos del contenido anterior

### 7.2. Ventajas del Sistema de Staging

**1. Cambios Sin Cortes**
- No hay frames negros entre transiciones
- No hay parpadeos o artefactos visuales
- La audiencia ve un cambio limpio e instantáneo

**2. Seguridad y Rollback**
- Si algo falla, el contenido actual continúa
- Cada paso valida antes de continuar
- Sistema de rollback automático en caso de error

**3. Preparación Anticipada**
- Las fuentes pesadas se cargan antes del cambio
- Videos y streams tienen tiempo para buffer
- Reduce la carga del sistema en el momento del cambio

### 7.3. Logs y Diagnóstico de Cambios

La terminal del backend muestra información detallada de cada cambio:

**Mensajes de Información (Debug):**
- Creating source in TEMP scene: Inicio del staging
- Duplicating to MAIN scene: Promoción exitosa
- Activating in MAIN scene: Cambio visible
- Cleanup completed: Proceso finalizado

**Mensajes de Advertencia:**
- Source already exists: Reutilizando fuente existente
- Transform partially applied: Algunos parámetros no se aplicaron
- Cleanup skipped: Elementos no encontrados para limpiar

**Mensajes de Error:**
- Failed to create source: No se pudo crear la fuente
- Duplication failed: Error al promocionar a escena principal
- Activation failed - rollback initiated: Cambio abortado

---

## 8. Casos de Uso Comunes

### 8.1. Transmisión de TV/Radio Online

Configuración típica:

```json
{
  "title": "Programa Matutino",
  "source": {
    "inputKind": "ffmpeg_source",
    "uri": "rtmp://servidor/live/stream"
  },
  "timing": {
    "isRecurring": true,
    "recurrence": {
      "daysOfWeek": ["MON","TUE","WED","THU","FRI"],
      "startRecur": "2024-01-01",
      "endRecur": "2024-12-31"
    }
  }
}
```

### 8.2. Pantallas Informativas

Para lobbies, salas de espera, comercios:

```json
{
  "title": "Información del Día",
  "source": {
    "inputKind": "browser_source",
    "uri": "https://tuempresa.com/pantalla-info",
    "inputSettings": {
      "width": 1920,
      "height": 1080,
      "fps": 30
    }
  },
  "timing": {
    "start": "08:00:00",
    "end": "20:00:00",
    "isRecurring": true
  }
}
```

### 8.3. Streaming de Videojuegos/Eventos

Para torneos y eventos programados:

```json
{
  "title": "Torneo CS:GO - Semifinales",
  "source": {
    "inputKind": "game_capture",
    "inputSettings": {
      "capture_mode": "window",
      "window": "Counter-Strike: Global Offensive"
    }
  },
  "timing": {
    "start": "2024-03-15T19:00:00",
    "end": "2024-03-15T23:00:00"
  }
}
```

### 8.4. Contenido Educativo

Clases y tutoriales programados:

```json
{
  "title": "Clase de Matemáticas - Álgebra",
  "source": {
    "inputKind": "ffmpeg_source",
    "uri": "C:/Clases/algebra_leccion_5.mp4",
    "inputSettings": {
      "local_file": true,
      "looping": false,
      "restart_on_activate": true
    }
  }
}
```

---

## 9. Apéndice y Solución de Problemas

### A.1. Referencia Completa del Fichero config.json

Esta sección detalla todas las opciones disponibles en el archivo config.json, agrupadas por sección.

#### Sección "obs" - Conexión con OBS

| Clave | Descripción | Requerido | Valor por Defecto | Tipo |
|-------|-------------|-----------|-------------------|------|
| host | Dirección del PC donde corre OBS | No | "localhost" | string |
| port | Puerto del servidor WebSocket de OBS | No | 4455 | integer |
| password | Contraseña del WebSocket. Vacío = sin auth | No | "" | string |
| reconnectInterval | Segundos entre reintentos de conexión | No | 5 | integer |
| scheduleScene | Nombre de la escena principal visible | Sí | N/A | string |
| scheduleSceneTmp | Nombre de la escena temporal de staging | Sí | N/A | string |
| sourceNamePrefix | Prefijo para identificar fuentes gestionadas | No | "SS_" | string |

Notas importantes:
- Los nombres de scheduleScene y scheduleSceneTmp deben coincidir EXACTAMENTE con las escenas en OBS
- El sourceNamePrefix se usa para identificar y limpiar fuentes huérfanas automáticamente

#### Sección "webServer" - Servidor Web

| Clave | Descripción | Requerido | Valor por Defecto | Tipo |
|-------|-------------|-----------|-------------------|------|
| port | Puerto para la interfaz web | No | "8080" | string |
| user | Usuario para autenticación básica | No | "" | string |
| password | Contraseña para autenticación básica | No | "" | string |
| hlsPath | Directorio para previsualizaciones HLS (relativo) | No | "hls" | string |
| enableTls | Activar HTTPS | No | false | boolean |
| certFilePath | Ruta al certificado SSL | Condicional* | "" | string |
| keyFilePath | Ruta a la clave privada SSL | Condicional* | "" | string |

*Requerido si enableTls es true

Configuraciones de seguridad:
- Sin protección: Deja user y password vacíos (solo para uso local)
- Autenticación básica: Establece user y password
- HTTPS: Configura enableTls: true con certificados válidos

Restricciones de hlsPath:
- Solo se permiten paths relativos al directorio de ejecución
- No se aceptan paths absolutos (ej: "/var/hls")
- No se permite navegación de directorios (ej: "../data")

#### Sección "scheduler" - Planificador

| Clave | Descripción | Requerido | Valor por Defecto | Tipo |
|-------|-------------|-----------|-------------------|------|
| defaultSource | Fuente a mostrar cuando no hay eventos | No | null | object |

Estructura de defaultSource:

```json
{
  "name": "string",           // Nombre de la fuente
  "inputKind": "string",      // Tipo (image_source, ffmpeg_source, etc.)
  "uri": "string",            // Ruta o URL del contenido
  "inputSettings": {},        // Configuración específica del tipo
  "transform": {}             // Posición y transformación
}
```

#### Sección "mediaSource" - Previsualización

| Clave | Descripción | Requerido | Valor por Defecto | Tipo |
|-------|-------------|-----------|-------------------|------|
| videoDeviceIdentifier | Nombre del dispositivo de video | No | "" | string |
| audioDeviceIdentifier | Nombre del dispositivo de audio | No | "default" | string |
| quality | Calidad de codificación | No | "low" | string |

Valores de quality: "low", "medium", "high"

#### Sección "paths" - Rutas de Sistema

| Clave | Descripción | Requerido | Valor por Defecto | Tipo |
|-------|-------------|-----------|-------------------|------|
| logFile | Archivo de logs | No | "./scene-scheduler.log" | string |
| schedule | Archivo de programación | No | "./schedule.json" | string |

### A.2. Herramienta de Línea de Comandos

Scene Scheduler incluye herramientas útiles por línea de comandos:

#### Listar Dispositivos (-list-devices)

Para encontrar los identificadores exactos de dispositivos:

Windows:
```
scene-scheduler.exe -list-devices
```

Linux/Mac:
```
./scene-scheduler -list-devices
```

Salida ejemplo:
```
----------- Available Media Devices -----------
INFO: Use the 'Friendly Name' or 'DeviceID' for your config.

Device #0:
  - Kind          : Video Input
  - Friendly Name : OBS Virtual Camera
  - DeviceID      : v4l2:/dev/video6

Device #1:
  - Kind          : Audio Input
  - Friendly Name : Monitor of Built-in Audio Analog Stereo
  - DeviceID      : alsa:pulse_

----------------------------------------------
```

Copia el "Friendly Name" o "DeviceID" exacto en tu config.json.

#### Validar Configuración (-validate)

Verifica que tu configuración sea válida:

```
./scene-scheduler -validate
```

#### Modo Debug (-debug)

Inicia con logging detallado para diagnóstico:

```
./scene-scheduler -debug
```

### A.3. Ejemplo Completo de config.json

Aquí tienes un ejemplo completamente funcional con todas las secciones:

```json
{
  "scheduler": {
    "defaultSource": {
      "name": "standby_screen",
      "inputKind": "image_source",
      "uri": "C:/Scene-Scheduler/assets/standby.png",
      "inputSettings": {
        "file": "C:/Scene-Scheduler/assets/standby.png",
        "unload": false
      },
      "transform": {
        "positionX": 0,
        "positionY": 0,
        "scaleX": 1.0,
        "scaleY": 1.0
      }
    }
  },
  "mediaSource": {
    "videoDeviceIdentifier": "OBS Virtual Camera",
    "audioDeviceIdentifier": "default",
    "quality": "medium"
  },
  "webServer": {
    "port": "8080",
    "user": "admin",
    "password": "secure_password_123",
    "hlsPath": "hls",
    "enableTls": false,
    "certFilePath": "",
    "keyFilePath": ""
  },
  "obs": {
    "host": "localhost",
    "port": 4455,
    "password": "obs_websocket_password",
    "reconnectInterval": 5,
    "scheduleScene": "Schedule",
    "scheduleSceneTmp": "Schedule_Temp",
    "sourceNamePrefix": "SS_"
  },
  "paths": {
    "logFile": "./scene-scheduler.log",
    "schedule": "./schedule.json"
  }
}
```

### A.4. Solución de Problemas Comunes

#### Problemas de Inicio

**La aplicación se cierra inmediatamente:**
- **Causa:** Error en config.json
- **Solución:**
  - Verifica la sintaxis JSON (comas, comillas, llaves)
  - Asegúrate que scheduleScene y scheduleSceneTmp están definidos
  - Ejecuta con -validate para ver errores específicos

**Error "Cannot parse config file":**
- **Causa:** JSON malformado
- **Solución:** Usa un validador JSON online o un editor con resaltado de sintaxis

**Mensaje "Scene Scheduler has expired":**
- **Causa:** La versión beta ha expirado
- **Solución:** Contacta al desarrollador para obtener una versión actualizada

#### Problemas de Conexión con OBS

**"Failed to connect to OBS":**
- **Causas y soluciones:**
  - OBS no está ejecutándose → Inicia OBS primero
  - WebSocket no activado → Herramientas > Ajustes del servidor WebSocket
  - Puerto incorrecto → Verifica que coincida con OBS
  - Contraseña incorrecta → Revisa la contraseña en ambos lados
  - Firewall bloqueando → Añade excepción para Scene Scheduler

**"Scene not found":**
- **Causa:** Las escenas no existen en OBS
- **Solución:** Crea las escenas exactamente como están en config.json

**Conexión intermitente:**
- **Causa:** Red inestable o OBS sobrecargado
- **Solución:** Aumenta reconnectInterval a 10-15 segundos

#### Problemas con la Interfaz Web

**No puedo acceder al calendario:**
- **Verificaciones:**
  - La terminal muestra "WebServer running on port 8080"
  - Usas la URL correcta: http://localhost:[puerto]
  - El firewall no bloquea el puerto
  - Si hay autenticación, usas las credenciales correctas

**El calendario no carga:**
- **Causa:** Problemas con el servidor web embebido
- **Solución:** Reinicia Scene Scheduler y verifica que el puerto no esté ocupado

**WebSocket desconectado constantemente:**
- **Causas:**
  - Proxy o VPN interfiriendo
  - Extensiones del navegador bloqueando WebSockets
  - Timeout por inactividad
- **Solución:** Prueba en modo incógnito o diferente navegador

#### Problemas con Eventos

**Los eventos no se ejecutan:**
- **Verificaciones:**
  - El evento está habilitado (enabled: true)
  - La fecha/hora es correcta
  - No hay eventos superpuestos
  - El archivo/URL del source existe

**Error al crear fuente:**
- **Causas comunes:**
  - Tipo de fuente no soportado
  - Archivo no encontrado
  - URL inaccesible
  - Settings JSON inválido

**Videos que no se reproducen:**
- **Solución:**
  - Verifica que el archivo existe y no está corrupto
  - Usa rutas absolutas, no relativas
  - Para ffmpeg_source, añade: "local_file": true
  - Instala códecs necesarios en el sistema

#### Problemas de Rendimiento

**Alto uso de CPU:**
- **Causas:**
  - Demasiados eventos browser_source activos
  - Videos en resolución muy alta
  - Transforms complejos
- **Soluciones:**
  - Reduce la calidad de previsualización
  - Optimiza los videos antes de usarlos
  - Cierra pestañas innecesarias del calendario

**Memoria aumentando constantemente:**
- **Causa:** Fuentes no se liberan correctamente
- **Solución:**
  - Reinicia Scene Scheduler diariamente
  - Usa onEndAction: "stop" para videos pesados

### A.5. Mensajes de Error Comunes y Soluciones

| Mensaje de Error | Significado | Solución |
|------------------|-------------|----------|
| Config file not found | No existe config.json | Crea o restaura el archivo |
| Invalid JSON in config | Sintaxis JSON incorrecta | Valida el JSON |
| Schedule file not found | No existe schedule.json | Se creará automáticamente |
| OBS connection refused | OBS rechaza la conexión | Verifica puerto y contraseña |
| Scene does not exist | Escena no encontrada en OBS | Crea las escenas requeridas |
| Source creation failed | No se pudo crear la fuente | Verifica tipo y parámetros |
| WebSocket upgrade failed | Error en handshake WS | Revisa configuración de red |
| Permission denied | Sin permisos de archivo | Ejecuta como administrador |
| Port already in use | Puerto ocupado | Cambia el puerto o cierra la otra aplicación |

---

## 10. Mejores Prácticas y Recomendaciones

### 10.1. Configuración Inicial

- **Planifica tu estructura:** Antes de empezar, diseña tu programación en papel
- **Prueba localmente:** Configura y prueba todo en local antes de producción
- **Documenta tu configuración:** Mantén notas de tu setup específico
- **Backup de configuración:** Guarda copias de config.json y schedule.json

### 10.2. Operación Diaria

- **Sincronización matutina:** Siempre usa "Get from Server" al iniciar el día
- **Guardado frecuente:** Haz "Commit to Server" después de cambios importantes
- **Monitoreo regular:** Revisa la Vista Monitor periódicamente
- **Logs para diagnóstico:** Revisa los logs si algo no funciona como esperas

### 10.3. Mantenimiento

- **Limpieza semanal:** Elimina eventos antiguos del calendario
- **Actualización de contenido:** Verifica que todos los archivos referenciados existen
- **Reinicio programado:** Considera reiniciar Scene Scheduler semanalmente
- **Respaldos regulares:** Exporta tu programación a archivo regularmente

### 10.4. Seguridad

- **Contraseñas fuertes:** Usa contraseñas seguras para WebSocket y web
- **Acceso limitado:** En producción, usa autenticación en el servidor web
- **Red segura:** Para acceso remoto, considera usar VPN
- **Permisos de archivos:** Limita quién puede modificar config.json

---

## 11. Glosario de Términos

- **Backend:** La parte del servidor de Scene Scheduler que gestiona la lógica
- **Commit:** Guardar cambios en el servidor para que se apliquen
- **Editor View:** Vista de edición con calendario editable y menú de acciones
- **EventBus:** Sistema interno para comunicación entre módulos
- **Frontend:** La interfaz web del calendario
- **Hot-reload:** Recarga automática sin reiniciar la aplicación
- **Input/Source:** Fuente de contenido en OBS (video, imagen, web)
- **Modal:** Ventana de edición de eventos
- **Monitor View:** Vista de solo lectura con registro de actividad y preview en vivo
- **Prefijo (Prefix):** Texto añadido al inicio de nombres de fuentes
- **Rollback:** Revertir cambios si algo falla
- **Scene:** Contenedor en OBS donde se colocan las fuentes
- **Scheduler:** El planificador que evalúa qué mostrar
- **Server Schedule:** Schedule oficial activo en el backend
- **Staging:** Preparación segura antes del cambio visible
- **VirtualCam:** Cámara virtual de OBS para output de video
- **WebSocket:** Protocolo para comunicación en tiempo real
- **WHEP:** WebRTC-HTTP Egress Protocol para streaming de baja latencia
- **Working Schedule:** Copia local del schedule en el Editor que puede divergir del servidor

---

## 12. Contacto y Soporte

### Recursos de Ayuda

- **Documentación técnica:** Consulta las especificaciones técnicas completas
- **Logs de aplicación:** Revisa el archivo de log para detalles de errores
- **Comunidad OBS:** Para problemas específicos de OBS Studio

### Información de Versión

- **Versión actual:** Beta 0.1
- **Fecha de lanzamiento:** Octubre 2025

### Características de esta Versión

**Implementadas en Beta 0.1:**
- Sistema dual de vistas (Monitor/Editor)
- Triple sistema de indicadores de estado
- Previsualización WebRTC con protocolo WHEP
- Hot-reload automático de schedules
- Sistema de staging de 5 pasos
- Reconexión automática con sincronización de estado
- Registro de actividad en tiempo real
- Gestión completa de eventos recurrentes

**Limitaciones conocidas:**
- API REST limitada
- Sin templates de eventos
- Backup manual únicamente

### Próximas Características (Roadmap)

- Sistema de templates para eventos comunes
- API REST completa para integración externa
- Backup automático programado
- Estadísticas y analíticas de emisión
- Soporte para múltiples escenas simultáneas
- Editor visual de transforms
- Importación desde Google Calendar/iCal

---

**Scene Scheduler Beta 0.1 - Manual de Usuario**
© 2025 - Todos los derechos reservados
