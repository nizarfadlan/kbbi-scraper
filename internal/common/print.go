/*
 *  Copyright (c) 2024 Nizar Izzuddin Yatim Fadlan <hello@nizarfadlan.dev>
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */
package common

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	errorPrinter   = color.New(color.FgRed).Add(color.Bold)
	successPrinter = color.New(color.FgGreen).Add(color.Bold)
	infoPrinter    = color.New(color.FgBlue)
	warningPrinter = color.New(color.FgYellow)
	defaultPrinter = color.New(color.FgWhite)
)

func PrintMessage(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	message = strings.TrimSpace(message)
	if strings.HasPrefix(message, "[") {
		parts := strings.SplitN(message, "]", 2)
		if len(parts) == 2 {
			event := strings.TrimSpace(strings.Trim(parts[0], "[]"))
			content := strings.TrimSpace(parts[1])
			printEventMessage(event, content)
			return
		}
	}
	defaultPrinter.Println(message)
}

func printEventMessage(event, content string) {
	var printer *color.Color
	switch strings.ToLower(event) {
	case "error":
		printer = errorPrinter
	case "success":
		printer = successPrinter
	case "info":
		printer = infoPrinter
	case "warning":
		printer = warningPrinter
	default:
		printer = defaultPrinter
	}

	fmt.Printf("[%s] ", event)
	printer.Println(content)
}

func PrintError(format string, a ...interface{}) {
	PrintMessage("[ERROR] "+format, a...)
}

func PrintSuccess(format string, a ...interface{}) {
	PrintMessage("[SUCCESS] "+format, a...)
}

func PrintInfo(format string, a ...interface{}) {
	PrintMessage("[INFO] "+format, a...)
}

func PrintWarning(format string, a ...interface{}) {
	PrintMessage("[WARNING] "+format, a...)
}

func PrintCustom(format string, textColor color.Attribute, isBold bool, a ...interface{}) {
	printer := color.New(textColor)
	if isBold {
		printer = printer.Add(color.Bold)
	}
	printer.Printf(format+"\n", a...)
}

func DisplayMenu() {
	PrintCustom("=== Menu Wordlist ===", color.FgHiMagenta, true)
	PrintCustom("1. Find Wordlist (Alpha)", color.FgHiMagenta, true)
	PrintCustom("2. Fetch Wordlist Contents", color.FgHiMagenta, true)
	PrintCustom("3. Quit", color.FgHiMagenta, true)
	fmt.Print("Choose an option (1-3): ")
}

func GetUserChoice() string {
	var choice string
	fmt.Scanln(&choice)
	return strings.TrimSpace(choice)
}

func GetInput(message string) string {
	fmt.Print(message)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}
