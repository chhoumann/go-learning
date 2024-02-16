package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("Hello, world!")
	fmt.Println("--- Int Division ---")

	var result, remainder, err = intDivision(10, 2)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%v/%v=%v (%v)\n", 10, 2, result, remainder)

	fmt.Println("--- Arrays ---")
	arrs()
	fmt.Println("--- Strings ---")
	strings()
	fmt.Println("--- Structs ---")
	structs()
	fmt.Println("--- Pointers ---")
	pointers()
	fmt.Println("--- Goroutines ---")
	goroutines()
	fmt.Println("--- Channels ---")
	channels()
	priceChecker()
	fmt.Println("--- Generics ---")
	generics()
}

func intDivision(numerator int, denominator int) (int, int, error) {
	var err error

	if denominator == 0 {
		err = errors.New("cannot divide by zero")
		return 0, 0, err
	}

	var result int = numerator / denominator
	var remainder int = numerator % denominator

	return result, remainder, nil
}

func arrs() {
	var intArr [3]int32
	intArr[1] = 123
	fmt.Println(intArr[0])
	fmt.Println(intArr[1:3]) // access element 1 and 2

	var intArr2 [3]int32 = [3]int32{1, 2, 3}
	fmt.Println(intArr2)

	intArr3 := [...]int32{1, 2, 3}
	fmt.Println(intArr3)

	// Slices wrap around arrays and are more gneral, powerful, and convenient
	var intSlice []int32 = []int32{4, 5, 6}
	fmt.Println(intSlice)
	// len(...) is byte length, not char length
	fmt.Printf("Length: %v | Capacity: %v\n", len(intSlice), cap(intSlice))
	intSlice = append(intSlice, 7)
	fmt.Printf("Length: %v | Capacity: %v\n", len(intSlice), cap(intSlice))
	fmt.Println(intSlice)

	var intSlice2 []int32 = []int32{8, 9}
	intSlice = append(intSlice, intSlice2...) // concatenate
	fmt.Println(intSlice)

	// Can also specify capacity, which by default would be the length of the slice:
	var intSlice3 []int32 = make([]int32, 3, 8) // capacity = 8
	fmt.Printf("Length: %v | Capacity: %v\n", len(intSlice3), cap(intSlice3))
	// Good if you want to prevent reallocations

	var myMap map[string]uint8 = make(map[string]uint8)
	fmt.Println(myMap)

	var myMap2 = map[string]uint8{"Christian": 25}
	fmt.Println(myMap2["Christian"]) // 25
	fmt.Println(myMap2["Ida"])       // 0, because it returns the default value for the uint8 type
	var age, ok = myMap2["Ida"]
	if ok {
		fmt.Printf("The age is %v", age)
	} else {
		fmt.Println("Invalid name")
	}

	delete(myMap2, "Christian")

	for name, age := range myMap2 {
		fmt.Printf("Name: %v, Age: %v \n", name, age)
	}

	for i, v := range intArr {
		fmt.Printf("Index: %v, Value: %v \n", i, v)
	}
}

func strings() {
	// golang uses utf8 to represent strings.
	// runes are unicode point numbers that represent the character
}

type gasEngine struct {
	mpg       uint8 // miles per gallon
	gallons   uint8
	ownerInfo owner
}

type owner struct {
	name string
}

func (e gasEngine) milesLeft() uint8 {
	return e.gallons * e.mpg
}

type electricEngine struct {
	mpkwh uint8
	kwh   uint8
}

func (e electricEngine) milesLeft() uint8 {
	return e.kwh * e.mpkwh
}

type engine interface {
	milesLeft() uint8
}

func canMakeIt(e engine, miles uint8) {
	if miles <= e.milesLeft() {
		fmt.Println("You can make it there!")
	} else {
		fmt.Println("Need to fuel up first!")
	}
}

func structs() {
	var myEngine gasEngine = gasEngine{
		mpg:     25,
		gallons: 10,
		ownerInfo: owner{
			name: "Christian",
		},
	}

	fmt.Println(myEngine.gallons, myEngine.mpg, myEngine.ownerInfo.name)
	fmt.Println(myEngine.milesLeft())

	var elecEng electricEngine = electricEngine{
		mpkwh: 1,
		kwh:   1,
	}

	fmt.Println(elecEng)
	canMakeIt(elecEng, 1)
}

func pointers() {
	var p *int32 = new(int32) // initially would be `nil` without `new`, but with `new` we get a memory location
	*p = 10
	fmt.Printf("The value p points to is: %v (addr: %v)\n", *p, p)
	var i int32
	fmt.Printf("The value of i is: %v\n", i)
	// can also make a pointer to i:
	p = &i // address of i, so now p and i reference the same memory address
	*p = 1 // both are now 1

	var thing1 = [5]float64{1, 2, 3, 4, 5}
	fmt.Printf("\nThe value of thing1 is: %v\n", thing1)
	fmt.Printf("\nThe memory location of the thing1 array is: %p", &thing1)
	var result [5]float64 = square(&thing1)
	fmt.Printf("\nThe result is: %v\n", result)
	fmt.Printf("\nThe value of thing1 is: %v\n", thing1)
}

// by using a pointer, you don't need to copy the array, which could lead to performance issues for
// large arrays. But when you modify the array, you modify the 'original' array.
func square(thing2 *[5]float64) [5]float64 {
	fmt.Printf("\nThe memory location of the thing2 array is: %p", thing2)

	for i := range thing2 {
		thing2[i] = thing2[i] * thing2[i]
	}

	return *thing2
}

var m = sync.Mutex{}
var wg = sync.WaitGroup{}
var dbData = []string{"id1", "id2", "id3", "id4", "id5"}
var results = []string{}

func dbCall(i int) {
	// simulating delay
	var delay float32 = rand.Float32() * 2000
	time.Sleep(time.Duration(delay) * time.Millisecond)
	fmt.Println("The result from the database is:", dbData[i])
	m.Lock()
	results = append(results, dbData[i])
	m.Unlock()
	wg.Done()
}

func goroutines() {
	t0 := time.Now()
	for i := 0; i < len(dbData); i++ {
		wg.Add(1)
		go dbCall(i)
	}
	wg.Wait()

	fmt.Printf("\nTotal execution time: %v\n", time.Since(t0))
	fmt.Printf("\nThe results are %v\n", results)
}

func channels() {
	var c = make(chan int, 5) // set a buffer of 5
	// this buffer lets the process channel finish quickly by letting it add up to 5 items to the channel
	// without having to wait for this function to make room in the channel by popping out a value
	go process(c)

	// wait for something to be added to the channel (pop values here)
	for i := range c {
		fmt.Println(i)
		time.Sleep(time.Second)
	}
}

func process(c chan int) {
	defer close(c)

	for i := 0; i < 5; i++ {
		c <- i
	}
}

var MAX_CHICKEN_PRICE float32 = 5
var MAX_TOFU_PRICE float32 = 5

func priceChecker() {
	var chickenChannel = make(chan string)
	var tofuChannel = make(chan string)
	var websites = []string{"walmart.com", "costco.com", "wholefoods.com"}

	for i := range websites {
		go checkChickenPrices(websites[i], chickenChannel)
		go checkTofuPrices(websites[i], tofuChannel)
	}

	sendMessage(chickenChannel, tofuChannel)
}

func checkChickenPrices(website string, chickenChannel chan string) {
	for {
		time.Sleep(time.Second)
		var chikenPrice = rand.Float32() * 20
		if chikenPrice <= MAX_CHICKEN_PRICE {
			chickenChannel <- website
			break
		}
	}
}

func checkTofuPrices(website string, tofuChannel chan string) {
	for {
		time.Sleep(time.Second)
		var tofuPrice = rand.Float32() * 20
		if tofuPrice <= MAX_TOFU_PRICE {
			tofuChannel <- website
			break
		}
	}
}

func sendMessage(chickenChannel chan string, tofuChannel chan string) {
	select {
	case website := <-chickenChannel:
		fmt.Printf("\nText Sent: Found a deal on chicken at %s\n", website)
	case website := <-tofuChannel:
		fmt.Printf("\nEmail Sent: Found a deal on tofu at %s\n", website)
	}
}

func generics() {
	var intSlice = []int{1, 2, 3}
	fmt.Println(sumSlice(intSlice), isEmpty(intSlice))

	var float32Slice = []float32{1, 2, 3}
	fmt.Println(sumSlice(float32Slice), isEmpty(float32Slice))
}

func sumSlice[T int | float32 | float64](slice []T) T {
	var sum T

	for _, v := range slice {
		sum += v
	}

	return sum
}

func isEmpty[T any](slice []T) bool {
	return len(slice) == 0
}

type car[T gasEngine | electricEngine] struct {
	carMake  string
	carModel string
	engine   T
}
