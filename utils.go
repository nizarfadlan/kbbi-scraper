package main

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