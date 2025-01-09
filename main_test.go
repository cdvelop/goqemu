package goqemu

import (
	"testing"
	"time"
)

func TestVentanaGrafica(t *testing.T) {
	// Crear VM con configuración gráfica
	vm, err := NewQemuVM()
	if err != nil {
		t.Fatalf("Error creando VM: %v", err)
	}

	// Iniciar ventana gráfica
	err = vm.OpenWindow()
	if err != nil {
		t.Fatalf("Error abriendo ventana gráfica: %v", err)
	}
	defer vm.Stop()

	// Esperar un momento para que la ventana se abra
	time.Sleep(5 * time.Second)

	// Verificar que el proceso está corriendo
	if vm.process == nil {
		t.Error("El proceso de QEMU no se inició correctamente")
	}

	// Verificar conectividad con ping
	err = vm.Ping()
	if err != nil {
		t.Fatalf("Error en ping: %v", err)
	}

	t.Logf("IP de la máquina virtual: %s", vm.ip)

}

func TestMainScenario(t *testing.T) {
	// vm, err := NewQemuVM()
	// if err != nil {
	// 	t.Fatalf("Error creando VM: %v", err)
	// }

	// // 2. Iniciar VM
	// err = vm.Start()
	// if err != nil {
	// 	t.Fatalf("Error iniciando VM: %v", err)
	// }
	// defer vm.Stop()

	// // Esperar a que la VM esté lista
	// time.Sleep(30 * time.Second)

	// // 3. Crear archivo de texto mediante SSH
	// cmd := vm.SendCommand("echo 'Hola mundo' > test.txt")
	// if cmd.Err != nil {
	// 	t.Fatalf("Error creando archivo: %v", cmd.Err)
	// }

	// // 4. Crear snapshot
	// snapshot, err := vm.CreateSnapshot("[test inicial]")
	// if err != nil {
	// 	t.Fatalf("Error creando snapshot: %v", err)
	// }

	// // 5. Eliminar archivo
	// cmd = vm.SendCommand("rm test.txt")
	// if cmd.Err != nil {
	// 	t.Fatalf("Error eliminando archivo: %v", cmd.Err)
	// }

	// // 6. Restaurar snapshot
	// err = vm.RestoreSnapshot(snapshot.ID)
	// if err != nil {
	// 	t.Fatalf("Error restaurando snapshot: %v", err)
	// }

	// // 7. Verificar existencia del archivo
	// cmd = vm.SendCommand("cat test.txt")
	// if cmd.Err != nil {
	// 	t.Fatalf("Error verificando archivo: %v", cmd.Err)
	// }

	// if cmd.Result != "Hola mundo\n" {
	// 	t.Errorf("Contenido del archivo incorrecto. Esperado: 'Hola mundo\\n', Obtenido: '%s'", cmd.Result)
	// }
}
