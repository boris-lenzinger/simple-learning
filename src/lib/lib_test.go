package lib

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestAddEntry is testing not only the AddEntry function but the GetCount
// too and the initialization of a QA structure.
func TestAddEntry(t *testing.T) {
	qa := NewQA()
	if qa.GetCount() != 0 {
		t.Errorf("A freshly created QA structure should not contain any element. But the GetCount function reports %d\n", qa.GetCount())
	}
	qa.AddEntry("question-1", "answer-1")
	if qa.GetCount() != 1 {
		t.Errorf("Expected 1 entry but received %d\n", qa.GetCount())
	}
	qa.AddEntry("question-2", "answer-2")
	if qa.GetCount() != 2 {
		t.Errorf("Expected 2 entry but received %d\n", qa.GetCount())
	}
}

func TestConcatenate(t *testing.T) {
	qa := NewQA()
	qa.AddEntry("question", "answer")

	otherQa := NewQA()
	otherQa.AddEntry("q1", "a1")
	otherQa.AddEntry("q2", "a2")
	qa.Concatenate(otherQa)

	count := qa.GetCount()
	if count != 3 {
		t.Errorf("Concatenate does not work find. We are expecting 3 but get %d\n", count)
	}
}

// TestNewTopic valides the construction of a topic.
func TestNewTopic(t *testing.T) {
	topic := NewTopic()
	if topic.list == nil {
		t.Errorf("A new topic should not have its list empty.")
	}
	count := topic.GetCount()
	if count != 0 {
		t.Errorf("Was expecting 0 but received a count of %d\n", count)
	}
}

// TestParsing validates the parsing of the command line.
func TestParsingEmptyParameters(t *testing.T) {
	p, err := Parse()
	if err != nil {
		t.Errorf("Parsing should not fail with empty parameters")
	}
	if p.interactive {
		t.Errorf("Default is to be in non interactive. But the parameters says the contrary.")
	}
	if p.wait != 2*time.Second {
		t.Errorf("Default is to wait for 2 seconds. But the current value is %v.\n", p.wait)
	}
}

// TestParsingNonEmptyParameters checks that passing an interactive mode and a different
// time are supported.
func TestParsingNonEmptyParameters(t *testing.T) {
	wt := 1500
	arguments := []string{"-i", "-t", strconv.Itoa(wt)}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("A valid list of parameters must not trigger a parsing error.")
	}
	if !p.interactive {
		t.Errorf("The parameter -i was not detected.")
	}
	if p.wait != time.Duration(wt)*time.Millisecond {
		t.Errorf("Failed to detect wait time as %dms. Found %v instead.\n", wt, p.wait)
	}
}

func TestParsingSelectedTopics(t *testing.T) {
	selected := "Topic 1,Topic 2"
	arguments := []string{"-l", selected}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Parsing detects list of selected topics as an error")
	}
	if p.topics != selected {
		t.Errorf("Parsing failed to set the list of selected topics to the string that was passed as parameter.")
	}
	listAsArray := p.GetListOfTopics()
	if len(listAsArray) != 2 {
		t.Errorf("Retrieving the list of selected topics should have reported 2 elements but we received %d\n", len(listAsArray))
	}
}

// TestParsingSummaryMode checks that the feature about the parameter summary works fine.
func TestParsingSummaryMode(t *testing.T) {
	arguments := []string{"-s"}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Parsing detects summary mode as an error")
	}
	if p.mode != summary {
		t.Errorf("Parsing does not set the mode to summary when the option is set")
	}
	if !p.IsSummaryMode() {
		t.Errorf("Parsing does set mode to summary but the method IsSummaryMode fails to report it.")
	}
}

func TestDetectingLinearMode(t *testing.T) {
	arguments := []string{"-m", "linear"}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Parsing detects linear mode as an error")
	}
	if p.mode != linear {
		t.Errorf("Parsing does not set the mode to linear when the option is set")
	}
}

func TestErrorParsing(t *testing.T) {
	arguments := []string{"-t", "15aaa"}
	_, err := Parse(arguments[:]...)
	if err == nil {
		t.Errorf("We do not detect when a time is not an integer.")
	}
}

// Testing the way to get the data into the topic data structure.
func TestParseStream(t *testing.T) {
	content := `
### Lesson 1
1_Question 1;1_Answer 1

### Lesson 2
2_Question 1;2_Answer 1
2_Question 2;2_Answer 2

### Lesson 3
3_Question 1;3_Answer 1
3_Question 2;3_Answer 2
3_Question 3;3_Answer 3
	`

	r := strings.NewReader(content)
	p := TopicParsingParameters{
		TopicAnnounce: "### Lesson ",
		QaSep:         ";",
	}

	topic := ParseTopic(r, p)
	count := topic.GetCount()
	if count != 3 {
		t.Errorf("After parsing the stream should result in 3 subtopics. We have counted %d\n", count)
	}

	fmt.Println("=============")
	fmt.Printf("Topic %v\n", topic)
	fmt.Println("=============")

	qa := topic.BuildQuestionsSet()
	count = qa.GetCount()
	if count != 6 {
		t.Errorf("We should have a list of 6 questions for the global but we found %d\n", count)
	}

	for i := 1; i <= 3; i++ {
		qa = topic.BuildQuestionsSet(strconv.Itoa(i))
		count = qa.GetCount()
		if count != i {
			t.Errorf("We should have a list of %d questions for the global but we found %d\n", i, count)
		}
	}

}
