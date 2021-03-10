//Palette © Albert Bregonia 2021

package game

import (
	"strings"
	"time"
)

//Hide converts the string to underscores
func Hide(word string) string {
	hidden := ``
	for _, char := range word {
		if char != 32 {
			hidden += `_`
		} else {
			hidden += ` `
		}
	}
	return hidden
}

//Space spaces out the underscores, spaces become '&ensp;' for a better view in HTML
func Space(word string) string {
	spaced := ``
	for _, char := range word {
		if char == 32 {
			spaced += `&ensp;`
		} else {
			spaced += string(char) + ` `
		}
	}
	return spaced
}

//Count counts the number of instances of a character
func Count(word string, char rune) int {
	count := 0
	for _, c := range word {
		if c == char {
			count++
		}
	}
	return count
}

//Now formats time.Now() to log.Println() format
func Now() string {
	return strings.ReplaceAll(time.Now().Local().String()[:strings.Index(time.Now().Local().String(), ".")], "-", "/")
}
