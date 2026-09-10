# Explicación: falla del test de perfil de memoria y su solución

## Contexto

El test `MemoryProfile` (`tests/memory_profile.py`) mide el pico de memoria del contenedor
del cliente (`/sys/fs/cgroup/memory.peak`) procesando dos archivos de entrada: uno mediano
(1.000 apuestas, ~30KB) y uno grande (1.000.000 de apuestas, ~31MB). El test exige que la
diferencia entre ambos picos no supere los 8MB.

Al ejecutarlo, el test fallaba con:

```
Difference in memory profiles is too big: 15368192B vs 47308800B
```

diferencia de ~32MB, lineal con el tamaño del archivo de entrada.

## Diagnóstico

El cliente ya procesa las apuestas en lotes de a `BATCH_SIZE` (32), por lo que su heap no
crece con el dataset. Para confirmar dónde estaba la memoria, se inspeccionó el desglose del
cgroup del contenedor (`/sys/fs/cgroup/memory.stat`) mientras procesaba el archivo grande:

```
anon 4247552      ← heap del proceso: ~4MB, plano
file 30892032     ← page cache cargada al cgroup: ~31MB
kernel 1540096
slab 654936
```

Es decir: **la diferencia no era memoria del proceso sino page cache del kernel atribuida al
cgroup del contenedor**.

### Cadena causal

1. En cgroup v2 (el modo por defecto en las distribuciones actuales), `memory.peak` incluye
   toda la memoria atribuida al contenedor: heap, estructuras del kernel **y la page cache
   de los archivos que el contenedor lee**.
2. Al leer el archivo de entrada de 31MB, el kernel cachea sus páginas a nombre del cgroup
   del contenedor. El heap del cliente se mantiene plano, pero la memoria atribuida al
   contenedor crece junto con el archivo.
3. `memory.peak` es un **máximo histórico**: aunque la cache se libere después, el pico ya
   quedó registrado.

El kernel cachea por defecto porque es el comportamiento correcto en general (un archivo
leído suele releerse); el problema es que este cliente lee el archivo **una única vez, de
forma secuencial**, y nunca más lo necesita.

## Solución

Se modificó únicamente `services/client/src/repository/bet_reader.go`.

El lector acumula los bytes consumidos del archivo y, cada 1MiB, aplica
`posix_fadvise(DONTNEED)` sobre el rango ya leído mediante `syscall.Syscall6` con
`SYS_FADVISE64` (sin agregar dependencias externas):

```go
func (r *betRepository) trackRead(n int) {
	r.readBytes += n
	if r.readBytes < r.dropCacheUpTo {
		return
	}
	r.dropCacheUpTo = r.readBytes + dropCacheChunkBytes
	syscall.Syscall6(syscall.SYS_FADVISE64, r.file.Fd(), 0, uintptr(r.readBytes), posixFadvDontNeed, 0, 0)
}
```

`DONTNEED` le indica al kernel que las páginas limpias del rango ya no se necesitarán; el
kernel las evictiona y libera la carga del cgroup **en ese momento**. Así, la page cache viva
queda acotada a ~1MB sin importar el tamaño del archivo, y el pico de memoria del contenedor
deja de depender del dataset.

Dos detalles de la implementación:

- Se libera **durante** la lectura, no al final: si se liberara al terminar, el pico ya
  habría quedado registrado.
- La syscall es *best effort*: su resultado se ignora, de modo que en un entorno donde
  `fadvise` no esté disponible el comportamiento funcional del cliente no cambia (solo
  volvería a cargarse la cache).

## Resultado

- Test de perfil de memoria: **OK** (diferencia entre perfiles bien por debajo de los 8MB).
- Suite completa: los 9 tests pasan, incluidos los de comportamiento del protocolo
  (batching, short read/write, SIGTERM, etc.), ya que el cambio no altera la lógica de
  comunicación.

## Por qué esta solución y no otra

| Alternativa | Por qué no |
|---|---|
| Liberar la cache al terminar de leer | `memory.peak` ya registró el máximo; no reduce el pico |
| Leer con `O_DIRECT` | Requiere alineación de buffers; complejo y frágil con `bufio` |
| `mmap` + `madvise` | Mismo problema de pico; agrega complejidad sin beneficio |
| Aumentar el umbral / tocar el test o el compose | El enunciado pide un cliente de perfil de memoria plano; el fix es de código |

`posix_fadvise` es la API estándar precisamente para declarar el patrón de acceso a
archivos (la usan bases de datos y pipelines de datos); comunica al kernel que el patrón
de este programa es de lectura secuencial de un solo uso.
