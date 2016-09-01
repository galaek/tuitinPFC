package main
import (
	"os"
	"fmt"
)

func main() {
	
	var (
		cr *os.File
		err error
	)
	if cr, err = os.Create("CREDENTIALS"); err != nil {
		fmt.Println("Errro a crear CREDENTIALS")
		os.Exit(1)
	}	
	cr.Write([]byte("Cij69U0URS7d0dMNkw9p2nfJp")) 
	cr.Write([]byte("\n"))
	cr.Write([]byte("gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z")) 
	cr.Write([]byte("\n"))
	cr.Write([]byte("2425498608-LuXL0xwG2tpHJzzBJN4XTIP3W2RhzoMYukos57i"))
	cr.Write([]byte("\n"))	
	cr.Write([]byte("4UdIltWpL8S69k3gcYDFfGcjiqeIp2aUjca0UxALIM2BQ"))
	cr.Write([]byte("\n"))
	cr.Close()
	
}