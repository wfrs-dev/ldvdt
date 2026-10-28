# ldvdt — Rehacer la interfaz: Bootstrap 5, container y teclado SVG

Reemplaza la capa visual propia (tokens Kanso + teclado HTML/CSS) por Bootstrap 5 con la
paleta del usuario, un `div.container` de tres filas y el teclado dibujado en SVG.

## Specs

- **S21** — Bootstrap 5 + paleta. Pedido del usuario, verbatim: "Cambiar los estilos del css para usar bootstrap 5, usando los siguientes colores para el tema:" con el bloque `primary: #26a269`, `secondary: #6c757d`, `success: #198754`, `info: #0dcaf0`, `warning: #ffc107`, `danger: #dc3545`, `light: #f8f9fa`, `dark: #212529`. La paleta tiene que llegar a los componentes de Bootstrap (botones, focus ring, links), no sólo a las variables `--bs-*`.
- **S22** — Layout. Pedido del usuario, verbatim: "Cambiar el diseño del layout, establecer el contenido dentro de un `div.container` y mantener el diseño de tres filas (formulario, teclado y estadisticas)." Orden de las filas: 1 = formulario (+ toggle de tema), 2 = palabras + teclado, 3 = estadísticas + gráfico.
- **S23** — Teclado SVG. Pedido del usuario, verbatim: "No me convence del todo el actual diseño del teclado y cree un diseño en svg (está en `dacs/keyboard.svg`), quiero cambiar el teclado actual a este en svg, y a través de javascript resaltar cada tecla a presionar con un fondo en color '#bc8f8f' y cuando no se está activa devolverle su color original. edite el svg para que se acceda a cada tecla a traves del id `#key<id>`, actualicé el archivo `docs/layout.yaml` y a cada tecla agregué un campo `id` el cual contiene cada id asociado a cada tecla del svg para buscar." Los ids del yaml son la única fuente del mapeo tecla→elemento.
- **S24** — [Decisión] Se conserva el toggle manual de tema claro/oscuro (S20) y su persistencia en `localStorage`, pero el tema ahora se aplica con el atributo nativo de Bootstrap 5.3 `data-bs-theme` (`dark` por defecto, `light` al togglear) y la paleta de S21 rige en ambos temas.
- **S25** — [Decisión] El único resaltado del teclado es la tecla a presionar: fondo `#bc8f8f` mientras es la tecla actual y el relleno original del SVG al dejar de serlo. Se elimina el resaltado de las teclas habilitadas del nivel y, con él, el atributo `data-keys` y `highlightPracticeKeys`.
- **S26** — [Decisión] `docs/keyboard.svg` es la fuente de diseño: el binario embebe una copia generada (`internal/web/ui/keyboard.svg`) producida por `genlayout`, igual que `docs/layout.yaml` → `layout_gen.go`. La copia está versionada y CI falla si quedó desactualizada. El usuario corrige el SVG (ver T1).

## Tasks

- **T1** — Renombrar `id="rect83"` → `id="key7"` en `docs/keyboard.svg` (falta el id de la tecla `6`/`>`). [S23, S26] · usuario · n/a
- **T2** — ✅ Vendorizar Bootstrap 5.3.3 compilado desde Sass con la paleta de S21 + puente de tema `data-bs-theme` en `app.css`. [S21, S24] · inline · `ce98a20 feat: vendor themed Bootstrap 5 and the new palette`
- **T3** — ✅ Envolver el shell en `div.container` con las tres filas y componentes Bootstrap (form, stats, chart). [S22, S24] · inline · `6f38279 feat: wrap the SPA shell in a Bootstrap container`
- **T4** — ✅ Generador + dominio: campo `id` de `layout.yaml` → `Key.ID`, copia generada del SVG y componente `Keyboard` que embebe el SVG con su `data-keymap`. [S23, S26] · inline · `48d3d99 feat: render the keyboard from the SVG design`
- **T5** — ✅ Motor JS: resaltar la tecla actual por `#key<id>` con `#bc8f8f`, restaurar el relleno original y quitar el resaltado de nivel. [S23, S25] · inline · `88190e8 feat: highlight the SVG key to press`
- **T6** — ✅ Tests (Go + JS) y docs (`README.md`, CI de artefactos generados). [S21, S22, S23] · inline · `docs: document the Bootstrap shell and SVG keyboard`

## Log

- **L1** — 2026-10-28 · Pedido del usuario, verbatim (`docs/mejoras.md`):

  ```
  # Mejoras y correcciones a la aplicación

  1. Cambiar los estilos del css para usar bootstrap 5, usando los siguientes colores para el tema:

     primary: #26a269
     secondary: #6c757d
     success: #198754
     info: #0dcaf0
     warning: #ffc107
     danger: #dc3545
     light: #f8f9fa
     dark: #212529

  2. Cambiar el diseño del layout, establecer el contenido dentro de un `div.container` y mantener el diseño de tres filas (formulario, teclado y estadisticas).

  3. No me convence del todo el actual diseño del teclado y cree un diseño en svg (está en `dacs/keyboard.svg`), quiero cambiar el teclado actual a este en svg, y a través de javascript resaltar cada tecla a presionar con un fondo en color '#bc8f8f' y cuando no se está activa devolverle su color original. edite el svg para que se acceda a cada tecla a traves del id `#key<id>`, actualicé el archivo `docs/layout.yaml` y a cada tecla agregué un campo `id` el cual contiene cada id asociado a cada tecla del svg para buscar.
  ```

- **L2** — 2026-10-28 · Decisiones tomadas con el usuario antes de escribir código (tres preguntas cerradas):
  1. Tema (S24): "Conservar el toggle, pero con la paleta nueva" — el botón sigue alternando oscuro/claro, ahora con `data-bs-theme` de Bootstrap 5.3; sobreviven `test/theme_test.js` y S20.
  2. Tecla faltante en el SVG (S26/T1): "Renombrá rect83 a key7 vos en el SVG" — el usuario toca `docs/keyboard.svg`; el agente no lo edita. Detectado en la exploración: `layout.yaml` asigna `id: 7` a la tecla `6`/`>` y en el SVG ese rectángulo quedó con `id="rect83"` (mismo `x=124.41061`, `y=53.097702`, tamaño y relleno `#ffabff` que sus vecinos `key6`/`key8`).
  3. Resaltado (S25): "Solo #bc8f8f para la tecla actual" — se borra el resaltado de teclas del nivel.
- **L3** — 2026-10-28 · Decisiones técnicas de implementación (no consultadas, dentro de lo autorizado):
  - Bootstrap se vendoriza **compilado desde Sass** (`docs/bootstrap-theme.scss`) y no se parchea el CSS oficial con overrides: en 5.3.3 los componentes (`--bs-btn-bg`, el `box-shadow` del focus de `.form-control`, etc.) llevan los colores literales, así que overridear `--bs-primary` deja botones y focus ring azules. La única diferencia real con los defaults de Bootstrap es `primary`, el resto del bloque del usuario ya son los valores por defecto.
  - No se vendoriza JS de Bootstrap: `data-bs-theme` son reglas CSS y no se usan componentes que requieran JS (dropdown/modal/tooltip/collapse).
  - El SVG se embebe **inline** en el HTML (no `<img>` ni `/static/keyboard.svg`): el motor JS tiene que mutar el relleno de los rects, y un documento externo queda fuera del DOM.
  - El mapeo carácter→elemento viaja como JSON en `data-keymap` del wrapper `#keyboard`, generado en Go desde `domain.Keyboard` + `Key.ID`; la barra espaciadora no está en `layout.yaml` y se mapea al id `space` del SVG.
- **L4** — 2026-10-28 · Corte de slice al cerrar T2: el commit `ce98a20` mide 461 líneas "autoradas" según el contador de la sesión, pero ~228 KB son el artefacto vendorizado `internal/web/static/vendor/bootstrap.min.css` (compilado, no escrito a mano: 13 líneas de salida comprimida de Sass). La superficie real de revisión es `docs/bootstrap-theme.scss` (23 líneas), `app.css`, `app.js`, `page.templ` y los tests. No se parte el commit porque el binario no compila sin el artefacto: Bootstrap y su tema son una unidad de entrega.
- **L5** — 2026-10-28 · T4 y T5 entregadas. Decisiones de implementación:
  - El generador (`genlayout`) valida que todo `id` sea positivo y único: los ids son el puente al SVG, y un id faltante o repetido rompería el resaltado en silencio. No valida que el SVG dibuje cada id (ver T1: falta `#key7`); el motor degrada a "sin resaltado" cuando el elemento no existe, y eso está pineado en `test/keyboard_test.js`.
  - `genlayout` también copia `docs/keyboard.svg` → `internal/web/ui/keyboard.svg` quitando el prólogo `<?xml ?>` (inválido dentro de un documento HTML). La copia va versionada y CI la verifica, igual que `layout_gen.go`.
  - El mapeo carácter→elemento viaja como `data-keymap` en `#keyboard` (JSON generado en Go) y no como atributos por tecla: el SVG es un diseño crudo, no markup nuestro.
  - El fill del resaltado se lee de `--ldvdt-key-current` en `app.css` con fallback `#bc8f8f`, para no duplicar el color de la paleta en JS (mismo patrón que el chart).
  - `domain.Key.Finger()` queda sin consumidor de producción (el color por dedo ahora está pintado dentro del SVG). Se conserva a propósito: es el modelo que documenta la asignación de dedos y sigue cubierto por sus tests. `keyClass` (helper de UI que ya no renderiza nada) se eliminó.
  - Slice de T5 alineado con T4: `test/keyboard_test.js` es la suite nueva; se verificó RED contra el `app.js` de `HEAD` (`app.js must expose highlightKey`) antes de dejar GREEN.
