package main

import (
	"fmt"
	"math/rand"
)

const LINE_BRAKE string = "\r\n"

func main() {
	study_control_flow()
}

// 03_control_flow.md
func study_control_flow() {
	if x := rand.Int(); x > 0 {
		fmt.Println("positive: ", x)
	} else {
		fmt.Println("negative: ", x)
	}

	// normal loop
	for i := 0; i < 10; i++ {
		fmt.Println("count: ", i)
	}

	// white loop
	n := 1
	for n < 100 {
		n *= 10
		fmt.Println("num: ", n)
	}

	// infinity loop
	b := 0
	for {
		b++
		if b == 10 {
			fmt.Println("break point: ", b)
			break
		}
	}

	if result, err := check_weekend("Saturday"); err == nil {
		fmt.Println("weekend: ", result)
	} else {
		fmt.Println(err)
	}

	if result, err := check_weekend("Monday"); err == nil {
		fmt.Println("weekend: ", result)
	} else {
		fmt.Println(err)
	}

	if result, err := check_weekend("Mondy"); err == nil {
		fmt.Println("weekend: ", result)
	} else {
		fmt.Println(err)
	}
}

func check_weekend(day string) (bool, error) {
	defer fmt.Println("input: ", day)

	switch day {
	case "Saturday", "Sunday":
		return true, nil
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		return false, nil
	default:
		return false, fmt.Errorf("not a day of the week.")
	}
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
