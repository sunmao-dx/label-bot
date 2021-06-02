package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var (
	labelRegex    = regexp.MustCompile(`^//(comp|sig|good|bug|wg|stat|kind|device|env|ci|mindspore|DFX|usability|0|1|2)\s*(.*?)\s*$`)
	labelRegexTitle    = regexp.MustCompile(`^(.*)(Lite)\s*(.*?)\s*$`)
	labelRegexBody    = regexp.MustCompile(`^(.*)(/ops/|/kernel/|/minddata/|/parallel/|/optimizer/|/pynative/|/kernel_compiler/|/runtime/|/runtime/device/)\s*(.*?)\s*$`)
)

func TestRun(t *testing.T){
	orgOrigin := "[MS][Lite]dsadsa"
	var labelsToAdd []string
	labelMatches := labelRegexTitle.FindAllStringSubmatch(orgOrigin, -1)
	fmt.Println(labelMatches)
	if len(labelMatches) != 0 {
		labelsToAdd = getLabelsFromREMatches(labelMatches)
	}
	fmt.Println(labelsToAdd)
}

func getLabelsFromREMatches(matches [][]string) []string {
	var labels []string
	for _, match := range matches {
		label := strings.TrimSpace(strings.Trim(match[2],"/"))
		labels = append(labels, label)
	}
	return labels
}
