package utility

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// [仿forEach()](https://openhome.cc/Gossip/Go/Closure.html) => forEach(numbers, func(element int, index int) { sum += element })
func ForEach[T any](elements []T, closure func(T, int)) {
	for index, element := range elements {
		closure(element, index)
	}
}

// [仿map()](https://medium.com/starbugs/why-go-need-generics-c8f1495ef00a) => title := Map(numbers, func(number int, index int) string { return strconv.Itoa(number * 2) })
func Map[T, U any](elements []T, closure func(T, int) U) []U {

	newElements := make([]U, len(elements))

	ForEach(elements, func(element T, index int) {
		newElements[index] = closure(element, index)
	})

	return newElements
}

// [印出結果](https://dotblogs.com.tw/DizzyDizzy/2019/11/29/ColorfulCLI) => log.SetFlags(log.LstdFlags | log.Lshortfile)
func Println(value ...any) {
	fmt.Print("\033[37;103m[DEBUG]\033[0m ")
	log.Println(value...)
}

// 文字 => 數字
func StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

// 讀取出類別
func TypeOf(object any) reflect.Type {
	return reflect.TypeOf(object)
}

// 將網路傳來的[]byte => map
func RawBodyToMap(rawJSON []byte) map[string]interface{} {

	dictionary := map[string]interface{}{}
	json.Unmarshal(rawJSON, &dictionary)

	return dictionary
}
