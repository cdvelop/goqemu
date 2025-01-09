package goqemu

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// getHostIP obtiene la dirección IP del host
func getHostIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("error obteniendo interfaces de red: %v", err)
	}

	for _, iface := range interfaces {
		// Ignorar interfaces loopback y no activas
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			// Usar solo IPv4
			if ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}

	return "", errors.New("no se pudo obtener una dirección IP válida")
}

// assignVMIP asigna una IP válida a la máquina virtual
// Ejemplo de salida: vmIP="192.168.1.100", netmask="192.168.1.0/24", err=nil
func assignVMIP() (string, string, error) {
	hostIP, err := getHostIP()
	if err != nil {
		return "", "", err
	}

	// Obtener la máscara de red
	ip := net.ParseIP(hostIP)
	if ip == nil {
		return "", "", errors.New("dirección IP inválida")
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", "", fmt.Errorf("error obteniendo interfaces: %v", err)
	}

	var netmask string
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.String() == hostIP {
					ones, _ := ipnet.Mask.Size()
					netmask = fmt.Sprintf("%s/%d", ipnet.IP.Mask(ipnet.Mask).String(), ones)
					break
				}
			}
		}
	}

	if netmask == "" {
		return "", "", errors.New("no se pudo determinar la máscara de red")
	}

	// Asignar la IP a la VM (último octeto +1)
	octets := strings.Split(hostIP, ".")
	if len(octets) != 4 {
		return "", "", errors.New("formato de IP inválido")
	}

	octets[3] = "100" // Asignar un valor fijo para pruebas
	vmIp := strings.Join(octets, ".")

	return vmIp, netmask, nil
}

func (vm QemuVM) Ping() error {

	// Ping verifica la conectividad con la VMfunc (vm *QemuVM) Ping() error {
	if vm.ip == "" {
		return errors.New("la VM no tiene una IP asignada")
	}

	// Ejecutar ping (Windows)
	cmd := exec.Command("ping", "-n", "1", vm.ip)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error ejecutando ping: %v", err)
	}

	// Verificar si hubo respuesta
	if !strings.Contains(string(output), "Respuesta desde") {
		return fmt.Errorf("no hubo respuesta desde %s", vm.ip)
	}

	return nil
}
