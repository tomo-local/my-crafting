package main

import "fmt"

const LINE_BRAKE string = "\r\n"

func main() {
	study_functions()
}

// 02_functions.md
func study_functions() {
	a := 1
	b := 2

	fmt.Printf("add: 1+2=%d%s", add(a, b), LINE_BRAKE)

	if result, err := divide(10, 5); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("divide: 10/5=%f%s", result, LINE_BRAKE)
	}

	if result, err := divide(100, 0); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("divide: 100/0=%f%s", result, LINE_BRAKE)
	}
}

func add(a, b int) int {
	return a + b
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return a / b, nil
}

// 01_types.md

func study_types() {
	age := 30
	john := &Human{
		name: "John Done",
		age:  &age,
		city: "Tokyo",
	}

	// & is address
	fmt.Printf("address: %p%s", john, LINE_BRAKE)
	// * is values
	fmt.Printf("values: %+v%s", *john, LINE_BRAKE)

	john.Greet()

	sam := Human{
		name: "Samuel",
		city: "U.S",
	}

	fmt.Printf("values: %+v%s", sam, LINE_BRAKE)
	fmt.Printf("address: %p%s", &sam, LINE_BRAKE)

	sam.Greet()
}

type Human struct {
	name string
	// 未入力したい時は、*ポインターすることで、nilと0の場合を区別することができる
	// If you wish to indicate no input, you can hover over the field to distinguish between nil and 0.
	age  *int
	city string
}

func (h *Human) Greet() {
	if h.age != nil {
		fmt.Printf("I am %s, %d, and I am from %s.%s", h.name, h.age, h.city, LINE_BRAKE)
	}
	fmt.Printf("I am %s, and I am from %s.%s", h.name, h.city, LINE_BRAKE)
}
