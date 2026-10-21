# ldvdt — Entrenador de layout de teclado personalizado

Entrenador web (SPA) para practicar el layout de teclado **ldvd** a través de niveles.
Go + hefesto (Echo v5) + arquitectura hexagonal + htmx/templ SSR + SQLite + chart.js.

## Specs

- **S1** — Objetivo: "Crear una aplicion que actue como un pequeño servidor web que muestre una SPA para poder practicar el uso de este layout a traves de niveles para que el usuario pueda ir progresando en el uso de este layout."
- **S2** — Base: Go + librería `hefesto` (https://github.com/wfrs-dev/hefesto, extensión de Echo V5), arquitectura hexagonal, frontend con `htmx` vía SSR usando `a-h/templ`.
- **S3** — SQLite para el progreso, guardada en el directorio `.config` del espacio de trabajo del usuario.
- **S4** — Gráficos de avance con `chart.js`.
- **S5** — Todos los assets cargados por el binario Go vía `embed`.
- **S6** — Debe funcionar en Windows y Linux.
- **S7** — "El binario se encargara de iniciar el servidor web y despues lanzar el navegador web por defecto en el puerto 10987".
- **S8** — SPA responsive a pantalla completa: derecha = listado (nivel, idioma es/en, cantidad de palabras mín 20); izquierda = progreso; centro = texto con las palabras + imagen del teclado con las teclas a presionar.
- **S9** — Teclado HTML+CSS diseñado desde `docs/layout.yaml` (ISO), generado fijo sin cargar el yaml en cada inicio; colores distintos por dedo.
- **S10** — Palabras desde `resources/eng.dat` y `resources/spa.dat`; si faltan palabras en el nivel, texto al azar de 5 a 10 letras con las teclas a practicar.
- **S11** — Guardar progreso: WPM, accuracy, consistencia, duración por práctica y por idioma; global: cantidad de prácticas realizadas y tiempo total.
- **S12** — Niveles (acumulativos):
  - N1: fila media minúsculas (a, o, e, i, u, d, r, t, n, s).
  - N2: fila media mayúsculas (A..S) + N1.
  - N3: N1+N2 + fila alta minúsculas (ñ, ,, ., p, y, f, g, c, h, l) + acentuadas (á, é, í, ó, ú).
  - N4: N1..N3 + fila alta mayúsculas (Ñ, ,, ., P..L) + acentuadas (Á, É, Í, Ó, Ú).
  - N5: N1..N4 + fila baja minúsculas (*, -, q, k, x, b, m, w, v, z).
  - N6: N1..N5 + fila baja mayúsculas (*, -, Q..Z).
  - N7: N1..N6 + fila numerada mayúsculas (1..0, ?, =).
  - N8: N1..N7 + fila numerada minúsculas (&, [, {, (, <, >, ), }, ], /, `, +).
- **S13** — Semántica `top`/`bottom`: `top` = tecla con SHIFT (fila alta), `bottom` = tecla normal (fila baja). Corrección acordada: fila `middle` debe ser `bottom`=minúscula (normal) y `top`=mayúscula.
- **S14** — Al escribir, resaltar la tecla actual en el teclado de ayuda; caracteres acentuados muestran la tecla + la tecla de acento (`´`).
- **S15** — Fila numerada: generar palabras al azar con números y símbolos, además de teclas alfabéticas.
- **S16** — [Decisión 2026-10-20] Filtrar las líneas numéricas sueltas de `eng.dat`/`spa.dat`; el corpus queda solo con palabras alfabéticas.
- **S17** — [Decisión 2026-10-20] Consistencia = regularidad del ritmo por tecla (varianza del intervalo entre pulsaciones dentro de una misma práctica).
- **S18** — [Decisión 2026-10-20] Empezar con slice vertical (servidor + teclado + nivel 1 + SQLite + práctica completa), luego el resto.

## Tasks

- **T1** — Corregir `docs/layout.yaml` (fila `middle` → bottom=minúscula, top=mayúscula) y armar scaffold Go (deps: hefesto, echo v5, a-h/templ, modernc.org/sqlite). [S2, S13] · inline · `chore: scaffold ldvdt + fix home row layout`
- **T2** — Modelo de dominio: `Key`, `Row`, `Keyboard` (parseo del layout → structs Go embebidos/generados en build, no yaml en runtime) + definición de los 8 niveles con caracteres permitidos. [S9, S12, S13] · inline · `feat: keyboard + level domain model`
- **T3** — Cargador de corpus (eng/spa `.dat`, filtrar números) + selección de palabras por nivel con fallback de texto aleatorio 5-10 letras. [S10, S15, S16] · inline · `feat: corpus loader + word selection`
- **T4** — Almacenamiento SQLite (esquema + repos): práctica, progreso por idioma, stats globales. [S3, S11] · inline · `feat: sqlite progress storage`
- **T5** — Config + bootstrap del servidor (hefesto/echo, puerto 10987) + abrir navegador + embed de assets estáticos. [S5, S6, S7] · inline · `feat: server bootstrap + browser launch`
- **T6** — Shell SPA en templ (3 paneles) + componente de teclado ISO con colores por dedo + wiring htmx. [S8, S9] · inline · `feat: templ SPA shell + keyboard`
- **T7** — JS cliente: captura de teclas, timing, WPM/accuracy/consistencia, resaltado de tecla actual, envío de resultados. [S14, S17] · inline · `feat: typing engine client-side`
- **T8** — Handlers + casos de uso: niveles/palabras, iniciar práctica, enviar práctica, progreso; conectado a SQLite. [S11] · inline · `feat: practice handlers + use cases`
- **T9** — Prueba E2E nivel 1 (RED→GREEN): práctica completa con fila media, guardado y progreso. [S1, S12] · inline · `test: level 1 end-to-end`
- **T10** — Niveles 2–6 (mayúsculas fila media, fila alta ± acentos, fila baja). [S12] · inline · `feat: levels 2-6`
- **T11** — Niveles 7–8 (fila numerada). [S12, S15] · inline · `feat: levels 7-8`
- **T12** — Gráficos de avance con chart.js. [S4] · inline · `feat: progress charts`
- **T13** — Manejo de tecla muerta para acentuadas en teclado + tipeo. [S14] · inline · `feat: dead key accents`
- **T14** — Apertura de navegador cross-platform + verificación de build Windows. [S6, S7] · inline · `feat: cross-platform browser + windows build`
- **T15** — Suite completa + docs (README/help). [S1] · inline · `docs: readme + help`

## Log

- **L1** — 2026-10-20 · Pedido original: "Lo que quiero hacer es una aplicación en Go que en resumen es un entrenador para un layout de teclado personalizado que realicé yo, tengo la información de lo que quiero en el archivo @docs/idea.md, léelo para empezar a implementarlo y preguntame si hay algo que aclarar".
- **L2** — 2026-10-20 · Respuestas del usuario: (1) layout.yaml fila middle es error → corregir (bottom=minúscula, top=mayúscula); (2) filtrar números de los .dat; (3) consistencia = regularidad del ritmo por tecla; (4) empezar con slice vertical.
- **L5** — 2026-10-20 · Ajustes visuales pedidos por el usuario: (a) el espacio ahora resalta la tecla espacio del teclado (`data-bottom=" "` en la tecla espacio, mapeada por el motor de tipeo); (b) teclado con teclas decorativas ISO — tab, bloq mayús, shift izq/der, backspace y enter — en color neutro, sin data attributes para no interferir con el resaltado. `.words-row` pasó a flex+wrap con gap para separar palabras y ajustar líneas. Verificado por test de render (48 teclas de práctica + `data-bottom=" "` + las 6 teclas decorativas).
- **L4** — 2026-10-20 · Slice vertical completo (T6–T9): shell SPA templ + teclado ISO con colores por dedo, CSS, motor de tipeo JS (WPM/accuracy/consistencia, resaltado de tecla), handlers + casos de uso, y prueba E2E nivel 1. Verificado end-to-end con el binario real (GET /, GET /practice, POST /practice → stats).
- **L3** — 2026-10-20 · Progreso T1–T5 completado en rama `feat/ldvdt-trainer`. Commits: `a2cc5a3` (chore scaffold + fix yaml), `37a697d` (dominio keyboard/niveles + corpus), `bb23d1a` (sqlite + config + server). Nota de slice: `resources/*.dat` son datos (6233 líneas), no código autoral; el código Go real ronda las ~900 líneas. Supuestos yaml-derivados registrados: (a) nivel 4 usa `;`/`:` (shift de `,`/`.`) en vez de `,`/`.`; (b) niveles 5/6 incluyen `j`/`J`; (c) niveles 7/8 incluyen todas las 13 teclas numéricas (`~`/`$` incluidas); (d) BD en `os.UserConfigDir()/ldvdt/ldvdt.db`; (e) servidor bindea a `:10987` (todas las interfaces) por diseño de hefesto.
