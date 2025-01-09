package goqemu

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const minQemuVersion = "8.1.0" // Versión mínima requerida de QEMU

// checkQemuInstalled verifica si QEMU está instalado en el sistema
func checkQemuInstalled() error {
	_, err := exec.LookPath("qemu-system-x86_64")
	if err != nil {
		return errors.New("QEMU no está instalado. Por favor instale QEMU antes de continuar")
	}

	if err := checkQemuVersion(); err != nil {
		return err
	}

	return nil
}

// checkQemuVersion verifica si la versión instalada cumple con los requisitos
func checkQemuVersion() error {
	version, err := getQemuVersion()
	if err != nil {
		return err
	}

	if compareVersions(version, minQemuVersion) < 0 {
		return fmt.Errorf("versión de QEMU (%s) es menor que la requerida (%s)", version, minQemuVersion)
	}

	return nil
}

// getQemuVersion obtiene la versión instalada de QEMU
func getQemuVersion() (string, error) {
	out, err := exec.Command("qemu-system-x86_64", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo versión de QEMU: %v", err)
	}

	// Extraer versión usando regex
	re := regexp.MustCompile(`QEMU emulator version (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return "", errors.New("no se pudo determinar la versión de QEMU")
	}

	return matches[1], nil
}

// compareVersions compara dos versiones en formato semver
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		if v1Parts[i] < v2Parts[i] {
			return -1
		}
		if v1Parts[i] > v2Parts[i] {
			return 1
		}
	}

	return 0
}
