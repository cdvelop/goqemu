package goqemu

import (
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshCommand struct {
	Command string
	Result  string
	Err     error
}

// isPortAvailable verifica si un puerto está disponible
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// connectSSH establece la conexión SSH con la VM
func (vm *QemuVM) connectSSH() error {
	if vm.ip == "" {
		return errors.New("IP de la VM no disponible")
	}

	// Configuración básica del cliente SSH
	config := &ssh.ClientConfig{
		User: "user", // TODO: Hacer configurable
		Auth: []ssh.AuthMethod{
			ssh.Password("password"), // TODO: Implementar autenticación por clave
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Mejorar seguridad
		Timeout:         10 * time.Second,
	}

	// Establecer conexión
	client, err := ssh.Dial("tcp", vm.ip+":22", config)
	if err != nil {
		return err
	}

	vm.sshClient = client
	return nil
}

func (vm *QemuVM) SendCommand(cmd string) SshCommand {
	// Implementación de la función SendCommand
	return SshCommand{
		Command: cmd,
		Result:  "",
		Err:     nil,
	}
}

// waitForSSH espera hasta que el servicio SSH esté disponible
func waitForSSH(host string, port int, timeout int) error {
	start := time.Now()
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}

		if time.Since(start) > time.Duration(timeout)*time.Second {
			return fmt.Errorf("timeout esperando SSH")
		}

		time.Sleep(1 * time.Second)
	}
}
