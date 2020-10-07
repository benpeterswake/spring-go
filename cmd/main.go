package main

import (
	"gitlab.com/BPeters58/spring-go"
	_ "gitlab.com/BPeters58/spring-go/pkg/controller"
	_ "gitlab.com/BPeters58/spring-go/pkg/service"
)

func main() {
	spring.Run()
}
