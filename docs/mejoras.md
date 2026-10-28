# Mejoras y correcciones a la aplicación

1. Cambiar los estilos del css para usar bootstrap 5, usando los siguientes colores para el tema:

```
primary: #26a269
secondary: #6c757d
success: #198754
info: #0dcaf0
warning: #ffc107
danger: #dc3545
light: #f8f9fa
dark: #212529
```

2. Cambiar el diseño del layout, establecer el contenido dentro de un `div.container` y mantener el diseño de tres filas
(formulario, teclado y estadisticas).

3. No me convence del todo el actual diseño del teclado y cree un diseño en svg (está en `dacs/keyboard.svg`), quiero cambiar el
teclado actual a este en svg, y a través de javascript resaltar cada tecla a presionar con un fondo en color '#bc8f8f' y cuando no
se está activa devolverle su color original. edite el svg para que se acceda a cada tecla a traves del id `#key<id>`, actualicé el
archivo `docs/layout.yaml` y a cada tecla agregué un campo `id` el cual contiene cada id asociado a cada tecla del svg para
buscar.
