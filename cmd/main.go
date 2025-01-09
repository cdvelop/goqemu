package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cdvelop/goqemu"
)

func main() {
	// Configuración básica de la VM usando valores por defecto
	config := &goqemu.QemuConfig{
		Display: "sdl", // Usar SDL como interfaz gráfica
		VNCPort: 5900,  // Puerto VNC por defecto para Windows
	}

	// Crear directorio para imágenes si no existe
	imgDir := filepath.Join(os.Getenv("HOME"), "qemu", "img")
	err := os.MkdirAll(imgDir, 0755)
	if err != nil {
		log.Fatalf("Error creando directorio de imágenes: %v", err)
	}

	// Configurar handlers HTTP
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		startVMHandler(w, r, config)
	})
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/stop", stopHandler)

	// Iniciar servidor
	fmt.Println("Servidor HTTP iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// startVMHandler maneja el inicio de la VM
func startVMHandler(w http.ResponseWriter, r *http.Request, config *goqemu.QemuConfig) {
	// Crear nueva VM
	vm, err := goqemu.NewQemuVM(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creando VM: %v", err), http.StatusInternalServerError)
		return
	}

	// Iniciar VM
	err = vm.Start()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error iniciando VM: %v", err), http.StatusInternalServerError)
		return
	}

	// Iniciar VM con interfaz gráfica
	err = vm.StartWithGUI()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error iniciando interfaz gráfica: %v", err), http.StatusInternalServerError)
		return
	}

	// Responder con estado
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "running",
		"message": "VM iniciada correctamente",
	})
}

// statusHandler muestra el estado de la VM
func statusHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica para obtener estado real
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "unknown",
	})
}

// stopHandler detiene la VM
func stopHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica para detener VM
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "stopped",
		"message": "VM detenida",
	})
}
