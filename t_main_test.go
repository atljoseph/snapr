package main

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Tests Starting")
	// EnvFilePath = ".tests.env"
	exitCode := m.Run()
	log.Println("Tests Done")
	os.Exit(exitCode)
}
