# Principios

Reglas que guían todas las decisiones de diseño.

## P1. Especifica lo que sabes, cuando lo sabes

Lo que no se sabe queda como **pregunta abierta**, visible, nunca inventado. Es el antídoto contra la cascada de supuestos y contra que el agente llene huecos. Se aplica en el arranque de un repo existente y en el flujo de trabajo (`method`).

## P2. El modelo propone, el código decide

Todo lo que tiene respuesta correcta lo decide el binario: si un check pasó, qué está listo, si un cambio debilita una promesa, qué contexto aplica a un archivo, qué está desactualizado. Las buenas decisiones humanas no sobreviven al contacto con un agente a menos que algo las haga cumplir.

## P3. El proceso escala con la tarea

El trabajo pequeño no deja registro ni pide nada. Solo el trabajo que no se puede reanudar desde la petición y el diff recibe un registro de trabajo. **Nunca se decide por número de archivos** (`method`).

## P4. Grande en efecto, chico en superficie

Reglas medibles de simplicidad:

1. **Cero comandos que aprender para el usuario.** Se habla con el agente.
2. **Valor en la primera sesión.** El arranque en un repo existente da contexto útil en minutos.
3. **Ninguna ceremonia en el trabajo pequeño.**
4. **Una sola herramienta para el agente:** el CLI `beacon`, sin servidor MCP.
5. **Una pregunta por decisión importante.** Nunca un formulario.
6. **Una sola dependencia:** el binario. Sin Node, Python, Docker ni servidores.

## P5. Vigente, no histórico

El estado describe lo que es verdad **ahora**. Lo viejo no se borra: se **cierra** con fecha y motivo (lo reemplazó X) y pasa a `archive/`. El historial completo lo da git.

## P6. Nada importante cambia sin la persona; todo queda a la vista

Los cambios se clasifican en tres niveles según el riesgo de debilitar lo prometido. Los de nivel 3 piden un sí o no, y quien los detecta es el binario, no el modelo. Detalle en `deltas`.

## P7. Los mensajes llevan a una salida

Si Beacon bloquea algo, dice exactamente qué hacer, y lo que dice funciona. Un mensaje solo nombra un comando si ejecutarlo resuelve el bloqueo. Cero callejones sin salida; se mide (`engine`).

## P8. Apuntar el instrumento a uno mismo

Beacon se desarrolla con Beacon desde el primer momento posible, y tiene un benchmark de fricción propio (`status`).

## P9. Spider es estructura activa

La memoria no es un archivo al que se le pregunta: es una estructura que empuja al agente. Una pregunta abierta bloquea, una promesa en rojo avisa, una relación tensa a sus vecinos (`spider`).