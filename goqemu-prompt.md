# Objetivo
Crear un SDK en Go llamado 'goqemu' para interactuar programáticamente con el emulador QEMU, principalmente desde Windows pero compatible con Linux y macOS. El SDK debe proporcionar una API clara y tipada para gestionar máquinas virtuales, enfocándose en casos de uso de pruebas de configuración y despliegue.

# Características Principales
1. Gestión de Máquinas Virtuales:
   - Configuración por defecto: 4GB RAM, 2 cores CPU, 10GB almacenamiento
   - Soporte exclusivo para arquitectura x86_64
   - Formato de imagen qcow2
   - Red configurada automáticamente en el mismo rango que el host
   - Detección automática de IP de la VM (IPv4)

2. Gestión de Imágenes:
   - Descarga automática de Debian 12 por defecto
   - Soporte para URLs personalizadas
   - Sistema de caché local en `~/qemu/img`
   - Detección automática del sistema operativo basado en la URL

3. Sistema de Snapshots:
   - Almacenamiento en disco por defecto
   - Opción para snapshots en memoria
   - Formato de identificación: ID corto + descripciones en corchetes
   - Almacenamiento local en `~/qemu/snapshots`

4. Comunicación SSH:
   - Sistema de comandos asíncrono mediante channels
   - Estructura SshCommand para gestión de resultados y errores
   - Reconexión automática en caso de desconexión

# Estructura del Código

```go
type QemuConfig struct {
    RAM              int    // GB, default 4
    CPU              int    // cores, default 2
    DiskSize         int    // GB, default 10
    ImageURL         string // opcional, default debian 12
    SnapshotsInMemory bool
}

type QemuVM struct {
    config      *QemuConfig
    ip          string
    sshClient   *ssh.Client
    commandChan chan SshCommand
}

type SshCommand struct {
    Command string
    Result  string
    Err    error
}

type Snapshot struct {
    ID          string    // ID corto único
    Description string    // Formato: [desc1, desc2, desc3]
    CreatedAt   time.Time
}
```

# Métodos Requeridos

## Gestión de VM
```go
func NewQemuVM(config *QemuConfig) (*QemuVM, error)
func (vm *QemuVM) Start() error 
func (vm *QemuVM) Stop() error
```

## Comandos SSH
```go
func (vm *QemuVM) SendCommand(cmd string) SshCommand
```

## Gestión de Snapshots
```go
func (vm *QemuVM) CreateSnapshot(description string) (*Snapshot, error)
func (vm *QemuVM) RestoreSnapshot(id string) error
func (vm *QemuVM) ListSnapshots() ([]Snapshot, error)
func (vm *QemuVM) DeleteSnapshot(id string) error
```

## Gestión de Caché
```go
func (vm *QemuVM) ListCachedImages() ([]string, error)
func (vm *QemuVM) DeleteCachedImage(name string) error
```

# Caso de Uso Principal (Test)
Crear un test que demuestre:
1. Inicialización de VM con Debian 12
2. Creación de un archivo de texto mediante SSH
3. Creación de snapshot
4. Eliminación del archivo
5. Restauración del snapshot
6. Verificación de la existencia del archivo

# Requisitos de Implementación
1. Usar `errors` standard de Go para manejo de errores
2. Tests secuenciales sin framework adicional
3. Documentación clara con ejemplos de uso
4. Manejo automático de directorios de caché
5. Detección confiable de IP de VM
6. Logging claro de operaciones

# Consideraciones
1. La API debe ser simple y difícil de usar incorrectamente
2. Evitar configuraciones basadas en YAML
3. Automatizar la mayor cantidad de procesos posible
4. Priorizar la confiabilidad sobre la flexibilidad
5. Mantener las dependencias al mínimo necesario

# Notas Adicionales
- Incluir manejo de timeouts apropiados
- Proporcionar mensajes de error claros y útiles
- Documentar prerequisitos (versión de QEMU, etc.)
- Incluir ejemplos de uso común en la documentación
