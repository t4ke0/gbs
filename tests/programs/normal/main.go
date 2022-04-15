package main

import "fmt"

/*
TESTING go_build_system (gbs)
*/

func main() {
	data := []string{
		"Tamarillo",
		"Cloudberry",
		"Grape",
		"Huckleberry",
		"Dragonfruit",
		"Melon",
		"Pumpkin",
		"Surinam cherry",
		"Elderberry",
		"Soursop",
		"Orange",
		"Banana",
		"Grapefruit",
		"Quince",
		"Red currant",
		"Olive",
		"Nectarine",
		"Jujube",
	}
	fmt.Println(filterFruits(data, 'O'))
}

func filterFruits(content []string, letter rune) (out []string) {
	const index = 0
	for _, n := range content {
		if rune(n[index]) == letter {
			out = append(out, n)
		}
	}
	return
}
