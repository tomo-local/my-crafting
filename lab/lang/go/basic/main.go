package main

import "fmt"

func main() {
	john := Human{
		name: "John Done",
		age:  30,
		city: "Tokyo",
	}

	john.greet()
}

type Human struct {
	name string
	age  uint8
	city string
}

func (h *Human) greet() {
	fmt.Printf("I am %s, %d, and I am from %s\n\r", h.name, h.age, h.city)
}
