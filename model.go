package goqemu

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

type vmDisplay string

const (
	DisplayNone vmDisplay = "none"
	DisplayGTK  vmDisplay = "gtk"
	DisplaySDL  vmDisplay = "sdl"
	DisplayVNC  vmDisplay = "vnc"
)

// QemuConfig define la configuración básica de una VM
type QemuConfig struct {
	RAM               int    // GB, default 4
	CPU               int    // cores, default 2
	DiskSize          int    // GB, default 10
	ImageURL          string // opcional, default debian 12
	SnapshotsInMemory bool
	Display           vmDisplay // "none", "gtk", "sdl", "vnc"
	VNCPort           int       // Puerto VNC si Display = "vnc"
}

// QemuVM representa una instancia de máquina virtual
type QemuVM struct {
	config      *QemuConfig
	ip          string
	sshClient   *ssh.Client
	commandChan chan SshCommand
	process     *os.Process
	defaultArgs []string // Argumentos de inicialización por defecto

	cmd *exec.Cmd

	ctx    context.Context
	cancel context.CancelFunc
}

// NewQemuVM crea una nueva instancia de QemuVM
func NewQemuVM(configs ...*QemuConfig) (*QemuVM, error) {
	var config *QemuConfig

	if err := checkQemuInstalled(); err != nil {
		return nil, err
	}

	// Si no se proporciona configuración o es nil, crear una por defecto
	if len(configs) == 0 || configs[0] == nil {
		config = &QemuConfig{
			RAM:      4,
			CPU:      2,
			DiskSize: 10,
			Display:  DisplayGTK,
			// Display:  DisplaySDL,
		}
	} else {
		config = configs[0]
		// Validar valores mínimos
		if config.RAM < 1 {
			return nil, errors.New("RAM debe ser al menos 1GB")
		}
		if config.CPU < 1 {
			return nil, errors.New("CPU debe ser al menos 1 core")
		}
		if config.DiskSize < 1 {
			return nil, errors.New("tamaño de disco debe ser al menos 1GB")
		}
		// Si no se especifica display, usar gtk por defecto
		if config.Display == "" {
			config.Display = DisplaySDL
		}
	}

	// Crear instancia con valores por defecto si es necesario
	if config.ImageURL == "" {
		config.ImageURL = "https://cloud.debian.org/images/cloud/bookworm/daily/latest/debian-12-nocloud-amd64-daily.qcow2"
	}

	// Crear imagen de disco si no existe
	imgPath := getImagePath(config.ImageURL)
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		err := downloadImage(config.ImageURL, imgPath)
		if err != nil {
			return nil, fmt.Errorf("error descargando imagen: %v", err)
		}
	}

	// Asignar IP a la VM
	ip, mask, err := assignVMIP()
	if err != nil {
		return nil, err
	}

	fmt.Printf("IP asignada: %s mascara %v \n", ip, mask)

	netConfig := fmt.Sprintf("user,id=net0,net=%s,dhcpstart=%s,hostfwd=tcp::2222-:22", mask, ip)

	defaultArgs := []string{
		"-m", fmt.Sprintf("%dG", config.RAM),
		"-smp", fmt.Sprintf("%d", config.CPU),
		"-hda", imgPath,
		"-netdev", netConfig,
		"-device", "e1000,netdev=net0",
	}

	// Configurar interfaz gráfica
	switch config.Display {
	case "gtk":
		defaultArgs = append(defaultArgs, "-display", "gtk")
	case "sdl":
		defaultArgs = append(defaultArgs, "-display", "sdl")
	case "vnc":
		if config.VNCPort == 0 {
			return nil, fmt.Errorf("se requiere un puerto VNC válido")
		}
		defaultArgs = append(defaultArgs,
			"-vnc", fmt.Sprintf(":%d", config.VNCPort-5900),
			"-display", "none")
	default:
		defaultArgs = append(defaultArgs, "-display", "none")
	}

	// Crear estructura QemuVM
	vm := &QemuVM{
		config:      config,
		ip:          ip,
		commandChan: make(chan SshCommand, 100), // Buffer de 100 comandos
		defaultArgs: defaultArgs,
	}

	// Create context with cancel
	vm.ctx, vm.cancel = context.WithCancel(context.Background())

	// Crear directorios necesarios
	err = createQemuDirs()
	if err != nil {
		return nil, fmt.Errorf("error creando directorios: %v", err)
	}

	return vm, nil
}

// Start inicia la máquina virtual
func (vm *QemuVM) Start() error {
	// Validar configuración mínima
	if vm.config == nil {
		return errors.New("configuración no inicializada")
	}

	// Verificar disponibilidad del puerto SSH
	if !isPortAvailable(2222) {
		return fmt.Errorf("el puerto 2222 ya está en uso")
	}

	args := make([]string, len(vm.defaultArgs))
	copy(args, vm.defaultArgs)

	// Solo daemonizar si no hay interfaz gráfica
	if vm.config.Display == "none" {
		args = append(args, "-daemonize")
	}

	cmd := exec.Command("qemu-system-x86_64", args...)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error iniciando QEMU: %v", err)
	}

	// Esperar a que el servicio SSH esté disponible
	err = waitForSSH(vm.ip, 2222, 30)
	if err != nil {
		return fmt.Errorf("error esperando SSH: %v", err)
	}

	// Establecer conexión SSH
	err = vm.connectSSH()
	if err != nil {
		return fmt.Errorf("error conectando SSH: %v", err)
	}

	return nil
}

// Stop detiene la máquina virtual
func (vm *QemuVM) Stop() error {

	// Cerrar conexión SSH si está abierta
	if vm.sshClient != nil {
		err := vm.sshClient.Close()
		if err != nil {
			return fmt.Errorf("error cerrando conexión SSH: %v", err)
		}
	}

	// Buscar y matar el proceso QEMU
	cmd := exec.Command("pkill", "qemu-system-x86_64")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error deteniendo QEMU: %v", err)
	}

	// Liberar recursos
	vm.config = nil
	vm.ip = ""
	close(vm.commandChan)

	vm.cancel() // Cancel context to stop goroutine

	return nil
}

// OpenWindow abre la ventana gráfica de QEMU
func (vm *QemuVM) OpenWindow() error {
	vm.cmd = exec.Command("qemu-system-x86_64", vm.defaultArgs...)

	// Create pipe for stderr
	stderr, err := vm.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the process
	err = vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("error iniciando QEMU: %v", err)
	}

	// Filter stderr in a goroutine with context
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-vm.ctx.Done():
				return
			default:
				line := scanner.Text()

				switch {
				case strings.Contains(line, "WARNING"):
					// Ignorar líneas con "WARNING"
					continue
				case strings.Contains(line, "pixbuf"):
					// ignorar pixbuf loaders or the mime database could not be found
					continue
				default:
					fmt.Fprintln(os.Stderr, line)
				}

			}
		}
	}()

	// Guardar el proceso para poder detenerlo luego
	vm.process = vm.cmd.Process

	return nil
}

// StartWithGUI inicia la VM con interfaz gráfica
func (vm *QemuVM) StartWithGUI() error {
	err := vm.Start()
	if err != nil {
		return err
	}

	return vm.OpenWindow()
}

// IsWindowOpen verifica si la ventana de QEMU ya está abierta
func (vm *QemuVM) IsWindowOpen() bool {
	if vm.process == nil {
		return false
	}

	// Verificar si el proceso sigue activo usando tasklist en Windows
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", vm.process.Pid))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Si el proceso está en la lista, está activo
	if strings.Contains(string(output), fmt.Sprintf("%d", vm.process.Pid)) {
		fmt.Println("La ventana de QEMU ya está abierta")
		return true
	}

	return false
}
