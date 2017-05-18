package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	//"strconv"
	//"bytes"
	"strings"
)

type TestSuite struct {
	XMLName      xml.Name   `xml:"testsuite"`
	ErrorCount   int        `xml:"errors,attr"`
	FailureCount int        `xml:"failures,attr"`
	SuiteName    string     `xml:"name,attr"`
	SkipCount    int        `xml:"skips,attr"`
	TCCount      int        `xml:"tests,attr"`
	TimeDuration string     `xml:"time,attr"`
	TestCases    []TestCase `xml:"testcase"`
}

type TestCase struct {
	XMLName    xml.Name `xml:"testcase"`
	TCClasName string   `xml:"classname,attr"`
	TCFile     string   `xml:"file,attr"`
	TCLine     string   `xml:"line,attr"`
	TCName     string   `xml:"name,attr"`
	Value      string   `xml:",chardata"`
	//TCDuration float32  `xml:"time,attr"`
}

type Failure struct {
	XMLName          xml.Name `xml:"failure"`
	FailureMessage   string   `xml:"message,attr"`
	InnerFailureText string   `xml:",chardata"`
}

func main() {
	inputFileName := ""
	outputFileName := ""

	if len(os.Args) <= 2 {
		fmt.Println("Input or output files not given!")
		return
	}

	inputFileName, outputFileName = os.Args[1], os.Args[2]

	inputFile, inputError := os.Open(inputFileName)
	if inputError != nil {
		fmt.Println("inputerror line 34")
		return
	}
	defer inputFile.Close()

	testCaseSlice := make([]TestCase, 0)
	finalSlice := make([]TestCase, 0)
	suiteStartIndex := 0
	var currentTestCase TestCase

	totalErrorCount := 0
	totalFailureCount := 0
	totalSkipCount := 0
	totalTCCount := 0
	currentSuiteName := ""
	suiteTimeDuration := ""

	//var currentTCClassName string
	currentTCFile := ""
	currentTCLine := ""
	currentTCName := ""

	isInAssertion := false
	inputReader := bufio.NewReader(inputFile)
	isHasFailure := false
	var currentFailureMessage string
	var currentFailureInnerText string
	for {
		inputString, readerError := inputReader.ReadString('\n')
		if readerError == io.EOF {
			break
		}
		// fmt.Printf(inputString)
		if strings.HasPrefix(inputString, "--- PASS: ") || strings.HasPrefix(inputString, "--- FAIL: ") {
			//duration
			suiteTimeDuration = strings.TrimRight(strings.Split(inputString, "(")[1], ")\n")
			currentSuiteName = strings.Split(inputString, " ")[2]
			fmt.Println("Current duration is: ", suiteTimeDuration)
			for ; suiteStartIndex < len(testCaseSlice); suiteStartIndex++ {
				tc := testCaseSlice[suiteStartIndex]
				tc.TCClasName = currentSuiteName
				finalSlice = append(finalSlice, tc)
			}
			continue
		}

		if isInAssertion {
			if strings.Contains(inputString, "],") {
				//fmt.Println("Now out of assertion")
				isInAssertion = false
				continue
			} else {
				if strings.Contains(inputString, "\"Failure\": ") {
					if !strings.Contains(inputString, `"Failure": "",`) {
						totalFailureCount++
						isHasFailure = true
						currentFailureMessage = strings.TrimRight(strings.TrimSpace(strings.TrimLeft(inputString, `"Failure": "`)), "\",")
						fmt.Println("=========Has failure=========", currentFailureMessage)
						continue
					}
				} else if strings.Contains(inputString, `"StackTrace"`) {
					if !strings.Contains(inputString, `"StackTrace": "",`) {
						currentFailureInnerText = strings.TrimRight(strings.TrimSpace(strings.TrimLeft(inputString, `"StackTrace": "`)), `",\n`)
						fmt.Println("==========Has callstack=======", currentFailureInnerText)
						continue
					}
				} else if strings.Contains(inputString, `"Error": `) {
					if !strings.Contains(inputString, `"Error": null,`) {
						totalErrorCount++
						continue
					}
					continue
				} else if strings.Contains(inputString, `"Skipped": true`) {
					totalSkipCount++
					continue
				}
			}
		}
		//if strings.Contains(inputString, "=== RUN") {
		//	currentSuiteName = strings.TrimSpace(strings.TrimLeft(inputString, "=== RUN "))
		//currentTCClassName = currentSuiteName
		//	fmt.Println("The current suite name: ", currentSuiteName)
		//	continue
		//}
		if strings.Contains(inputString, "\"Title\":") {
			currentTCName = strings.TrimRight(strings.TrimLeft(inputString, "\"Title\": \""), "\",\n")
			//fmt.Println("The current TC name 99: ", currentTCName)
		} else if strings.Contains(inputString, "\"File\":") {
			currentTCFile = strings.TrimRight(strings.TrimLeft(inputString, "\"File\": \""), "\",\n")
			//fmt.Println("The current TC File 104: ", currentTCFile)
		} else if strings.Contains(inputString, "\"Line\":") {
			currentTCLine = strings.TrimRight(strings.TrimLeft(inputString, "\"Line\": \""), "\",\n")
			//fmt.Println("The current Line 109: ", currentTCLine)
		} else if strings.Contains(inputString, "\"Depth\":") {
			//fmt.Println("Ignore depth 111")
			continue
		} else if strings.Contains(inputString, "\"Output\":") {
			//fmt.Println("ingore output 115")
			continue
		} else if strings.Contains(inputString, `"Assertions": `) && !strings.Contains(inputString, "[]") {
			isInAssertion = true
			//fmt.Println(inputString)
			//fmt.Println("Now in assertion >>>>>>>>>>")
			continue
		}
		if strings.HasPrefix(inputString, "},") {
			totalTCCount++
			// end of current test case, create a test case object
			//fmt.Println("Has Test Cases in total: ", totalTCCount)
			// use the method name as the class name
			//fmt.Println(currentTCClassName)

			//var currentFailure Failure
			if isHasFailure {
				//fmt.Println("currentFailureMessage is: ", currentFailureMessage)
				//fmt.Println("currentFailureInnerText is : ", currentFailureInnerText)
				/*
					var buff bytes.Buffer
					buff.WriteString("<failure message=\"")
					buff.WriteString(currentFailureMessage)
					buff.WriteString("\">")
					buff.WriteString(currentFailureInnerText)
					buff.WriteString(`</failure>`)
				*/
				content := "<failure message=\""
				content += currentFailureMessage
				content += "\">"
				content += currentFailureInnerText
				content += `</failure>`
				/*
					currentFailure = Failure{
						FailureMessage:   currentFailureMessage,
						InnerFailureText: currentFailureInnerText,
					}*/
				currentTestCase = TestCase{
					TCClasName: currentSuiteName,
					TCFile:     currentTCFile,
					TCLine:     currentTCLine,
					TCName:     currentTCName,
					Value:      content,
				}
				isHasFailure = false
				currentFailureMessage = ""
				currentFailureInnerText = ""
				//fmt.Println("\n\nLine 185: ", currentTestCase.Value)
			} else {
				currentTestCase = TestCase{
					TCClasName: currentSuiteName,
					TCFile:     currentTCFile,
					TCLine:     currentTCLine,
					TCName:     currentTCName,
					Value:      "",
				}
				//fmt.Println("\n\nLine 194: ", currentTestCase.Value)
			}
			isHasFailure = false
			fmt.Println(currentTestCase.TCName)
			testCaseSlice = append(testCaseSlice, currentTestCase)
			//fmt.Println(testCaseSlice)
		}
	}
	// make the testsuite root node

	report := TestSuite{
		ErrorCount:   totalErrorCount,
		FailureCount: totalFailureCount,
		SuiteName:    "PushServerTest",
		SkipCount:    totalSkipCount,
		TCCount:      totalTCCount,
		TimeDuration: suiteTimeDuration,
		TestCases:    finalSlice,
	}

	//fmt.Println(report)

	//xmlBytes, _ := xml.Marshal(report)
	//fmt.Printf("XML Report format is: %s", xmlBytes)
	fmt.Println("\n Write the xml to file\n")

	//if the output file already exist, delete it
	if _, err := os.Stat(outputFileName); err == nil {
		fmt.Println("File already exist. Delete it...")
		os.Remove(outputFileName)
	}
	if _, err := os.Stat("temp.dat"); err == nil {
		fmt.Println("File already exist. Delete it...")
		os.Remove("temp.dat")
	}

	file, _ := os.OpenFile("temp.dat", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	enc := xml.NewEncoder(file)

	erro := enc.Encode(report)
	if erro != nil {
		fmt.Println("Error in the last")
	}

	fmt.Println("Print the file content:")
	bufout, readErr := ioutil.ReadFile("temp.dat")
	if readErr != nil {
		fmt.Println("Read File Error: ", readErr)
	}
	replaceLT := strings.Replace(string(bufout), `&lt;`, "<", -1)
	replaceGT := strings.Replace(string(replaceLT), `&gt;`, ">", -1)
	replaceQ := strings.Replace(string(replaceGT), `&#34;`, "\"", -1)
	fmt.Println(string(replaceQ))
	errz := ioutil.WriteFile(outputFileName, []byte(replaceQ), 0)
	if errz != nil {
		panic(errz)
	}
}

