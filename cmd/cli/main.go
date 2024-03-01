package main

import (
	"MinecraftServerController/cmd/cli/commands"
	"fmt"
)

func main() {
	fmt.Println("Kubecraft CLI for creating and managing Minecraft servers on Kubernetes")
	commands.Execute()
}
