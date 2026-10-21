# Entrenador de layout de teclado personalizado `ldvd`

**ldvd** es un layout de teclado personalizado basado en los layouts dvorak latinoamericano y dvorak para programadores.

## Objetivo

Crear una aplicion que actue como un pequeño servidor web que muestre una SPA para poder practicar el uso de este layout a traves
de niveles para que el usuario pueda ir progresando en el uso de este layout.

## Tecnologías

El desarrollo base de la aplicacion va a ser creada en go, usando la libreria `hefesto` ubicado en el repositorio de github
<https://github.com/wfrs-dev/hefesto>, el cual es una extensión del framework Echo V5, el desarrollo va a ser con arquitectura
hexagonal y para la parte frontend se va a incluir `htmx` a traves de Server Side Rendering y usando la libreria `a-h/templ`.

Va a contener una base de datos sqlite en la cual se va a almacenar el progreso del usuario en el uso del layout, esta base de
datos se va a guardar en el directorio `.config` en el espacio de trabajo del usuario.

La aplicacion va a mostrar graficos de avance del progreso del usuario, estos graficos van a usar la libreria `chart.js`

Todos los assets los cargara el binarrio de go a traves de embed.

La aplicación por el momento debe funcionar en Windows y Linux

## Interfaz de usuario

El binario se encargara de iniciar el servidor web y despues lanzar el navegador web por defecto en el puerto 10987

La interfaz es una SPA responsive que a pantalla completa, al lado derecho se muestre un listado en cual se pueda
seleccionar el nivel que se desea practicar, el idioma a practicar (español o inglés) y la cantidad de palabras a practicar (como
mínimo 20), al lado izquierdo se muestra el progreso del usuario en el uso del layout, en el centro se muestra el texto con las
palabras que se deben escribir para practicar el layout y una imagen del teclado que muestra las teclas que se deben presionar
para practicar el layout.

El teclado se va a diseñar usando HTML y CSS, para el diseño se va a usar el archivo `layout.yaml` que se encuentra en el
directorio `docs` del repositorio, en este archivo se definen la estructura del layout, conforme se seleccione un nivel, muestre
las teclas a practicar y con colores diferentes muestren los dedos que se deben presionar para practicar el layout. El teclado a
diseñar es un teclado ISO. En el archivo yaml se definen la ubicacion de las teclas, se debe generar el teclado fijo sin necesidad
de cargar el yaml en cada inicio de la aplicación.

En `resources/eng.dat` y `resources/spa.dat` se encuentran las palabras a practicar en ingles y español respectivamente,
dependiendo del nivel seleccionado, pueden ocurrir que falten palabras en el nivel seleccionado, en este caso se muestra un texto
de 5 a 10 letras al azar con las teclas que se deben presionar para practicar el layout.

Guardar en la base de datos el progreso del usuario en el uso del layout, su WPM, su accuracy, su consistencia, el tiempo de
duración por practica y por idioma; a nivel global guardar la cantidad de practicas realizadas, el tiempo de practica total.

## Niveles

Los niveles son los diferentes niveles de dificultad del layout, los niveles son:

- **Nivel 1:** Se practican las teclas de la fila media (middle) en minúsculas (a, o, e, i, u, d, r, t, n, s). Con palabras que
    contengan estas letras.
- **Nivel 2:** Se practican las teclas de la fila media (middle) en mayúsculas (A, O, E, I, U, D, R, T, N, S) e incluyendo las
    teclas del nivel 1.
- **Nivel 3:** Se practican las teclas de los niveles 1 y 2 junto con las teclas de la fila alta (upper) en minúsculas (ñ, ,, ., p,
  y, f, g, c, h, l) y las teclas acentuadas (á, é, í, ó, ú).
- **Nivel 4:** Se practican las teclas de los niveles 1, 2 y 3 junto con las teclas de la fila alta (upper) en mayúsculas (Ñ, ,, ., P,
  Y, F, G, C, H, L) y las teclas acentuadas (Á, É, Í, Ó, Ú).
- **Nivel 5:** Se practican las teclas de los niveles 1, 2, 3 y 4 junto con las teclas de la fila baja (bottom) en minúsculas (*, -, q, k, x, b, m, w, v, z).
- **Nivel 6:** Se practican las teclas de los niveles 1, 2, 3, 4 y 5 junto con las teclas de la fila baja (bottom) en mayúsculas (*, -, Q, K, X, B, M, W, V, Z).
- **Nivel 7:** Se practican las teclas de los niveles 1, 2, 3, 4, 5 y 6 junto con las teclas de la fila numerada (numeric) en mayúsculas (1, 2, 3, 4, 5, 6, 7, 8, 9, 0, ?, =).
- **Nivel 8:** Se practican las teclas de los niveles 1, 2, 3, 4, 5, 6 y 7 junto con las teclas de la fila numerada (numeric) en minúsculas (&, [, {, (, <, >, ), }, ], /, `, +).

En el archivo yaml cada tecla tiene dos campos, el campo `top` es donde se encuentra la tecla en la fila alta (cuando se usa SHIFT) y el campo `bottom` es
donde se encuentra la tecla en la fila baja (cuando se presiona la tecla normalmente).

Cuando se esté escribiendo en la tecla actual se resalte las teclas del teclado en la pantalla de ayuda indicando que teclas se
deben presionar, para los caracteres acentuados se debe mostrar la tecla y la tecla de acento (`´`).

Con la fila numerica(numeric), como en esta se icluyen los numeros y caracteres no alfanumericos, se deben generar palabras al azar con las
teclas de numeros y simbolos, ademas de incluir teclas alfabeticas.
