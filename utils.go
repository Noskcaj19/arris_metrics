package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var WhitespaceSplitRe = regexp.MustCompile(`\s+`)

func parseFreq(freq string) (float64, error) {
	parts := WhitespaceSplitRe.Split(freq, -1)
	if len(parts) == 1 {
		return strconv.ParseFloat(parts[0],32)
	} else if len(parts) == 2 {
		switch parts[1] {
		case "MHz":
			freq, err := strconv.ParseFloat(parts[0],32)
			if err != nil {
				return 0, err
			}
			return freq * 1000 * 1000, nil
		default:
			return 0, errors.New(fmt.Sprintf("Unknown unit: %v", parts[1]))
		}
	} else {
		return 0, errors.New(fmt.Sprintf("Got more than 2 freq parts: %v", parts))
	}
}


type noneError struct {
}

func (e noneError) Error() string {
	return "NA"
}

func parseSNR(snr string) (float64, error) {
	if snr == "NA" {
		return 0, noneError{}
	}

	parts := strings.Split(snr, " ")
	if len(parts) == 1 {
		return strconv.ParseFloat(parts[0], 64)
	} else if len(parts) == 2 {
		switch parts[1] {
		case "dB":
			snr, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, err
			}
			return snr, nil
		default:
			return 0, errors.New(fmt.Sprintf("Unknown unit: %v", parts[1]))
		}
	} else {
		return 0, errors.New(fmt.Sprintf("Got more than 2 SNR parts: %v", parts))
	}
}

func parsePowerLevel(powerLevel string) (float64, error) {
	if powerLevel == "NA" {
		return 0, noneError{}
	}

	parts := WhitespaceSplitRe.Split(powerLevel, -1)
	if len(parts) == 1 {
		return strconv.ParseFloat(parts[0], 64)
	} else if len(parts) == 2 {
		switch parts[1] {
		case "dBmV":
			powerLevel, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, err
			}
			return powerLevel, nil
		default:
			return 0, errors.New(fmt.Sprintf("Unknown unit: %v", parts[1]))
		}
	} else {
		return 0, errors.New(fmt.Sprintf("Got more than 2 power level parts: %v", parts))
	}
}

func parseSymbolRate(powerLevel string) (int, error) {
	parts := WhitespaceSplitRe.Split(powerLevel, -1)
	if len(parts) == 1 {
		return strconv.Atoi(parts[0])
	} else if len(parts) == 2 {
		switch parts[1] {
		case "kSym/s":
			symbolRate, err := strconv.Atoi(parts[0])
			if err != nil {
				return 0, err
			}
			return symbolRate, nil
		default:
			return 0, errors.New(fmt.Sprintf("Unknown unit: %v", parts[1]))
		}
	} else {
		return 0, errors.New(fmt.Sprintf("Got more than 2 symbol rate parts: %v", parts))
	}
}

