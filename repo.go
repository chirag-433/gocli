package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>>>")
		scanner.Scan()
		text := scanner.Text()
		fmt.Println("echoing: ", text)
	}
}
func cleanInput(str string) []string {

	words := strings.Fields(strings.ToLower(str))
	return words

}
