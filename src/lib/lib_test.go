package lib

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	emptyLine, _ = regexp.Compile("^\\s*$")
	loop, _ = regexp.Compile("^Loop\\s{1,}\\([0-9]{1,}/[0-9]{1,}\\)$")
	separator, _ = regexp.Compile("^-{1,}")
	nbOfQuestions, _ = regexp.Compile("^Nb of questions:\\s[0-9]{1,}")
	limitReached, _ = regexp.Compile("^Limit reached. Exiting. Number of loops set to:\\s[0-9]{1,}")
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

// TestConcatenate check that adding a set of questions/answers to another
// set is working fine.
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
	count := topic.GetSubsectionsCount()
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

// TestParsingSelectedTopics checks that the option -l (picking specific topics)
func TestParsingSelectedTopics(t *testing.T) {
	selected := "Topic 1,Topic 2"
	arguments := []string{"-l", selected}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Parsing detects list of selected topics as an error")
	}
	if p.subsections != selected {
		t.Errorf("Parsing failed to set the list of selected topics to the string that was passed as parameter.")
	}
	listAsArray := p.GetListOfSubsections()
	if len(listAsArray) != 2 {
		t.Errorf("Retrieving the list of selected topics should have reported 2 elements but we received %d\n", len(listAsArray))
	}
}

// TestNoSelectedTopicsReturnsNil checks that when the user sets no
// specific topics, the array in nil.
func TestNoSelectedTopicsReturnsNil(t *testing.T) {
	arguments := []string{}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Passing no argument make the parsing fail")
	}
	if p.GetListOfSubsections() != nil {
		t.Errorf("No argument passed but the parameters holds a non nil list of selected topics as '%v'", p.GetListOfSubsections())
	}
}

// TestParsingReverseMode checks that reverse mode is detected and works.
func TestParsingReverseMode(t *testing.T) {
	arguments := []string{"-r"}
	p, err := Parse(arguments[:]...)
	if err != nil {
		t.Errorf("Parsing detects reverse mode as an error")
	}
	if p.IsReversedMode() != true {
		t.Errorf("Parsing failed to set the reverse mode.")
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

func getSampleCsvAsStream() string {
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

	return content
}

func getTpp() TopicParsingParameters {
	return TopicParsingParameters {
		TopicAnnounce: "### Lesson ",
		QaSep:         ";",
	}
}

// Testing the way to get the data into the topic data structure.
func TestParseStream(t *testing.T) {

	r := strings.NewReader(getSampleCsvAsStream())
	p := getTpp()

	topic := ParseTopic(r, p)
	count := topic.GetSubsectionsCount()
	if count != 3 {
		t.Errorf("After parsing the stream should result in 3 subtopics. We have counted %d\n", count)
	}

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

func getGenericInterrogationParameters() InterrogationParameters {
	ip := InterrogationParameters{
		interactive: false,
		wait:        1*time.Millisecond,
		limit:       10,
		mode:        linear,
		qachan:	     make(chan string),
		command:	   make(chan string),
		publisher:	 make(chan string),
	}
	return ip
}

func getGenericUnattendedInterrogationParameters() InterrogationParameters {
	ip := getGenericInterrogationParameters()
	ip.interactive = false
	return ip
}

func getGenericInteractiveInterrogationParameters() InterrogationParameters {
	ip := getGenericInterrogationParameters()
	ip.interactive = true
	return ip
}

// TestAskQuestions tests that, in case of linear run of the questions,
// you get the good questions and good answers that respects the requested
// order.
func TestAskQuestionsInUnattendedMode(t *testing.T) {

	r := strings.NewReader(getSampleCsvAsStream())
	tpp := getTpp()
	topic := ParseTopic(r, tpp)

	pr, pw := io.Pipe()
	defer pw.Close()
	ip := getGenericUnattendedInterrogationParameters()
  ip.out = pw

	fmt.Println("    ****************")
	fmt.Println("Test Ask Question in Linear Mode...")
	fmt.Println("    ****************")

	questionsSet := topic.BuildQuestionsSet()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		AskQuestions(questionsSet, ip)
		pw.Close()
	}()

	s := bufio.NewScanner(pr)

	validateOutput(tpp, questionsSet, *s, t, ip.reversed)

}

// TestAskQuestionsInReverseMode tests that, in case of linear and reverse run
// of the questions, you get the good questions and good answers that respects
// the requested order.
func TestAskQuestionsInReverseAndUnattendedMode(t *testing.T) {

	r := strings.NewReader(getSampleCsvAsStream())
	tpp := getTpp()
	topic := ParseTopic(r, tpp)

	pr, pw := io.Pipe()
	ip := getGenericUnattendedInterrogationParameters()
  ip.out = pw
	ip.reversed = true

	questionsSet := topic.BuildQuestionsSet()
	go func() {
		defer pw.Close()
		AskQuestions(questionsSet, ip)
	}()

	fmt.Println("    ****************")
	fmt.Println("    Test Ask Question in Linear Reversed Mode...")
	fmt.Println("    ****************")

	s := bufio.NewScanner(pr)

	validateOutput(tpp, questionsSet, *s, t, ip.reversed)
}

//
func validateOutput(tpp TopicParsingParameters, questionsSet QuestionsAnswers, s bufio.Scanner, t *testing.T, reverseMode bool) {

	announcement, _ := regexp.Compile("^" + tpp.TopicAnnounce)
	questionsCount := questionsSet.GetCount()
	i := 0
	var (
		isAnnounce     bool
		isEmpty        bool
		isLoop         bool
		isSeparator    bool
		isNbOfQ        bool
		isLimitReached bool
		expected       string
		computed       string
	)
	for s.Scan() {
		isAnnounce = announcement.MatchString(s.Text())
		isEmpty = emptyLine.MatchString(s.Text())
		isLoop = loop.MatchString(s.Text())
		isSeparator = separator.MatchString(s.Text())
		isNbOfQ = nbOfQuestions.MatchString(s.Text())
		isLimitReached = limitReached.MatchString(s.Text())
		if !isAnnounce && !isEmpty && !isLoop && !isSeparator && !isNbOfQ && ! isLimitReached {
			// default is non reverse mode
			expected = questionsSet.questions[i] + "     --> " + questionsSet.answers[i]
			if reverseMode {
			  expected = questionsSet.answers[i] + "     --> " + questionsSet.questions[i]
			}
			computed = s.Text()
			if computed != expected {
				t.Errorf("Check of answers failed. We were expected '%s' but received '%s'\n", expected, computed)
			}
			i = (i + 1) % questionsCount
		}
	}
}

// TestAskQuestionsInteractive tests that, in case of linear and interactive
// run of the questions, the user gets the good questions and has to press
// return to get the matching answers and all of this in the requested order.
func TestAskQuestionsInLinearAndInteractiveMode(t *testing.T) {

	r := strings.NewReader(getSampleCsvAsStream())
	tpp := getTpp()
	topic := ParseTopic(r, tpp)

	pr, pw := io.Pipe()
	userIn, userOut := io.Pipe()
	ip := getGenericInteractiveInterrogationParameters()
	ip.in = userIn
	ip.out = pw

	fmt.Println("    ****************")
	fmt.Println("    Test Ask Question in Linear Mode (pseudo-interactive)...")
	fmt.Println("    ****************")

	questionsSet := topic.BuildQuestionsSet()
	questionsCount := questionsSet.GetCount()

	go func() {
		defer pw.Close()
		AskQuestions(questionsSet, ip)
	}()

	go func() {
		// Simulation of interactive mode: the "user" sends return
		// carriage to command.
		for i := 0; i < ip.limit * questionsCount; i++ {
			fmt.Fprintf(userOut, "\n")
		}
	}()

	s := bufio.NewScanner(pr)

	validateOutput(tpp, questionsSet, *s, t, ip.reversed)

}
